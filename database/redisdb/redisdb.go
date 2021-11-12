package redisdb

import (
    "context"
    "github.com/go-redis/redis/v8"
    "github.com/jinuopti/lilpop-server/configure"
    "strconv"
    "encoding/json"
    "time"
)

var (
    rdb *redis.Client
    ctx = context.Background()
)

func InitRedis() (*redis.Client, error) {
    if rdb != nil {
        return rdb, nil
    }

    config := configure.GetConfig()
    addr := config.Db.RedisAddr
    port := config.Db.RedisPort

    rdb = redis.NewClient(&redis.Options{
        Addr: addr + ":" + strconv.Itoa(port),
        Password: "",
        DB: 0,
    })

    return rdb, nil
}

func Set(key string, value interface{}, expiration time.Duration) error {
    if rdb == nil {
        _, _ = InitRedis()
    }
    var result *redis.StatusCmd

    switch value.(type) {
    case string:
        result = rdb.Set(ctx, key, value.(string), expiration)
    default:
        p, err := json.Marshal(value)
        if err != nil {
            return err
        }
        result = rdb.Set(ctx, key, p, expiration)
    }

    err := result.Err()
    if err != nil {
        return err
    }

    return nil
}

func SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
    if rdb == nil {
        _, _ = InitRedis()
    }
    p, err := json.Marshal(value)
    if err != nil {
        return false, err
    }
    result := rdb.SetNX(ctx, key, p, expiration)
    err = result.Err()
    if err != nil {
        return false, err
    }
    val := result.Val()

    return val, nil
}

func GetString(key string) (string, error) {
    if rdb == nil {
        _, _ = InitRedis()
    }

    val, err := rdb.Get(ctx, key).Result()
    if err == redis.Nil {
        return "", redis.Nil
    } else if err != nil {
        return "", err
    } else {
        return val, nil
    }
}

func GetStruct(key string, dest interface{}) error {
    if rdb == nil {
        _, _ = InitRedis()
    }
    val, err := rdb.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return redis.Nil
    } else if err != nil {
        return err
    } else {
        err = json.Unmarshal(val, dest)
        if err != nil {
            return err
        }
    }
    return nil
}

func Del(key string) int64 {
    del := rdb.Del(ctx, key).Val()
    return del
}
