package repository

import (
	"time"
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/amaretur/auth-service/internal/errors"

	"github.com/amaretur/auth-service/pkg/log"
	"github.com/amaretur/auth-service/pkg/reqid"
)

type TokenDocument struct {
	Token		string		`bson:"token"`
	ExpireAt	time.Time	`bson:"expire_at"`
}

type TokenRepositoryMongo struct {

	database	*mongo.Database
	collection	*mongo.Collection

	logger		log.Logger
}

func NewTokenRepositoryMongo(
	database *mongo.Database,
	logger log.Logger,
) *TokenRepositoryMongo {
	return &TokenRepositoryMongo{
		database: database,
		collection: database.Collection("token"),
		logger: logger,
	}
}

func (t *TokenRepositoryMongo) Save(
	ctx context.Context,
	token string,
	expire time.Duration,
) error {

	document := TokenDocument{
		Token: token,
		ExpireAt: time.Now().Add(time.Minute * expire),
	}

	_, err := t.collection.InsertOne(ctx, document)
	if err != nil {
		t.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
		}).Errorf("insert token: %s", err)

		return errors.Internal.New("repository internal").Wrap(err)
	}

	return nil
}

func (t *TokenRepositoryMongo) Delete(
	ctx context.Context,
	token string,
) (int, error) {

	filter := bson.D{{"token", token}}

	deleteResult, err := t.collection.DeleteOne(ctx, filter)
	if err != nil {
		t.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
		}).Error(err)

		return 0, errors.Internal.New("internal repository").Wrap(err)
	}

	return int(deleteResult.DeletedCount), nil
}
