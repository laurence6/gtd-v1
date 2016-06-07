package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/laurence6/gtd.go/core"
)

var tp *gtd.TaskPool
var rw = &sync.RWMutex{}

var defaultIndex = []*gtd.Task{}

var session = map[string]interface{}{}

var wg = &sync.WaitGroup{}

func init() {
	sessionFile, err := os.Open("session")
	if err == nil {
		dec := json.NewDecoder(sessionFile)
		err = dec.Decode(&session)
		if err != nil {
			log.Fatalln(err)
		}
		sessionFile.Close()
	} else {
		log.Println(err)
	}
	log.Println(session)

	tpFile, err := os.Open("taskpool")
	if err == nil {
		tp, err = gtd.UnmarshalTaskPool(tpFile)
		if err != nil {
			log.Fatalln(err)
		}
		tpFile.Close()
	} else {
		log.Println(err)
		tp = gtd.NewTaskPool()
	}

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
	rw.RLock()
	defaultIndex = tp.GetAll()
	rw.RUnlock()
	gtd.SortByDefault(defaultIndex)
}

func backupTaskPool() {
	os.Rename("taskpool", "taskpool.bak")
	tpFile, err := os.OpenFile("taskpool", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
	}
	rw.RLock()
	err = tp.Marshal(tpFile)
	rw.RUnlock()
	if err != nil {
		log.Println(err)
	}
	tpFile.Close()
}

func main() {
	// When backup has better performence
	//tp.HookFunc(backupTaskPool)

	tp.HookFunc(rebuildDefaultIndex)
	rebuildDefaultIndex()

	log.Println("Start web server")
	web()
	log.Println("Start notification")
	notification()

	wg.Add(1)
	wg.Wait()
}
