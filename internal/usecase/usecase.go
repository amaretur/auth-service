package usecase

import (
	"context"

	"github.com/amaretur/auth-service/internal/dto"

	"github.com/amaretur/auth-service/pkg/log"
)

type Usecase struct {
	logger log.Logger
}

func New(logger log.Logger) *Usecase {
	return &Usecase{
		logger: logger,
	}
}

func (u *Usecase) SignIn(
	ctx context.Context,
	uuid string,
) (*dto.Tokens, error) {

	u.logger.Info(uuid)

	return nil, nil
}

func (u *Usecase) Refresh(
	ctx context.Context,
	tokens *dto.Tokens,
) (*dto.Tokens, error) {

	u.logger.Info(tokens)

	return nil, nil
}
