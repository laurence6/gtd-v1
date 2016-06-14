package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"time"
)

var ErrPasswordNotFound = errors.New("Cannot find password 'password' in conf")

func encPassword(password string) string {
	h := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(h[:])
}

// CheckPassword returns true if password is valid.
func CheckPassword(password string) (bool, error) {
	if password == "" {
		return false, nil
	}
	storedPassword, ok := conf["password"].(string)
	if !ok {
		return false, ErrPasswordNotFound
	}
	if encPassword(password) == storedPassword {
		return true, nil
	}
	return false, nil
}

// CheckToken returns true if token is valid.
func CheckToken(token string) (bool, error) {
	err := redisClient.Get(GetNamespace("tok", token)).Err()
	if err == nil {
		return true, nil
	}
	return false, err
}

// SetToken sets token. Use "token:xxx" as key. An expiration in second is taken.
func SetToken(token string, expires int) error {
	err := redisClient.Set(GetNamespace("tok", token), "0", time.Duration(expires)*1e9).Err()
	if err != nil {
		return err
	}
	return nil
}

// NewToken returns a new token.
func NewToken() string {
	b := make([]byte, 32)
	io.ReadFull(rand.Reader, b)
	return base64.URLEncoding.EncodeToString(b)
}
