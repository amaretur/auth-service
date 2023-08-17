package service

import (
	"time"
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/golang-jwt/jwt/v5"

	"github.com/amaretur/auth-service/internal/dto"
	"github.com/amaretur/auth-service/internal/errors"

	"github.com/amaretur/auth-service/pkg/log"
	"github.com/amaretur/auth-service/pkg/reqid"
)

type AccessClaims struct {
	*jwt.RegisteredClaims
	Uuid string
}

type TokenRepository interface {
	Save(ctx context.Context, token string, expire time.Duration) error
	IsExist(ctx context.Context, token string) (bool, error)
}

type Jwt struct {

	accessExpire time.Duration
	refreshExpire time.Duration

	method jwt.SigningMethod

	secret []byte

	refreshLen int // длина refresh токена без учета метки access токена
	accessLabelInRefreshTokenLen int // длина метки access токена

	repo TokenRepository

	logger log.Logger
}

func NewJwt(
	accessExpire, refreshExpire time.Duration,
	secret string,
	logger log.Logger,
) *Jwt {
	return &Jwt{
		accessExpire: accessExpire,
		refreshExpire: refreshExpire,

		method: jwt.GetSigningMethod("SHA512"),
		secret: []byte(secret),

		refreshLen: 10,
		accessLabelInRefreshTokenLen: 6,

		logger: logger.WithFields(map[string]any{
			"unit": "jwt",
		}),
	}
}

func (j *Jwt) CreateTokens(
	ctx context.Context,
	uuid string,
) (*dto.Tokens, error) {
	return j.createTokens(ctx, uuid)
}

func (j *Jwt) RefreshTokens(
	ctx context.Context,
	tokens *dto.Tokens,
) (*dto.Tokens, error) {

	uuid, _, err := j.parseAccess(ctx, tokens.Access)
	if err != nil {
		return nil, errors.InvalidToken.New("invalid access token").Wrap(err)
	}

	err = j.validateRefreshToken(ctx, tokens.Access, tokens.Refresh)
	if err != nil {
		return nil, errors.InvalidToken.New("invalid refresh token").Wrap(err)
	}

	return j.createTokens(ctx, uuid)
}

func (j *Jwt) createTokens(
	ctx context.Context,
	uuid string,
) (*dto.Tokens, error) {

	access, err := j.createAccess(ctx, uuid)
	if err != nil {
		return nil, err
	}

	refresh, err := j.createRefresh(ctx, access)
	if err != nil {
		return nil, err
	}

	return &dto.Tokens{
		Access: access,
		Refresh: refresh,
	}, nil
}

func (j *Jwt) createAccess(ctx context.Context, uuid string) (string, error) {

	token := jwt.New(j.method)

	token.Claims = &AccessClaims{
		RegisteredClaims: &jwt.RegisteredClaims{
			ExpiresAt: j.expiresAt(j.accessExpire),
		},
		Uuid: uuid,
	}

	result, err := token.SignedString(j.secret)
	if err != nil {

		j.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
		}).Errorf("signed string: %s", err)

		return "", errors.Internal.New("signed string").Wrap(err)
	}

	return result, nil
}

// Функция создания refresh токена
// refresh токен имеет вид:
// "random base64 str" + "n-е количество символов взятое с конца access токена"
func (j *Jwt) createRefresh(
	ctx context.Context,
	access string,
) (string, error) {

	// Генерируем случайный токен
	base, err := j.generateRandomToken(ctx, j.refreshLen)
	if err != nil {
		return "", errors.Internal.New("generate random token").Wrap(err)
	}

	// Добавляем к нему фрагмент access токена для связки
	token := base + access[len(access) - j.accessLabelInRefreshTokenLen:]

	// Сохраняем токен в базу
	if err := j.saveRefresh(ctx, token); err != nil {
		return "", err
	}

	return token, nil
}

func (j *Jwt) saveRefresh(ctx context.Context, token string) error {

	if err := j.repo.Save(ctx, token, j.refreshExpire); err != nil {

		j.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
		}).Errorf("save refresh: %s", err)

		return errors.Internal.New("save refresh").Wrap(err)
	}

	return nil
}

// Генерирует рандомный токен указанного размера
// (используется для формирования refresh токена)
func (j *Jwt) generateRandomToken(ctx context.Context, length int) (string, error) {

	b := make([]byte, length)

	_, err := rand.Read(b)
	if err != nil {

		j.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
		}).Errorf("read from rand: %s", err)

		return "", errors.Internal.New("read from rand").Wrap(err)
	}

	return base64.URLEncoding.EncodeToString(b)[:length], nil
}

func (j *Jwt) parseAccess(
	ctx context.Context,
	token string,
) (string, bool, error) {

	claims, isValid, err := j.parseClaims(ctx, token, &AccessClaims{})
	if err != nil {

		j.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
		}).Warnf("parse: %s", err.Error())

		return "", false, err
	}

	accessClaims, ok := claims.(*AccessClaims)
	if !ok {

		j.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
		}).Warnf("parse claims error")

		return "", false, errors.InvalidToken.New("invalid token type")
	}

	return accessClaims.Uuid, isValid, nil
}

func (j *Jwt) parseClaims(
	ctx context.Context,
	token string,
	claims jwt.Claims,
) (jwt.Claims, bool, error) {

	parsedToken, err := jwt.ParseWithClaims(
		token,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return j.secret, nil
		},
	)

	if err != nil {

		j.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
		}).Warnf("parse: %s", err)

		return nil, false, errors.InvalidToken.New("invalid token").Wrap(err)
	}

	return parsedToken.Claims, parsedToken.Valid, nil
}

// Функция валидации refresh токена
// (подразумевается, что подлинность access токена уже подтверждена)
func (j *Jwt) validateRefreshToken(
	ctx context.Context,
	access string,
	refresh string,
) error {

	accessLable := access[len(access) - j.accessLabelInRefreshTokenLen:]
	refreshLable := refresh[len(refresh) - j.accessLabelInRefreshTokenLen:]

	if accessLable != refreshLable {
		return errors.InvalidToken.New("refresh token is invalid")
	}

	isExist, err := j.repo.IsExist(ctx, refresh)
	if err != nil {

		j.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
		}).Error(err)

		return errors.Internal.NewDefault().Wrap(err)
	}

	if !isExist {
		return errors.InvalidToken.New("refresh token is invalid")
	}

	return nil
}

func (j *Jwt) expiresAt(duration time.Duration) *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(time.Minute * duration))
}
