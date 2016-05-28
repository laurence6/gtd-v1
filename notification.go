package main

import (
	"log"
	"time"

	"github.com/laurence6/gtd.go/core"
)

// Notifier is the notifier interface.
type Notifier interface {
	Notify(*gtd.Task)
}

var notifiers []Notifier

var notificationIndex []*gtd.Task

func init() {
	notifiers = []Notifier{}
	notifiers = append(notifiers, &Stdout{})

	notificationIndex = []*gtd.Task{}
}

func notification() {
	c := make(chan int)
	go func() {
		for {
			rebuildNotificationIndex()
			c <- 1
			tp.Lock()
			tp.Wait()
		}
	}()
	go func() {
		for {
			if len(notificationIndex) > 0 {
				select {
				case <-time.After(time.Duration((notificationIndex[0].Notification.Get() - time.Now().Unix()) * 1e9)):
					taskList := []*gtd.Task{}
					for n, i := range notificationIndex {
						if n == 0 || i.Notification.Get() == notificationIndex[n-1].Notification.Get() {
							taskList = append(taskList, i)
						} else {
							break
						}
					}
					go notify(taskList)
					<-time.After(time.Second)
					rebuildNotificationIndex()
				case <-c:
					continue
				}
			} else {
				<-c
			}
		}
	}()
}

func notify(taskList []*gtd.Task) {
	for _, i := range taskList {
		if tp.Has(i.ID) {
			for _, notifier := range notifiers {
				go notifier.Notify(i)
			}
		}
	}
}

func rebuildNotificationIndex() {
	taskList, _ := tp.FindAll(func(task *gtd.Task) bool {
		if !task.Notification.EqualZero() {
			return true
		}
		return false
	})
	notificationIndex = notificationIndex[:0]
	now := time.Now().Unix()
	gtd.SortByNotification(taskList)
	for _, i := range taskList {
		notification := i.Notification.Get()
		if notification < now {
			continue
		}
		notificationIndex = append(notificationIndex, i)
	}
}

type Stdout struct {
}

func (s *Stdout) Notify(task *gtd.Task) {
	log.Println(task.Subject, task.Notification.Get())
}
