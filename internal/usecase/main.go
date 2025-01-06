package usecase

import (
	"github.com/golanguzb70/udevslabs-twitter/config"
	"github.com/golanguzb70/udevslabs-twitter/internal/usecase/repo"
	"github.com/golanguzb70/udevslabs-twitter/pkg/logger"
	"github.com/golanguzb70/udevslabs-twitter/pkg/postgres"
)

// UseCase -.
type UseCase struct {
	UserRepo    UserRepoI
	SessionRepo SessionRepoI
}

// New -.
func New(pg *postgres.Postgres, config *config.Config, logger *logger.Logger) *UseCase {
	return &UseCase{
		UserRepo:    repo.NewUserRepo(pg, config, logger),
		SessionRepo: repo.NewSessionRepo(pg, config, logger),
	}
}
