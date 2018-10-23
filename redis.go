package asd

import (
  "github.com/gomodule/redigo/redis"
  "time"
)

func newRedisPool(addr, pwd, db string) *redis.Pool {
	return &redis.Pool{
		MaxIdle: 3,
		IdleTimeout: 240 * time.Second,
		// Dial: func () (redis.Conn, error) { return redis.Dial("tcp", addr) },
    Dial: func () (redis.Conn, error) {
      c, err := redis.Dial("tcp", addr)
      if err != nil {
        return nil, err
      }
      if _, err := c.Do("AUTH", pwd); err != nil {
        c.Close()
        return nil, err
      }
      if _, err := c.Do("SELECT", db); err != nil {
        c.Close()
        return nil, err
      }
      return c, nil
    },
	}
}

var (
	redisPool *redis.Pool
)

func InitRedisPool(redisServer, pwd, db string) {
  redisPool = newRedisPool(redisServer, pwd, db)
}
