package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

type Service struct {
	secret []byte
	ttl    time.Duration
	redis  *redis.Client
}

func NewService(ctx context.Context) (*Service, error) {
	ttlMinutes := 60
	if raw := strings.TrimSpace(os.Getenv("JWT_TTL_MINUTES")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return nil, fmt.Errorf("invalid JWT_TTL_MINUTES")
		}
		ttlMinutes = value
	}
	redisAddr := env("REDIS_ADDR", "localhost:6379")
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Username: os.Getenv("REDIS_USER"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection: %w", err)
	}
	return &Service{
		secret: []byte(env("JWT_SECRET", "orbitlab-secret-change-me")),
		ttl:    time.Duration(ttlMinutes) * time.Minute,
		redis:  client,
	}, nil
}

func (s *Service) Close() error {
	return s.redis.Close()
}

func (s *Service) Issue(userID uint) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(s.ttl)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "orbitlab",
			Subject:   strconv.FormatUint(uint64(userID), 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        fmt.Sprintf("%d-%d", userID, now.UnixNano()),
		},
	})
	value, err := token.SignedString(s.secret)
	return value, expiresAt, err
}

func (s *Service) Parse(ctx context.Context, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid || claims.UserID == 0 || claims.ID == "" {
		return nil, errors.New("invalid token")
	}
	blocked, err := s.redis.Exists(ctx, blacklistKey(claims.ID)).Result()
	if err != nil {
		return nil, err
	}
	if blocked > 0 {
		return nil, errors.New("token is blacklisted")
	}
	return claims, nil
}

func (s *Service) Blacklist(ctx context.Context, tokenString string) error {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid || claims.ExpiresAt == nil || claims.ID == "" {
		return errors.New("invalid token")
	}
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}
	return s.redis.Set(ctx, blacklistKey(claims.ID), strconv.FormatUint(uint64(claims.UserID), 10), ttl).Err()
}

func (s *Service) Redis() *redis.Client { return s.redis }

func blacklistKey(jti string) string { return "orbitlab.jwt." + jti }

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
