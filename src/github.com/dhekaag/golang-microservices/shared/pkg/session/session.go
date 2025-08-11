package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionManager struct {
	redisClient *redis.Client
	prefix      string
	ttl         time.Duration
}

type UserSession struct {
	UserID    uint      `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	LastSeen  time.Time `json:"last_seen"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

type SessionConfig struct {
	RedisAddr     string `json:"redis_addr"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`
	SessionTTL    int    `json:"session_ttl"`
	SessionPrefix string `json:"session_prefix"`
}

func NewSessionManager(config SessionConfig) (*SessionManager, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})
	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &SessionManager{
		redisClient: rdb,
		prefix:      config.SessionPrefix,
		ttl:         time.Duration(config.SessionTTL) * time.Second,
	}, nil
}

func (sm *SessionManager) getSessionKey(sessionID string) string {
	return fmt.Sprintf("%s:%s", sm.prefix, sessionID)
}

func (sm *SessionManager) CreateSession(ctx context.Context, sessionID string, userSession *UserSession) error {
	sessionKey := sm.getSessionKey(sessionID)
	data, err := json.Marshal(userSession)
	if err != nil {
		return fmt.Errorf("failed to marshal user session: %w", err)
	}
	err = sm.redisClient.Set(ctx, sessionKey, data, sm.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*UserSession, error) {
	sessionKey := sm.getSessionKey(sessionID)
	data, error := sm.redisClient.Get(ctx, sessionKey).Result()

	if error != nil {
		if error == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", error)
	}

	var userSession UserSession

	if err := json.Unmarshal([]byte(data), &userSession); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user session: %w", err)
	}

	// update last seen time
	userSession.LastSeen = time.Now()
	if err := sm.UpdateSession(ctx, sessionID, &userSession); err != nil {
		return nil, fmt.Errorf("failed to update last seen time: %w", err)
	}
	return &userSession, nil
}

func (sm *SessionManager) UpdateSession(ctx context.Context, sessionID string, userSession *UserSession) error {
	sessionKey := sm.getSessionKey(sessionID)
	data, err := json.Marshal(userSession)
	if err != nil {
		return fmt.Errorf("failed to marshal user session: %w", err)
	}
	err = sm.redisClient.Set(ctx, sessionKey, data, sm.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	return nil
}

func (sm *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	sessionKey := sm.getSessionKey(sessionID)
	err := sm.redisClient.Del(ctx, sessionKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (sm *SessionManager) ExtendSession(ctx context.Context, sessionID string) error {
	sessionKey := sm.getSessionKey(sessionID)
	err := sm.redisClient.Expire(ctx, sessionKey, sm.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to extend session: %w", err)
	}
	return nil
}

func (sm *SessionManager) GetSessions(ctx context.Context) ([]*UserSession, error) {
	var sessions []*UserSession

	// Get all keys matching the session prefix
	keys, err := sm.redisClient.Keys(ctx, fmt.Sprintf("%s:*", sm.prefix)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session keys: %w", err)
	}

	// Retrieve each session
	for _, key := range keys {
		data, err := sm.redisClient.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, fmt.Errorf("failed to get session: %w", err)
		}

		var userSession UserSession
		if err := json.Unmarshal([]byte(data), &userSession); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user session: %w", err)
		}
		sessions = append(sessions, &userSession)
	}

	return sessions, nil
}

func (sm *SessionManager) DeleteSessions(ctx context.Context, userID uint) error {
	// Get all keys matching the session prefix
	keys, err := sm.redisClient.Keys(ctx, fmt.Sprintf("%s:*", sm.prefix)).Result()
	if err != nil {
		return fmt.Errorf("failed to get session keys: %w", err)
	}

	for _, key := range keys {
		data, err := sm.redisClient.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return fmt.Errorf("failed to get session: %w", err)
		}

		var userSession UserSession
		if err := json.Unmarshal([]byte(data), &userSession); err != nil {
			return fmt.Errorf("failed to unmarshal user session: %w", err)
		}

		if userSession.UserID == userID {
			if err := sm.redisClient.Del(ctx, key).Err(); err != nil {
				return fmt.Errorf("failed to delete session: %w", err)
			}
		}
	}

	return nil
}

func (sm *SessionManager) Close() error {
	if sm.redisClient != nil {
		return sm.redisClient.Close()
	}
	return nil
}
