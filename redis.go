package infrontend

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"log"
)

var (
	rdb *redis.Client
	ctx = context.Background()
)

func ConnectRedis() error {
	rediscfg := GlobalConfig.Redis

	log.Println("Connecting to redis")
	ctx = context.Background()

	rdb = redis.NewClient(&redis.Options{
		Addr:     rediscfg.Host + ":6379",
		Password: rediscfg.Pass,
		DB:       rediscfg.DB,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return err
	}
	return nil
}

func GetToken(token string) (APItoken, error) {
	item, err := rdb.Get(ctx, "token:"+token).Result()
	if err != nil {
		if err == redis.Nil {
			return APItoken{}, nil
		}
		return APItoken{}, err
	}

	var apitoken APItoken
	err = json.Unmarshal([]byte(item), &apitoken)
	if err != nil {
		return APItoken{}, err
	}

	return apitoken, nil
}

func GetUser(uuid uuid.UUID) (User, error) {
	item, err := rdb.Get(ctx, "user:"+uuid.String()).Result()
	if err != nil {
		if err == redis.Nil {
			return User{}, nil
		}
		return User{}, err
	}

	var user User
	err = json.Unmarshal([]byte(item), &user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func GetSession(cookie string) (Session, error) {
	item, err := rdb.Get(ctx, "session:"+cookie).Result()
	if err != nil {
		if err == redis.Nil {
			return Session{}, nil
		}
		return Session{}, err
	}

	var session Session
	err = json.Unmarshal([]byte(item), &session)
	if err != nil {
		return Session{}, err
	}

	return session, nil
}
