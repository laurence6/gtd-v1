package main

import (
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/laurence6/gtd.go/model"

	"gopkg.in/pg.v4"
	"gopkg.in/redis.v4"
)

var logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

const confPath = "conf"

var conf *Conf

var redisClient *redis.Client

func init() {
	var err error

	conf, err = parseConfFile(confPath)
	if err != nil {
		logger.Fatalln(err)
	}

	// Redis Client
	redisClient = redis.NewClient(&conf.RedisOptions)

	// PostgreSQL
	db := pg.Connect(&conf.PgOptions)
	model.DBConn = db

	// Signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		logger.Println("Interrupt received, bye")
		os.Exit(0)
	}()
}

func main() {
	logger.Println("PID:", os.Getpid())

	logger.Println("Start web server")
	web()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
