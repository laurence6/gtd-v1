package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"time"

	"gopkg.in/redis.v4"

	"github.com/laurence6/gtd.go/model"
)

func EncPassword(password string) string {
	h := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(h[:])
}

// CheckPassword returns UserID if UserID and password are valid.
func CheckPassword(userID, password string) (string, error) {
	if userID == "" || password == "" {
		return "", errors.New("Empty UserID or Password")
	}

	user, err := model.GetUser(userID)
	if err != nil {
		return "", err
	}

	if user.Password == password {
		return userID, nil
	}

	return "", nil
}

// NewToken returns a new token.
func NewToken() string {
	b := make([]byte, 32)
	io.ReadFull(rand.Reader, b)
	return base64.URLEncoding.EncodeToString(b)
}

// SetToken stores UserID and token in redis using "token:xxx" as key. An expiration in second is taken.
func SetToken(userID, token string, expires int) error {
	err := redisClient.Set(getNamespace("tok", token), userID, time.Duration(expires)*1e9).Err()
	return err
}

// CheckToken returns UserID if token is valid.
func CheckToken(token string) (string, error) {
	userID, err := redisClient.Get(getNamespace("tok", token)).Result()
	if err == nil {
		return userID, nil
	}

	if err == redis.Nil {
		return "", nil
	}

	return "", err
}
