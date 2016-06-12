package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/laurence6/gtd.go/core"
)

var conf = map[string]interface{}{}

var redisClient *Client

var tp *gtd.TaskPool
var tpRW = &sync.RWMutex{}

var defaultIndex = []*gtd.Task{}

var wg = &sync.WaitGroup{}

func init() {
	// Conf
	confFile, err := os.Open("conf")
	if err == nil {
		dec := json.NewDecoder(confFile)
		err = dec.Decode(&conf)
		if err != nil {
			log.Fatalln(err.Error())
		}
		confFile.Close()
	} else {
		log.Fatalln(err.Error())
	}
	log.Println("Conf:", conf)

	// Redis Client
	redisAddr, ok := conf["redis_addr"]
	if !ok {
		log.Fatalln("Cannot get redis server addr 'redis_addr'")
	}
	redisClient = NewRedisClient(redisAddr.(string))
	if !redisClient.IsOnline() {
		log.Fatalln("Redis server offline")
	}

	// TaskPool
	tp = gtd.NewTaskPool()
	serializedTp, err := redisClient.Get("taskpool").Bytes()
	if err == nil {
		err = tp.Unmarshal(serializedTp)
		if err != nil {
			log.Fatalln(err.Error())
		}
	} else {
		log.Println(err.Error())
	}

	// Signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("Interrupt received, bye")
		backupTaskPool()
		os.Exit(0)
	}()
}

func rebuildDefaultIndex() {
	tpRW.RLock()
	defaultIndex = tp.GetAll()
	tpRW.RUnlock()

	gtd.SortByDefault(defaultIndex)
}

func backupTaskPool() {
	buf := &bytes.Buffer{}

	tpRW.RLock()
	err := tp.Marshal(buf)
	tpRW.RUnlock()
	if err != nil {
		log.Println(err.Error())
	}

	err = redisClient.Set("taskpool", buf.String(), 0).Err()
	if err != nil {
		log.Println(err.Error())
	}
}

func main() {
	tp.HookFunc(backupTaskPool)

	rebuildDefaultIndex()
	tp.HookFunc(rebuildDefaultIndex)

	log.Println("Start web server")
	web()
	log.Println("Start notification")
	notification()

	wg.Add(1)
	wg.Wait()
}
