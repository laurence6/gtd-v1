package main

import (
	"strings"

	"gopkg.in/redis.v4"
)

var ErrRedisNil = redis.Nil

// GetNamespace returns a string from args separated by colons.
func GetNamespace(args ...string) string {
	return strings.Join(args, ":")
}

// Client has an embeded pointer to a redis client.
type Client struct {
	*redis.Client
}

// NewRedisClient returns a redis Client instance with specified redis server address.
func NewRedisClient(addr string) *Client {
	return &Client{
		redis.NewClient(
			&redis.Options{Addr: addr},
		),
	}
}

// IsOnline detects if redis server is online.
func (redis *Client) IsOnline() bool {
	err := redis.Ping().Err()
	if err != nil {
		return false
	}
	return true
}
