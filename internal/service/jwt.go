package service

import (
	"context"

	"github.com/amaretur/auth-service/internal/dto"

	"github.com/amaretur/auth-service/pkg/log"
)

type Jwt struct {
	logger log.Logger
}

func NewJwt(logger log.Logger) *Jwt {
	return &Jwt{
		logger: logger,
	}
}

func (j *Jwt) CreateTokens(
	ctx context.Context,
	uuid string,
) (*dto.Tokens, error) {

	j.logger.Info(uuid)

	return nil, nil
}

func (j *Jwt) RefreshTokens(
	ctx context.Context,
	tokens *dto.Tokens,
) (*dto.Tokens, error) {

	j.logger.Info(tokens)

	return nil, nil
}

