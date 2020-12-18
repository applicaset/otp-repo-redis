package otp_repo_redis_test

import (
	"context"
	"github.com/alicebob/miniredis"
	"github.com/applicaset/otp-repo-redis"
	"github.com/applicaset/otp-svc"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRepository_Create(t *testing.T) {
	ctx := context.Background()

	mr, err := miniredis.Run()
	require.NoError(t, err)

	rc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	repo := otp_repo_redis.New(rc)

	t.Run("Create New", func(t *testing.T) {
		entity := otp_svc.Entity{
			UUID:        uuid.New().String(),
			PhoneNumber: "+1234567890",
			PinCode:     "1234",
			ExpiresAt:   time.Now().Add(time.Minute * 5),
		}

		err = repo.Create(ctx, entity)
		assert.NoError(t, err)
	})

	t.Run("Create Duplicate", func(t *testing.T) {
		entity := otp_svc.Entity{
			UUID:        uuid.New().String(),
			PhoneNumber: "+1234567890",
			PinCode:     "1234",
			ExpiresAt:   time.Now().Add(time.Minute * 5),
		}

		err = repo.Create(ctx, entity)
		assert.NoError(t, err)

		err = repo.Create(ctx, entity)
		assert.Error(t, err)
	})
}

func TestRepository_Find(t *testing.T) {
	ctx := context.Background()

	mr, err := miniredis.Run()
	require.NoError(t, err)

	rc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	repo := otp_repo_redis.New(rc)

	t.Run("Find Unknown", func(t *testing.T) {
		res, err := repo.Find(ctx, uuid.New().String())
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("Find Created", func(t *testing.T) {
		entity := otp_svc.Entity{
			UUID:        uuid.New().String(),
			PhoneNumber: "+1234567890",
			PinCode:     "1234",
			ExpiresAt:   time.Now().Add(time.Minute * 5),
		}

		err = repo.Create(ctx, entity)
		assert.NoError(t, err)

		res, err := repo.Find(ctx, entity.UUID)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, entity.UUID, res.UUID)
		assert.Equal(t, entity.PhoneNumber, res.PhoneNumber)
		assert.Equal(t, entity.PinCode, res.PinCode)
		assert.True(t, entity.ExpiresAt.Equal(res.ExpiresAt))
	})

	t.Run("Find After Expire", func(t *testing.T) {
		entity := otp_svc.Entity{
			UUID:        uuid.New().String(),
			PhoneNumber: "+1234567890",
			PinCode:     "1234",
			ExpiresAt:   time.Now().Add(time.Second * 2),
		}

		err = repo.Create(ctx, entity)
		assert.NoError(t, err)

		mr.FastForward(time.Second * 2)

		res, err := repo.Find(ctx, entity.UUID)
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}
