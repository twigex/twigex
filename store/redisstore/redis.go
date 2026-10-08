// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package redisstore

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"

	"github.com/go-redis/redis"
)

type redisRepository struct {
	client *redis.Client
}

const (
	DBPingAttempts    = 18
	DBPingTimeoutSecs = 10
	ExitPing          = 102
)

func NewRedisClient(o *redis.Options) (*redisRepository, error) {
	rdb := redis.NewClient(o)

	for i := 0; i < DBPingAttempts; i++ {
		status := rdb.Ping()
		str, _ := status.Result()
		if str != "" {
			break
		} else {
			if i == DBPingAttempts-1 {
				tlog.Error("failed to reach redis at %s. exiting", o.Addr, DBPingTimeoutSecs)
				time.Sleep(time.Second)
				os.Exit(ExitPing)
			} else {
				tlog.Warn("failed to reach redis at %s. retrying in %d seconds", o.Addr, DBPingTimeoutSecs)
				time.Sleep(DBPingTimeoutSecs * time.Second)
			}
		}
	}

	r := &redisRepository{}
	r.client = rdb

	return r, nil
}

func (r *redisRepository) Set(key string, expireTime time.Duration, value interface{}) error {
	p, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = r.client.Set(key, p, expireTime).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *redisRepository) Get(key string, dest interface{}) error {
	p, err := r.client.Get(key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(p), dest)
}

func (r *redisRepository) GetSessions(userID string) ([]model.SessionDetails, error) {
	sessions := make([]model.SessionDetails, 0)

	iter := r.client.Scan(0, "*", 0).Iterator()
	for iter.Next() {
		p, err := r.client.Get(iter.Val()).Result()
		if err != nil {
			return nil, err
		}

		var session model.SessionDetails
		if err := json.Unmarshal([]byte(p), &session); err != nil {
			continue
		}

		if session.UserID == userID {
			sessions = append(sessions, session)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *redisRepository) UpdateSession(session model.SessionDetails, sessionLength int) error {
	p, err := json.Marshal(session)
	if err != nil {
		return err
	}

	err = r.client.Set(session.AccessToken, p, time.Hour*24*time.Duration(sessionLength)).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *redisRepository) UpdateLastActivity(session model.SessionDetails) error {
	p, err := json.Marshal(session)
	if err != nil {
		return err
	}

	time, err := r.client.TTL(session.AccessToken).Result()
	if err != nil {
		return err
	}

	err = r.client.Set(session.AccessToken, p, time).Err()
	if err != nil {
		return err
	}

	return nil
}

// GetUserID returns authorized user
func (r *redisRepository) GetUserID(req *http.Request) (*string, error) {
	sidString, err := req.Cookie(model.SessionCookieToken)
	if err != nil {
		return nil, err
	}

	var l model.SessionDetails
	if err := r.Get(sidString.Value, &l); err != nil {
		return nil, err
	}

	return &l.UserID, nil
}

func (r *redisRepository) Delete(key string) error {
	err := r.client.Del(key).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *redisRepository) Incr(key string, ttl time.Duration) (int64, error) {
	counter, err := r.client.Incr(key).Result()
	if err != nil {
		return 0, err
	}

	if counter == 1 {
		if err := r.client.Expire(key, ttl).Err(); err != nil {
			r.client.Del(key) // don't leave a counter with no expiry
			return 0, err
		}
	}

	return counter, nil
}
