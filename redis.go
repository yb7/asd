package asd

import (
  "github.com/gomodule/redigo/redis"
  "time"
)

func newRedisPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle: 3,
		IdleTimeout: 240 * time.Second,
		Dial: func () (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

var (
	redisPool *redis.Pool
)

func InitRedisPool(redisServer string) {
  redisPool = newRedisPool(redisServer)
}
