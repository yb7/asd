package asd

import (
  "github.com/gomodule/redigo/redis"
  "time"
  "strconv"
  "os"
)

var (
  redisPool *redis.Pool
)
type RedisOptions struct {
  Host string
  Password string
  DbNo     int
  MaxIdle  string
}

func newRedisPool(opt RedisOptions) *redis.Pool {
  var maxIdle = 3
  if i, err := strconv.Atoi(opt.MaxIdle); err == nil {
    maxIdle = i
  }
	return &redis.Pool{
		MaxIdle: maxIdle,
		IdleTimeout: 240 * time.Second,
    Dial: func () (redis.Conn, error) {
      c, err := redis.Dial("tcp", opt.Host)
      if err != nil {
        return nil, err
      }
      if opt.Password != "" {
        if _, err := c.Do("AUTH", opt.Password); err != nil {
          c.Close()
          return nil, err
        }
      }

      if _, err := c.Do("SELECT", opt.DbNo); err != nil {
        c.Close()
        return nil, err
      }
      return c, nil
    },
	}
}

func InitRedisPool(in RedisOptions) {
  var opt = RedisOptions{}
  if in.Host != "" {
    opt = in
  } else {
    opt.Host = os.Getenv("REDIS_ADDR")
    opt.Password = os.Getenv("REDIS_PWD")
    opt.DbNo = os.Getenv("REDIS_DB")
    opt.MaxIdle = os.Getenv("REDIS_MAXIDLE")
  }
  redisPool = newRedisPool(opt)
}
