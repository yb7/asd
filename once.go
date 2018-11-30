package asd

import (
  "encoding/json"
  "errors"
  "fmt"
  "github.com/gomodule/redigo/redis"
  "reflect"
  "sync"
  "time"
  "math"
)

type onceVo struct {
  Once *sync.Once
  ExpiresAt time.Time
  Data interface{}
}

var onceMap sync.Map
var lockMap sync.Map

func init () {
  go func() {
    for true {
      clearExpiresKey()
      time.Sleep(1 * time.Second)
    }
  }()
}

func clearExpiresKey() {
  onceMap.Range(func(key, value interface{}) bool {
    v := value.(*onceVo)
    if v.ExpiresAt.Add(1 * time.Second).Before(time.Now()) {
      onceMap.Delete(key)
    }
    return true
  })
}

func setV(source, dst interface{}) error {
  // ValueOf to enter reflect-land
  dstPtrValue := reflect.ValueOf(dst)
  if dstPtrValue.Kind() != reflect.Ptr {
    return errors.New("destination must be kind of ptr")
  }
  if dstPtrValue.IsNil() {
    return errors.New("destination cannot be nil")
  }
  //dstType := dstPtrType.Elem()
  // the *dst in *dst = zero
  dstValue := reflect.Indirect(dstPtrValue)
  // the = in *dst = 0
  dstValue.Set(reflect.ValueOf(source))
  return nil
}

func unmarshalFromRedis(conn redis.Conn, key string, dst interface{}) error {
  bytes, err := redis.Bytes(conn.Do("GET", key))
  if err != nil {
    return err
  }
  return json.Unmarshal(bytes, dst)
}

func loadOnce(key string, duration time.Duration) *onceVo {
  lockI, _ := lockMap.LoadOrStore(key, &sync.Mutex{})
  lock := lockI.(*sync.Mutex)

  lock.Lock()

  onceObj, ok := onceMap.Load(key)
  if ok {
    once := onceObj.(*onceVo)
    if once.ExpiresAt.Before(time.Now()) {
      onceMap.Delete(key)
    }
  }
  onceObj, ok = onceMap.Load(key)
  if !ok {
    onceObj = &onceVo{
      Once: &sync.Once{},
      ExpiresAt: time.Now().Add(duration),
    }
    onceMap.Store(key, onceObj)
  }

  lock.Unlock()
  return onceObj.(*onceVo)
}
func OnceInMem(key string, duration time.Duration, fallback func() (interface{}, error), dst interface{}) error {
  newOnce := loadOnce(key, duration)

  var err error
  newOnce.Once.Do(func() {
    var result interface{}
    result, err = fallback()
    if err == nil {
      newOnce.Data = result
      onceMap.Store(key, newOnce)
    }
  })
  if err != nil {
    onceMap.Delete(key)
    return err
  } else {
    onceObj, ok := onceMap.Load(key)
    if ok {
      once := onceObj.(*onceVo)
      if once.Data != nil {
        setV(once.Data, dst)
      } else {
        onceMap.Delete(key)
      }
    }
  }
  return nil
}


func OnceInRedis(key string, duration time.Duration, fallback func() (interface{}, error), dst interface{}) error {
  newOnce := loadOnce(key, duration)

  conn := redisPool.Get()
  defer conn.Close()

  var hasValue = false

  var err error
  newOnce.Once.Do(func() {
    var result interface{}
    result, err = fallback()

    if err != nil {
      fmt.Errorf("get data error: %s", err.Error())
      //conn.Do("DEL", key) 不用删除，等待自动过期
    } else {
      if err = setV(result, dst); err != nil {
        return
      } else {
        hasValue = true
      }
      var bytes []byte
      bytes, err = json.Marshal(result)
      if err != nil {
        return
      }
      _, err = conn.Do("SET", key, bytes)
      if err != nil {
        return
      }

      var expireTime = math.Max(math.Ceil(duration.Seconds() * 2), 1)
      _, err = conn.Do("EXPIRE", key, expireTime)

    }
  })
  if err != nil {
    onceMap.Delete(key)
    return err
  }
  if !hasValue {
    return unmarshalFromRedis(conn, key, dst)
  }
  return nil
}
