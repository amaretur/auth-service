package repository

import (
	"time"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"

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
) (string, error) {

	document := TokenDocument{
		Token: token,
		ExpireAt: time.Now().Add(time.Minute * expire),
	}

	res, err := t.collection.InsertOne(ctx, document)
	if err != nil {
		t.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
		}).Errorf("insert token: %s", err)

		return "", errors.Internal.New("repository internal").Wrap(err)
	}

	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (t *TokenRepositoryMongo) GetById(
	ctx context.Context,
	id string,
) (string, error) {

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return "", errors.Internal.New("invalid object id").Wrap(err)
	}

	filter := bson.M{"_id": objectId}

	var data struct {
		Token	string	`bson:"token"`
	}

	if err := t.collection.FindOne(ctx, filter).Decode(&data); err != nil {

		logger := t.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
			"_id": id,
		})

		if err == mongo.ErrNoDocuments {
			logger.Warn(err)

			return "", errors.NotFound.New("token not found").Wrap(err)
		}

		logger.Error(err)

		return "", errors.Internal.NewDefault().Wrap(err)
	}

	return data.Token, nil
}

func (t *TokenRepositoryMongo) Delete(
	ctx context.Context,
	id string,
) error {

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.Internal.New("invalid object id").Wrap(err)
	}

	deleteResult, err := t.collection.DeleteOne(ctx, bson.M{"_id": objectId})
	if err != nil {
		t.logger.WithFields(map[string]any{
			"req_id": reqid.FromContext(ctx),
		}).Error(err)

		return errors.Internal.New("internal repository").Wrap(err)
	}

	t.logger.Infof("deleted count: %s", deleteResult.DeletedCount)

	return nil
}
