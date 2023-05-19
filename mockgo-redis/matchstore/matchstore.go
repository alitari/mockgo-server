package matchstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/alitari/mockgo-server/mockgo/matches"
	"github.com/redis/go-redis/v9"
)

const mismatchesKey = "__mismatches__"
const counterKey = "__counter__"

type redisMatchstore struct {
	client   *redis.Client
	capacity uint16
}

func (r *redisMatchstore) checkConnectivity() error {
	var ctx = context.Background()
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
	get := r.client.Get(ctx, endpointID+counterKey)
	if get.Err() != nil {
		return 0, get.Err()
	}
	return get.Uint64()
}

func (r *redisMatchstore) GetMismatches() ([]*matches.Mismatch, error) {
	ctx := context.Background()
	lrange := r.client.LRange(ctx, mismatchesKey, 0, -1)
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
	mval, err := json.Marshal(match)
	if err != nil {
		return err
	}
	rpush := r.client.RPush(ctx, endpointID, mval)
	if rpush.Err() != nil {
		return rpush.Err()
	}
	if r.client.LLen(ctx, endpointID).Val() > int64(r.capacity) {
		lpop := r.client.LPop(ctx, endpointID)
		if lpop.Err() != nil {
			return lpop.Err()
		}
	}
	incr := r.client.Incr(ctx, endpointID+counterKey)
	if incr.Err() != nil {
		return incr.Err()
	}
	return nil
}

func (r *redisMatchstore) AddMismatch(mismatch *matches.Mismatch) error {
	ctx := context.Background()
	mval, err := json.Marshal(mismatch)
	if err != nil {
		return err
	}
	lpush := r.client.RPush(ctx, mismatchesKey, mval)
	if lpush.Err() != nil {
		return lpush.Err()
	}
	if r.client.LLen(ctx, mismatchesKey).Val() > int64(r.capacity) {
		lpop := r.client.LPop(ctx, mismatchesKey)
		if lpop.Err() != nil {
			return lpop.Err()
		}
	}
	incr := r.client.Incr(ctx, mismatchesKey+counterKey)
	if incr.Err() != nil {
		return incr.Err()
	}
	return nil
}

func (r *redisMatchstore) GetMismatchesCount() (uint64, error) {
	ctx := context.Background()
	get := r.client.Get(ctx, mismatchesKey+counterKey)
	if get.Err() != nil {
		return 0, get.Err()
	}
	return get.Uint64()
}

func (r *redisMatchstore) DeleteMatches(endpointID string) error {
	ctx := context.Background()
	del := r.client.Del(ctx, endpointID)
	if del.Err() != nil {
		return del.Err()
	}
	set := r.client.Set(ctx, endpointID+counterKey, 0, 0)
	if set.Err() != nil {
		return set.Err()
	}
	return nil
}

func (r *redisMatchstore) DeleteMismatches() error {
	ctx := context.Background()
	del := r.client.Del(ctx, mismatchesKey)
	if del.Err() != nil {
		return del.Err()
	}
	set := r.client.Set(ctx, mismatchesKey+counterKey, 0, 0)
	if set.Err() != nil {
		return set.Err()
	}
	return nil
}

// NewRedisMatchstore creates a new redis matchstore
func NewRedisMatchstore(address, password string, db int, capacity uint16) (matches.Matchstore, error) {
	matchstore := &redisMatchstore{
		client: redis.NewClient(&redis.Options{
			Addr:            address,
			Password:        password,
			DB:              db,
			MaxRetries:      30,
			MinRetryBackoff: 500 * time.Millisecond,
			MaxRetryBackoff: 2 * time.Second,
		}),
		capacity: capacity,
	}
	if err := matchstore.checkConnectivity(); err != nil {
		return nil, err
	}
	return matchstore, nil
}
