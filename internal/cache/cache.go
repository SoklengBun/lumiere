package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// ErrMiss indicates that a key is not present in the cache. Cache misses are
// expected and should fall back to the primary data store.
var ErrMiss = errors.New("cache miss")

type Cache interface {
	Get(ctx context.Context, key string, destination any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Close() error
}

// Counter is implemented by cache backends that can atomically increment a
// numeric key. It is used for invalidating versioned collection caches.
type Counter interface {
	Increment(ctx context.Context, key string) (int64, error)
}

type Redis struct {
	client *redis.Client
}

func NewRedis(ctx context.Context, rawURL string) (*Redis, error) {
	options, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &Redis{client: client}, nil
}

func (r *Redis) Get(ctx context.Context, key string, destination any) error {
	value, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return ErrMiss
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(value, destination)
}

func (r *Redis) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, encoded, ttl).Err()
}

func (r *Redis) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *Redis) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *Redis) Close() error {
	return r.client.Close()
}

// UpstashREST is an HTTP-based cache client for serverless environments such
// as Vercel. It avoids keeping a TCP connection pool between invocations.
type UpstashREST struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewUpstashREST(ctx context.Context, rawURL, token string) (*UpstashREST, error) {
	if strings.TrimSpace(rawURL) == "" || strings.TrimSpace(token) == "" {
		return nil, errors.New("upstash REST URL and token are required")
	}

	client := &UpstashREST{
		baseURL: strings.TrimRight(rawURL, "/"),
		token:   token,
		client:  &http.Client{Timeout: 3 * time.Second},
	}
	if _, err := client.command(ctx, "PING"); err != nil {
		return nil, err
	}
	return client, nil
}

func (r *UpstashREST) command(ctx context.Context, args ...any) (json.RawMessage, error) {
	body, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.baseURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+r.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("upstash REST request failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var response struct {
		Result json.RawMessage `json:"result"`
		Error  string          `json:"error"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, err
	}
	if response.Error != "" {
		return nil, errors.New(response.Error)
	}
	return response.Result, nil
}

func (r *UpstashREST) Get(ctx context.Context, key string, destination any) error {
	result, err := r.command(ctx, "GET", key)
	if err != nil {
		return err
	}
	if string(result) == "null" || len(result) == 0 {
		return ErrMiss
	}

	var encoded string
	if err := json.Unmarshal(result, &encoded); err != nil {
		return err
	}
	return json.Unmarshal([]byte(encoded), destination)
}

func (r *UpstashREST) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if ttl <= 0 {
		_, err = r.command(ctx, "SET", key, string(encoded))
		return err
	}
	_, err = r.command(ctx, "SET", key, string(encoded), "EX", int(ttl/time.Second))
	return err
}

func (r *UpstashREST) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	args := make([]any, 0, len(keys)+1)
	args = append(args, "DEL")
	for _, key := range keys {
		args = append(args, key)
	}
	_, err := r.command(ctx, args...)
	return err
}

func (r *UpstashREST) Increment(ctx context.Context, key string) (int64, error) {
	result, err := r.command(ctx, "INCR", key)
	if err != nil {
		return 0, err
	}

	var value int64
	if err := json.Unmarshal(result, &value); err != nil {
		return 0, err
	}
	return value, nil
}

func (r *UpstashREST) Close() error { return nil }
