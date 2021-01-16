package otpreporedis

import (
	"context"
	"encoding/json"
	"github.com/applicaset/otp-svc"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"time"
)

type repository struct {
	client redis.Cmdable
}

func (repo *repository) Create(ctx context.Context, entity otpsvc.Entity) error {
	res := repo.client.Get(ctx, entity.UUID)

	err := res.Err()
	if err == nil {
		return errors.Errorf("otp with id '%s' already exists", entity.UUID)
	}

	if !errors.Is(err, redis.Nil) {
		return errors.Wrap(err, "error on get key")
	}

	b, err := json.Marshal(entity)
	if err != nil {
		return errors.Wrap(err, "error on marshal entity")
	}

	err = repo.client.SetEX(ctx, entity.UUID, b, time.Until(entity.ExpiresAt)).Err()
	if err != nil {
		return errors.Wrap(err, "error on set key")
	}

	return nil
}

func (repo *repository) Find(ctx context.Context, otpUUID string) (*otpsvc.Entity, error) {
	res := repo.client.Get(ctx, otpUUID)

	err := res.Err()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, errors.Wrap(err, "error on get key")
	}

	var b []byte

	err = res.Scan(&b)
	if err != nil {
		return nil, errors.Wrap(err, "error on scan response")
	}

	var rsp otpsvc.Entity

	err = json.Unmarshal(b, &rsp)
	if err != nil {
		return nil, errors.Wrap(err, "error on unmarshal entity")
	}

	return &rsp, nil
}

func New(client redis.Cmdable) otpsvc.Repository {
	repo := repository{
		client: client,
	}

	return &repo
}
