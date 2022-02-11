package gotools

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type MyRedis struct {
	*redis.Client
}

func (myRedis *MyRedis) acquireLock(lockName string, acquireTimeout int) string {
	end := time.Now().Add(time.Duration(acquireTimeout) * time.Second)
	identifier := uuid.New().String()
	lockName = fmt.Sprintf("lock:%s", lockName)
	for time.Now().Before(end) {
		res, _ := myRedis.SetNX(context.TODO(), lockName, identifier, time.Duration(1)*time.Second).Result()
		if res {
			return identifier
		} else {
			time.Sleep(time.Millisecond)
		}
	}
	return ""
}

func (myRedis *MyRedis) releaseLock(lockName, identifier string) bool {
	ctx := context.TODO()
	lockName = fmt.Sprintf("lock:%s", lockName)
	transactionFuc := func(t *redis.Tx) error {
		res, _ := t.Get(ctx, lockName).Result()
		if res == identifier {
			_, err := myRedis.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Del(ctx, lockName)
				return nil
			})
			return err
		}
		return nil
	}

	for {
		err := myRedis.Watch(ctx, transactionFuc, lockName) // Watch 锁
		if err == nil {
			return true
		}
	}
}
