package redis

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/danielhoward314/packet-sentry/dao"
)

const (
	registrationTokenTTL = time.Minute * 10
	characterSpace       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type registrationDatastore struct {
	client *redis.Client
}

// NewRegistrationDatastore returns a redis implementation for the registration key-value datastore
func NewRegistrationDatastore(client *redis.Client) dao.RegistrationDatastore {
	return &registrationDatastore{
		client: client,
	}
}

func (rds *registrationDatastore) Create(registration *dao.Registration) (string, string, error) {
	if registration.OrganizationID == "" {
		return "", "", errors.New("invalid organization id")
	}
	if registration.AdministratorID == "" {
		return "", "", errors.New("invalid administrator id")
	}
	token, err := generateRandomString(60)
	if err != nil {
		return "", "", err
	}
	emailCode, err := generateRandomString(6)
	if err != nil {
		return "", "", err
	}
	registration.EmailCode = emailCode
	registrationJSON, err := json.Marshal(registration)
	if err != nil {
		return "", "", err
	}
	status := rds.client.Set(context.Background(), token, registrationJSON, registrationTokenTTL)
	if status.Err() != nil {
		return "", "", err
	}
	return token, emailCode, nil
}

func (rds *registrationDatastore) Read(token string) (*dao.Registration, error) {
	registrationJSON, err := rds.client.Get(context.Background(), token).Result()
	if err != nil {
		return nil, err
	}
	var registration dao.Registration
	err = json.Unmarshal([]byte(registrationJSON), &registration)
	if err != nil {
		return nil, err
	}
	return &registration, nil
}

func (rds *registrationDatastore) Delete(token string) error {
	cmdStatus := rds.client.Del(context.Background(), token)
	if cmdStatus.Err() != nil {
		return cmdStatus.Err()
	}
	return nil
}

func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(characterSpace))))
		if err != nil {
			return "", err
		}
		b[i] = characterSpace[num.Int64()]
	}
	return string(b), nil
}
