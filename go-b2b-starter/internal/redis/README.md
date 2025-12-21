# Redis Module Guide

Simple guide for using Redis cache in your modules.

## Setup

Add to your `.env`:

```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=          # Optional
REDIS_DB=0               # Default database
```

For local development, start Redis with Docker:
```bash
make run-deps  # Starts Redis and other dependencies
```

## Usage in Your Module

### 1. Inject the Redis Client

```go
import (
    "github.com/moasq/go-b2b-starter/pkg/redis"
)

type UserService struct {
    cache redis.Client
}

func NewUserService(cache redis.Client) *UserService {
    return &UserService{cache: cache}
}
```

### 2. Basic Operations

**Set a value with TTL:**
```go
func (s *UserService) CacheUser(ctx context.Context, userID string, data string) error {
    key := fmt.Sprintf("user:%s", userID)
    ttl := 1 * time.Hour

    return s.cache.Set(ctx, key, data, ttl)
}
```

**Get a value:**
```go
func (s *UserService) GetCachedUser(ctx context.Context, userID string) (string, error) {
    key := fmt.Sprintf("user:%s", userID)

    value, err := s.cache.Get(ctx, key)
    if err != nil {
        return "", err // redis.Nil if key doesn't exist
    }

    return value, nil
}
```

**Check if key exists:**
```go
exists, err := s.cache.Exists(ctx, "user:123")
if exists {
    // Key is in cache
}
```

**Delete a value:**
```go
err := s.cache.Delete(ctx, "user:123")
```

### 3. Real-World Example: Cache-Aside Pattern

```go
func (s *UserService) GetUser(ctx context.Context, userID int32) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", userID)

    // 1. Try to get from cache first
    cached, err := s.cache.Get(ctx, cacheKey)
    if err == nil {
        // Cache hit - deserialize and return
        var user User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }

    // 2. Cache miss - get from database
    user, err := s.repo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, err
    }

    // 3. Store in cache for next time
    userJSON, _ := json.Marshal(user)
    s.cache.Set(ctx, cacheKey, string(userJSON), 5*time.Minute)

    return user, nil
}
```

### 4. Invalidation on Update

```go
func (s *UserService) UpdateUser(ctx context.Context, userID int32, updates *UserUpdates) error {
    // 1. Update database
    err := s.repo.UpdateUser(ctx, userID, updates)
    if err != nil {
        return err
    }

    // 2. Invalidate cache
    cacheKey := fmt.Sprintf("user:%d", userID)
    s.cache.Delete(ctx, cacheKey)

    return nil
}
```

## Available Methods

```go
type Client interface {
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    Get(ctx context.Context, key string) (string, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
}
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_HOST` | `localhost` | Redis server host |
| `REDIS_PORT` | `6379` | Redis server port |
| `REDIS_PASSWORD` | `` | Redis password (optional) |
| `REDIS_DB` | `0` | Redis database number |

## Common Patterns

**Session storage:**
```go
sessionKey := fmt.Sprintf("session:%s", sessionID)
s.cache.Set(ctx, sessionKey, userID, 24*time.Hour)
```

**Rate limiting:**
```go
key := fmt.Sprintf("ratelimit:%s", userID)
s.cache.Set(ctx, key, "1", 1*time.Minute)
```

**Temporary data:**
```go
key := fmt.Sprintf("temp:%s", requestID)
s.cache.Set(ctx, key, data, 5*time.Minute)
```

## TTL Guidelines

- **User sessions**: 24 hours
- **API responses**: 5-15 minutes
- **Rate limit counters**: 1 minute
- **Temporary data**: 5-10 minutes

That's it! Just inject `redis.Client` and start caching.
