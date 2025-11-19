package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/adedejiosvaldo/safetrace/backend/internal/models"
)

type RedisDB struct {
	client *redis.Client
}

func NewRedisDB(redisURL string) (*RedisDB, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &RedisDB{client: client}, nil
}

func (r *RedisDB) Close() error {
	return r.client.Close()
}

// User state operations
func (r *RedisDB) SetUserState(ctx context.Context, state *models.UserState) error {
	key := fmt.Sprintf("user:state:%s", state.UserID)
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, 24*time.Hour).Err()
}

func (r *RedisDB) GetUserState(ctx context.Context, userID uuid.UUID) (*models.UserState, error) {
	key := fmt.Sprintf("user:state:%s", userID)
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var state models.UserState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// Rate limiting
func (r *RedisDB) CheckRateLimit(ctx context.Context, userID uuid.UUID, window time.Duration, limit int) (bool, error) {
	key := fmt.Sprintf("ratelimit:%s", userID)
	
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// Set expiry on first request
	if count == 1 {
		r.client.Expire(ctx, key, window)
	}

	return count <= int64(limit), nil
}

// Alert deduplication
func (r *RedisDB) CheckAlertSent(ctx context.Context, userID uuid.UUID, window time.Duration) (bool, error) {
	key := fmt.Sprintf("alert:sent:%s", userID)
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (r *RedisDB) MarkAlertSent(ctx context.Context, userID uuid.UUID, window time.Duration) error {
	key := fmt.Sprintf("alert:sent:%s", userID)
	return r.client.Set(ctx, key, "1", window).Err()
}

// Caching
func (r *RedisDB) CacheUser(ctx context.Context, user *models.User, ttl time.Duration) error {
	key := fmt.Sprintf("user:cache:%s", user.ID)
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisDB) GetCachedUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	key := fmt.Sprintf("user:cache:%s", userID)
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}
	return &user, nil
}
