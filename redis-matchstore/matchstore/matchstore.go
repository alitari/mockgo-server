package matchstore

import (
	"context"
	"encoding/json"
	"github.com/alitari/mockgo-server/mockgo/matches"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type redisMatchstore struct {
	client        *redis.Client
	capacity      uint16
	mismatchesKey string
}

func (r *redisMatchstore) checkConnectivity() error {
	ctx := context.Background()
	status := r.client.Ping(ctx)
	return status.Err()
}

func (r *redisMatchstore) GetMatches(endpointID string) ([]*matches.Match, error) {
	ctx := context.Background()
	lrange := r.client.LRange(ctx, endpointID, 0, -1)
	if lrange.Err() != nil {
		return nil, lrange.Err()
	}
	var result []*matches.Match
	for _, v := range lrange.Val() {
		var match matches.Match
		err := json.Unmarshal([]byte(v), &match)
		if err != nil {
			return nil, err
		}
		result = append(result, &match)
	}
	return result, nil
}

func (r *redisMatchstore) GetMatchesCount(endpointID string) (uint64, error) {
	ctx := context.Background()
	llen := r.client.LLen(ctx, endpointID)
	if llen.Err() != nil {
		return 0, llen.Err()
	}
	return uint64(llen.Val()), nil
}

func (r *redisMatchstore) GetMismatches() ([]*matches.Mismatch, error) {
	ctx := context.Background()
	lrange := r.client.LRange(ctx, r.mismatchesKey, 0, -1)
	if lrange.Err() != nil {
		return nil, lrange.Err()
	}
	var result []*matches.Mismatch
	for _, v := range lrange.Val() {
		var mismatch matches.Mismatch
		err := json.Unmarshal([]byte(v), &mismatch)
		if err != nil {
			return nil, err
		}
		result = append(result, &mismatch)
	}
	return result, nil
}

func (r *redisMatchstore) AddMatch(endpointID string, match *matches.Match) error {
	ctx := context.Background()
	// marshal match
	mval, err := json.Marshal(match)
	if err != nil {
		return err
	}
	lpush := r.client.RPush(ctx, endpointID, mval)
	if lpush.Err() != nil {
		return lpush.Err()
	}
	return nil
}

func (r *redisMatchstore) AddMismatch(mismatch *matches.Mismatch) error {
	ctx := context.Background()
	// marshal match
	mval, err := json.Marshal(mismatch)
	if err != nil {
		return err
	}
	lpush := r.client.RPush(ctx, r.mismatchesKey, mval)
	if lpush.Err() != nil {
		return lpush.Err()
	}
	return nil
}

func (r *redisMatchstore) GetMismatchesCount() (uint64, error) {
	ctx := context.Background()
	llen := r.client.LLen(ctx, r.mismatchesKey)
	if llen.Err() != nil {
		return 0, llen.Err()
	}
	return uint64(llen.Val()), nil
}

func (r *redisMatchstore) DeleteMatches(endpointID string) error {
	ctx := context.Background()
	del := r.client.Del(ctx, endpointID)
	if del.Err() != nil {
		return del.Err()
	}
	return nil
}

func (r *redisMatchstore) DeleteMismatches() error {
	ctx := context.Background()
	del := r.client.Del(ctx, r.mismatchesKey)
	if del.Err() != nil {
		return del.Err()
	}
	return nil
}

// NewRedisMatchstore creates a new redis matchstore
func NewRedisMatchstore(address, password string, db int, capacity uint16) (matches.Matchstore, error) {
	// create uuid string
	uuid, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	mismatchesKey := "mismatches_" + uuid.String()
	matchstore := &redisMatchstore{
		client: redis.NewClient(&redis.Options{
			Addr:     address,
			Password: password,
			DB:       db,
		}),
		capacity:      capacity,
		mismatchesKey: mismatchesKey,
	}
	if err := matchstore.checkConnectivity(); err != nil {
		return nil, err
	}
	return matchstore, nil
}
