package main

import (
	"fmt"
	"time"

	"github.com/laurence6/gtd.go/core"
)

var tp gtd.TaskPool

func init() {
	tp = gtd.TaskPool{}
}

func main() {
	web()
}
