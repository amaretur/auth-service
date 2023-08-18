package service

import (
	"time"
	"context"
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v5"

	"github.com/amaretur/auth-service/internal/dto"
	"github.com/amaretur/auth-service/internal/errors"

	"github.com/amaretur/auth-service/pkg/log"
	"github.com/amaretur/auth-service/pkg/reqid"
	errutil "github.com/amaretur/auth-service/pkg/errors"
)

type AccessClaims struct {
	*jwt.RegisteredClaims
	Uuid		string	`json:"uuid"`
	RefreshId	string	`json:"r_id"`
}

type TokenRepository interface {
	Save(
		ctx context.Context,
		token string,
		expire time.Duration,
	) (string, error)

	GetById(ctx context.Context, id string) (string, error)
	Delete(ctx context.Context, id string) error
}

type Jwt struct {

	accessExpire time.Duration
	refreshExpire time.Duration

	method jwt.SigningMethod

	secret []byte

	refreshLen int // длина refresh токена

	repo TokenRepository

	logger log.Logger
}

func NewJwt(
	repo TokenRepository,
	accessExpire, refreshExpire time.Duration,
	secret string,
	logger log.Logger,
) *Jwt {
	return &Jwt{
		repo: repo,

		accessExpire: accessExpire,
		refreshExpire: refreshExpire,

		method: jwt.GetSigningMethod("HS512"),
		secret: []byte(secret),

		refreshLen: 32,

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

	uuid, refreshId, _, err := j.parseAccess(ctx, tokens.Access)
	if err != nil {
		return nil, errors.InvalidToken.New("invalid access token").Wrap(err)
	}

	err = j.validateRefreshToken(ctx, tokens.Refresh, refreshId)
	if err != nil {
		return nil, errors.InvalidToken.New("invalid refresh token").Wrap(err)
	}

	return j.createTokens(ctx, uuid)
}

func (j *Jwt) createTokens(
	ctx context.Context,
	uuid string,
) (*dto.Tokens, error) {

	refresh, refreshId, err := j.createRefresh(ctx)
	if err != nil {
		return nil, err
	}

	access, err := j.createAccess(ctx, uuid, refreshId)
	if err != nil {
		return nil, err
	}

	return &dto.Tokens{
		Access: access,
		Refresh: refresh,
	}, nil
}

func (j *Jwt) createAccess(
	ctx context.Context,
	uuid string,
	refreshId string,
) (string, error) {

	token := jwt.New(j.method)

	token.Claims = &AccessClaims{
		RegisteredClaims: &jwt.RegisteredClaims{
			ExpiresAt: j.expiresAt(j.accessExpire),
		},
		Uuid: uuid,
		RefreshId: refreshId,
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

func (j *Jwt) createRefresh(ctx context.Context) (string, string, error) {

	// Генерируем случайный токен
	token, err := j.generateRandomToken(ctx, j.refreshLen)
	if err != nil {
		return "", "", errors.Internal.New("generate random token").Wrap(err)
	}

	// Сохраняем токен в базу
	refreshId, err := j.saveRefresh(ctx, token)
	if err != nil {
		j.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
		}).Errorf("save refresh: %s", err)

		return "", "", errors.Internal.New("save refresh").Wrap(err)
	}

	return token, refreshId, nil
}

func (j *Jwt) saveRefresh(ctx context.Context, token string) (string, error) {

	hashedToken, err := j.hash(token)
	if err != nil {
		j.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
		}).Errorf("hash refresh: %s", err)

		return "", errors.Internal.New("hash refresh").Wrap(err)
	}

	refreshId, err := j.repo.Save(ctx, hashedToken, j.refreshExpire)
	if err != nil {

		j.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
		}).Errorf("save refresh: %s", err)

		return "", errors.Internal.New("save refresh").Wrap(err)
	}

	return refreshId, nil
}

// Генерирует рандомный токен указанного размера
// (используется для формирования refresh токена)
func (j *Jwt) generateRandomToken(
	ctx context.Context,
	length int,
) (string, error) {

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
) (string, string, bool, error) {

	var isExpired bool = false

	claims, isExpired, err := j.parseClaims(ctx, token, &AccessClaims{})
	if err != nil {

		j.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
		}).Warnf("parse: %s", err.Error())

		return "", "", false, err
	}

	accessClaims, ok := claims.(*AccessClaims)
	if !ok {

		j.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
		}).Warnf("parse claims error")

		return "", "", false, errors.InvalidToken.New("invalid token type")
	}

	return accessClaims.Uuid, accessClaims.RefreshId, isExpired, nil
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

		if errutil.Is(err, jwt.ErrTokenExpired) {
			return parsedToken.Claims, true, nil
		}

		j.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
		}).Warnf("parse: %s", err)

		return nil, false, errors.InvalidToken.New("invalid token").Wrap(err)
	}

	return parsedToken.Claims, false, nil
}

// Функция валидации refresh токена
func (j *Jwt) validateRefreshToken(
	ctx context.Context,
	refresh string,
	refreshId string,
) error {

	// Получаем хеш токена по id
	hashedRefresh, err := j.repo.GetById(ctx, refreshId)
	if err != nil {

		logger := j.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
			"refresh_id": refreshId,
		})

		if errutil.Has(err, errors.NotFound) {
			logger.Warn("token not found")

			return errors.InvalidToken.NewDefault().Wrap(err)
		}

		logger.Error("get token error")

		return errors.Internal.NewDefault().Wrap(err)
	}

	// Сравниваем хеш с токеном
	if err := j.hashCompare(hashedRefresh, refresh); err != nil {
		return errors.InvalidToken.NewDefault().Wrap(err)
	}

	// Удаляем токен из БД
	return j.repo.Delete(ctx, refreshId)
}

func (j *Jwt) hash(data string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(data), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (j *Jwt) hashCompare(hashedData, data string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedData), []byte(data))
}

func (j *Jwt) expiresAt(duration time.Duration) *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(time.Minute * duration))
}
