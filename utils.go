package main

import (
	"strconv"
	"strings"
	"time"
)

func newID() int64 {
	return time.Now().UnixNano()
}

func stoI64(str string) (int64, error) {
	if str == "" {
		return 0, nil
	}
	i64, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return i64, nil
}

// getNamespace returns a string from args separated by colons.
func getNamespace(args ...string) string {
	return strings.Join(args, ":")
}
