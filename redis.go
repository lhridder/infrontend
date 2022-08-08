package infrontend

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"log"
	"reflect"
	"strings"
	"time"
)

const expiration = 10 * time.Minute

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

func GetSession(uuid uuid.UUID) (Session, error) {
	item, err := rdb.Get(ctx, "session:"+uuid.String()).Result()
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

func FindUser(search string) (User, uuid.UUID, error) {
	var user User

	users, err := rdb.Keys(ctx, "user:*").Result()
	if err != nil {
		return user, uuid.UUID{}, err
	}

	var useruuid uuid.UUID
	for _, item := range users {
		useruuid, err = uuid.Parse(strings.Split(item, "user:")[1])
		user, err = GetUser(useruuid)
		if err != nil {
			return user, uuid.UUID{}, err
		}
		if user.Email == search || user.Username == search {
			break
		}
	}

	return user, useruuid, nil
}

func StoreUser(user User) error {
	serializedUser, err := json.Marshal(user)
	if err != nil {
		return err
	}

	var useruuid uuid.UUID
	for {
		useruuid = uuid.New()

		search, err := GetUser(useruuid)
		if err != nil {
			return err
		}

		if reflect.ValueOf(search).IsZero() {
			break
		}
	}

	_, err = rdb.Set(ctx, "user:"+useruuid.String(), string(serializedUser), 0).Result()
	if err != nil {
		return err
	}

	return nil
}

func StoreSession(session Session) (uuid.UUID, error) {
	serializedSession, err := json.Marshal(session)
	if err != nil {
		return uuid.UUID{}, err
	}

	var sessionuuid uuid.UUID
	for {
		sessionuuid = uuid.New()

		search, err := GetSession(sessionuuid)
		if err != nil {
			return uuid.UUID{}, err
		}

		if reflect.ValueOf(search).IsZero() {
			break
		}
	}

	_, err = rdb.Set(ctx, "session:"+sessionuuid.String(), string(serializedSession), expiration).Result()
	if err != nil {
		return uuid.UUID{}, err
	}

	return sessionuuid, nil
}
