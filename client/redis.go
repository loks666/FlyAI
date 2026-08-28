package client

import (
	"context"
	"github.com/loks666/FlyAI/config"
	redis "github.com/redis/go-redis/v9"
	"log"
	"sync"
	"time"
)

var (
	redisMu     sync.RWMutex
	redisClient *redis.Client
)

func InitRedisClient() {
	if !config.Conf.Redis.Enable {
		closeRedisClient(swapRedisClient(nil))
		log.Printf("Redis 未启用，频率限制使用进程内存，通讯录读取使用 MySQL")
		return
	}

	candidate := redis.NewClient(&redis.Options{
		Addr:     config.Conf.Redis.Address(),
		Password: config.Conf.Redis.Password,
		DB:       config.Conf.Redis.Db,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := candidate.Ping(ctx).Err(); err != nil {
		_ = candidate.Close()
		closeRedisClient(swapRedisClient(nil))
		log.Printf("Redis 连接失败，频率限制降级为进程内存，通讯录读取降级为 MySQL: %v", err)
		return
	}

	closeRedisClient(swapRedisClient(candidate))
	log.Printf("Redis连接成功: %s db=%d", config.Conf.Redis.Address(), config.Conf.Redis.Db)
}

func swapRedisClient(candidate *redis.Client) *redis.Client {
	redisMu.Lock()
	previous := redisClient
	redisClient = candidate
	redisMu.Unlock()
	return previous
}

func closeRedisClient(redisClient *redis.Client) {
	if redisClient != nil {
		_ = redisClient.Close()
	}
}

func GetRedisClient() *redis.Client {
	redisMu.RLock()
	defer redisMu.RUnlock()
	return redisClient
}
