package main

import (
	"fmt"
	"time"

	"github.com/laurence6/gtd.go/core"
)

// Notifier is the notifier interface.
type Notifier interface {
	Notify(*gtd.Task)
}

type notifiedList struct {
	// Use time as key
	m         map[int64][]int64
	keys      []int64
	maxlength int
}

func (n *notifiedList) Add(time, id int64) {
	if _, ok := n.m[time]; !ok {
		n.m[time] = []int64{}
		n.keys = append(n.keys, time)

		if len(n.keys) > n.maxlength {
			n.GC()
		}
	}
	n.m[time] = append(n.m[time], id)
}
func (n *notifiedList) Has(time, id int64) bool {
	if ids, ok := n.m[time]; ok {
		for _, i := range ids {
			if i == id {
				return true
			}
		}
	}
	return false
}
func (n *notifiedList) GC() {
	exceed := len(n.keys) - n.maxlength
	if exceed <= 0 {
		return
	}

	for i := 0; i < exceed; i++ {
		delete(n.m, n.keys[i])
	}
	n.keys = n.keys[exceed:]
}

var notifiers = []Notifier{}

var notificationIndex = []*gtd.Task{}

var notified = &notifiedList{
	map[int64][]int64{},
	[]int64{},
	60,
}

func init() {
	notifiers = append(notifiers, &stdout{})
}

func notification() {
	rebuildNotificationIndex()

	c := make(chan int)
	tp.HookFunc(func() {
		rebuildNotificationIndex()
		c <- 1
	})

	go func() {
		for {
			if len(notificationIndex) == 0 {
				<-c
				continue
			}

			select {
			// Start 1s before
			case <-time.After(time.Duration((notificationIndex[0].Notification.Get() - time.Now().Unix() - 1) * 1e9)):
				n0 := notificationIndex[0].Notification.Get()

				taskList := []*gtd.Task{}
				for n, i := range notificationIndex {
					if n == 0 || (i.Notification.Get()-n0) < 60 {
						if !notified.Has(i.Notification.Get(), i.ID) {
							taskList = append(taskList, i)
						}
					} else {
						break
					}
				}
				go notify(taskList)

				select {
				case <-time.After(time.Duration((n0 + 59 - time.Now().Unix()) * 1e9)):
					rebuildNotificationIndex()
				case <-c:
				}

			case <-c:
			}
		}
	}()
}

func notify(taskList []*gtd.Task) {
	for _, i := range taskList {
		if tp.Has(i.ID) && !notified.Has(i.Notification.Get(), i.ID) {
			for _, notifier := range notifiers {
				go notifier.Notify(i)
				notified.Add(i.Notification.Get(), i.ID)
			}
		}
	}
}

func rebuildNotificationIndex() {
	tpRW.RLock()
	taskList, _ := tp.FindAll(func(task *gtd.Task) bool {
		if !task.Notification.EqualZero() {
			return true
		}
		return false
	})
	tpRW.RUnlock()

	notificationIndex = notificationIndex[:0]

	now := time.Now().Unix()
	for _, i := range taskList {
		if i.Notification.Get() < now {
			continue
		}
		notificationIndex = append(notificationIndex, i)
	}

	gtd.SortByNotification(notificationIndex)
}

type stdout struct {
}

func (s *stdout) Notify(task *gtd.Task) {
	fmt.Println("Notification: ", time.Unix(task.Notification.Get(), 0).Format(time.RFC1123), task.Subject)
}
