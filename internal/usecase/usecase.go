package usecase

import (
	"context"

	"github.com/amaretur/auth-service/internal/dto"

	"github.com/amaretur/auth-service/pkg/log"
)

type JwtService interface {
	CreateTokens(ctx context.Context, uuid string) (*dto.Tokens, error)
	RefreshTokens(ctx context.Context, tokens *dto.Tokens) (*dto.Tokens, error)
}

type Usecase struct {
	jwt JwtService

	logger log.Logger
}

func New(jwt JwtService, logger log.Logger) *Usecase {
	return &Usecase{
		jwt: jwt,
		logger: logger,
	}
}

func (u *Usecase) SignIn(
	ctx context.Context,
	uuid string,
) (*dto.Tokens, error) {

	return u.jwt.CreateTokens(ctx, uuid)
}

func (u *Usecase) Refresh(
	ctx context.Context,
	tokens *dto.Tokens,
) (*dto.Tokens, error) {

	return u.jwt.RefreshTokens(ctx, tokens)
}
