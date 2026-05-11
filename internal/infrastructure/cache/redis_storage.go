package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	appAuth "kafka-order-demo/backend/internal/application/auth"
	"kafka-order-demo/backend/internal/infrastructure/config"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client *redis.Client
}

const refreshTokenKeyPrefix = "auth:refresh:"
const loginIPKeyPrefix = "auth:login:ip:"
const loginUserKeyPrefix = "auth:login:user:"
const sessionKeyPrefix = "auth:session:user:"

func NewRedisStorage(ctx context.Context, cfg config.RedisConfig) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			return nil, closeErr
		}
		return nil, err
	}

	return &RedisStorage{client: client}, nil
}

func (s *RedisStorage) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return s.client.Set(ctx, key, value, expiration).Err()
}

func (s *RedisStorage) Get(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

func (s *RedisStorage) Delete(ctx context.Context, keys ...string) error {
	return s.client.Del(ctx, keys...).Err()
}

func (s *RedisStorage) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	count, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if count == 1 {
		if err := s.client.Expire(ctx, key, expiration).Err(); err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (s *RedisStorage) StoreRefreshToken(ctx context.Context, token string, userID uint, expiration time.Duration) error {
	return s.Set(ctx, refreshTokenKey(token), strconv.FormatUint(uint64(userID), 10), expiration)
}

func (s *RedisStorage) GetRefreshTokenUserID(ctx context.Context, token string) (uint, error) {
	value, err := s.Get(ctx, refreshTokenKey(token))
	if err != nil {
		return 0, err
	}

	userID, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(userID), nil
}

func (s *RedisStorage) DeleteRefreshToken(ctx context.Context, token string) error {
	return s.Delete(ctx, refreshTokenKey(token))
}

func (s *RedisStorage) GetLoginFailureCount(ctx context.Context, username, ip string) (int64, error) {
	ipCount, err := s.getCount(ctx, loginIPKey(ip))
	if err != nil {
		return 0, err
	}

	userCount, err := s.getCount(ctx, loginUserKey(username))
	if err != nil {
		return 0, err
	}

	if ipCount > userCount {
		return ipCount, nil
	}
	return userCount, nil
}

func (s *RedisStorage) IncrementLoginFailure(ctx context.Context, username, ip string, expiration time.Duration) (int64, error) {
	ipCount, err := s.Increment(ctx, loginIPKey(ip), expiration)
	if err != nil {
		return 0, err
	}

	userCount, err := s.Increment(ctx, loginUserKey(username), expiration)
	if err != nil {
		return 0, err
	}

	if ipCount > userCount {
		return ipCount, nil
	}
	return userCount, nil
}

func (s *RedisStorage) ClearLoginFailures(ctx context.Context, username, ip string) error {
	return s.Delete(ctx, loginIPKey(ip), loginUserKey(username))
}

func (s *RedisStorage) StoreSession(ctx context.Context, session appAuth.Session, expiration time.Duration) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return s.Set(ctx, sessionKey(session.UserID), data, expiration)
}

func (s *RedisStorage) GetSession(ctx context.Context, userID uint) (*appAuth.Session, error) {
	value, err := s.Get(ctx, sessionKey(userID))
	if err != nil {
		return nil, err
	}

	var session appAuth.Session
	if err := json.Unmarshal([]byte(value), &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *RedisStorage) DeleteSession(ctx context.Context, userID uint) error {
	return s.Delete(ctx, sessionKey(userID))
}

func (s *RedisStorage) Close() error {
	return s.client.Close()
}

func refreshTokenKey(token string) string {
	return refreshTokenKeyPrefix + token
}

func loginIPKey(ip string) string {
	return loginIPKeyPrefix + ip
}

func loginUserKey(username string) string {
	return loginUserKeyPrefix + username
}

func sessionKey(userID uint) string {
	return sessionKeyPrefix + strconv.FormatUint(uint64(userID), 10)
}

func (s *RedisStorage) getCount(ctx context.Context, key string) (int64, error) {
	value, err := s.Get(ctx, key)
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	count, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid redis counter %s: %w", key, err)
	}

	return count, nil
}
