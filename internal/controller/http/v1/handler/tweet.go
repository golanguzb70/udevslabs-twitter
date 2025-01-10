package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golanguzb70/udevslabs-twitter/config"
	"github.com/golanguzb70/udevslabs-twitter/internal/entity"
)

// CreateTweet godoc
// @Router /tweet [post]
// @Summary Create a new tweet
// @Description Create a new tweet
// @Security BearerAuth
// @Tags tweet
// @Accept  json
// @Produce  json
// @Param tweet body entity.Tweet true "Tweet object"
// @Success 201 {object} entity.Tweet
// @Failure 400 {object} entity.ErrorResponse
func (h *Handler) CreateTweet(ctx *gin.Context) {
	var (
		body entity.Tweet
	)

	err := ctx.ShouldBindJSON(&body)
	if err != nil {
		h.ReturnError(ctx, config.ErrorBadRequest, "Invalid request body", 400)
		return
	}

	body.Owner.ID = ctx.GetHeader("sub")

	tweet, err := h.UseCase.TweetRepo.Create(ctx, body)
	if h.HandleDbError(ctx, err, "Error creating tweet") {
		return
	}

	tweet.Attachments, err = h.UseCase.TweetAttachmentsRepo.MultipleUpsert(ctx, entity.AttachmentMultipleInsertRequest{
		TweetId:     tweet.Id,
		Attachments: body.Attachments,
	})
	if h.HandleDbError(ctx, err, "Error creating tweet") {
		return
	}

	// create tags for the tweet.

	ctx.JSON(201, tweet)
}

// GetTweet godoc
// @Router /tweet/{id} [get]
// @Summary Get a tweet by ID
// @Description Get a tweet by ID
// @Security BearerAuth
// @Tags tweet
// @Accept  json
// @Produce  json
// @Param id path string true "Tweet ID"
// @Success 200 {object} entity.Tweet
// @Failure 400 {object} entity.ErrorResponse
func (h *Handler) GetTweet(ctx *gin.Context) {
	var (
		req entity.Id
	)

	req.ID = ctx.Param("id")

	tweet, err := h.UseCase.TweetRepo.GetSingle(ctx, req)
	if h.HandleDbError(ctx, err, "Error getting tweet") {
		return
	}

	tweetAttachments, err := h.UseCase.TweetAttachmentsRepo.GetList(ctx,
		entity.GetListFilter{
			Filters: []entity.Filter{
				{
					Column: "tweet_id",
					Type:   "eq",
					Value:  req.ID,
				},
			},
			Page:  1,
			Limit: 10,
		},
	)
	if h.HandleDbError(ctx, err, "Error getting tweet attachments") {
		return
	}

	tweet.Attachments = tweetAttachments.Items

	tweet.Owner, err = h.UseCase.UserRepo.GetSingle(ctx, entity.UserSingleRequest{ID: tweet.Owner.ID})
	if h.HandleDbError(ctx, err, "Error getting tweet owner") {
		return
	}

	ctx.JSON(200, tweet)
}

// GetTweets godoc
// @Router /tweet/list [get]
// @Summary Get a list of tweets
// @Description Get a list of tweets
// @Security BearerAuth
// @Tags tweet
// @Accept  json
// @Produce  json
// @Param page query number true "page"
// @Param limit query number true "limit"
// @Param search query string false "search"
// @Success 200 {object} entity.TweetList
// @Failure 400 {object} entity.ErrorResponse
func (h *Handler) GetTweets(ctx *gin.Context) {
	var (
		req entity.GetListFilter
	)

	page := ctx.DefaultQuery("page", "1")
	limit := ctx.DefaultQuery("limit", "10")
	search := ctx.DefaultQuery("search", "")

	req.Page, _ = strconv.Atoi(page)
	req.Limit, _ = strconv.Atoi(limit)
	req.Filters = append(req.Filters,
		entity.Filter{
			Column: "content",
			Type:   "search",
			Value:  search,
		},
	)

	req.OrderBy = append(req.OrderBy, entity.OrderBy{
		Column: "created_at",
		Order:  "desc",
	})

	tweets, err := h.UseCase.TweetRepo.GetList(ctx, req)
	if h.HandleDbError(ctx, err, "Error getting tweets") {
		return
	}

	ctx.JSON(200, tweets)
}

// UpdateTweet godoc
// @Router /tweet [put]
// @Summary Update a tweet
// @Description Update a tweet
// @Security BearerAuth
// @Tags tweet
// @Accept  json
// @Produce  json
// @Param tweet body entity.Tweet true "Tweet object"
// @Success 200 {object} entity.Tweet
// @Failure 400 {object} entity.ErrorResponse
func (h *Handler) UpdateTweet(ctx *gin.Context) {
	var (
		body entity.Tweet
	)

	err := ctx.ShouldBindJSON(&body)
	if err != nil {
		h.ReturnError(ctx, config.ErrorBadRequest, "Invalid request body", 400)
		return
	}

	if body.Owner.ID != ctx.GetHeader("sub") {
		h.ReturnError(ctx, config.ErrorForbidden, "You have no access to the tweet", http.StatusForbidden)
		return
	}

	tweet, err := h.UseCase.TweetRepo.Update(ctx, body)
	if h.HandleDbError(ctx, err, "Error updating tweet") {
		return
	}

	tweet.Attachments, err = h.UseCase.TweetAttachmentsRepo.MultipleUpsert(ctx, entity.AttachmentMultipleInsertRequest{
		TweetId:     tweet.Id,
		Attachments: body.Attachments,
	})
	if h.HandleDbError(ctx, err, "Error upserting tweet attachments") {
		return
	}

	ctx.JSON(200, tweet)
}

// DeleteTweet godoc
// @Router /tweet/{id} [delete]
// @Summary Delete a tweet
// @Description Delete a tweet
// @Security BearerAuth
// @Tags tweet
// @Accept  json
// @Produce  json
// @Param id path string true "Tweet ID"
// @Success 200 {object} entity.SuccessResponse
// @Failure 400 {object} entity.ErrorResponse
func (h *Handler) DeleteTweet(ctx *gin.Context) {
	var (
		req entity.Id
	)

	req.ID = ctx.Param("id")

	tweet, err := h.UseCase.TweetRepo.GetSingle(ctx, req)
	if h.HandleDbError(ctx, err, "Error getting tweet") {
		return
	}

	if tweet.Owner.ID != ctx.GetHeader("sub") {
		h.ReturnError(ctx, config.ErrorForbidden, "You have no access to the tweet", http.StatusForbidden)
		return
	}

	err = h.UseCase.TweetRepo.Delete(ctx, req)
	if h.HandleDbError(ctx, err, "Error deleting tweet") {
		return
	}

	ctx.JSON(200, entity.SuccessResponse{
		Message: "Tweet deleted successfully",
	})
}
