package main

import (
	"sort"
	"time"

	"github.com/laurence6/gtd.go/model"
)

type index []model.Task

func (index index) Len() int {
	return len(index)
}

func (index index) Swap(i, j int) {
	index[i], index[j] = index[j], index[i]
}

type byDue struct {
	index
}

func (bd byDue) Less(i, j int) bool {
	a := bd.index[i]
	b := bd.index[j]
	switch {
	case a.Due.Get() != b.Due.Get():
		return earlier(a.Due, b.Due)
	case a.Priority != b.Priority:
		return a.Priority < b.Priority
	default:
		return a.ID < b.ID
	}
}

type byNotification struct {
	index
}

func (bn byNotification) Less(i, j int) bool {
	a := bn.index[i]
	b := bn.index[j]
	switch {
	case a.Notification.Get() != b.Notification.Get():
		return earlier(a.Notification, b.Notification)
	case a.Due.Get() != b.Due.Get():
		return earlier(a.Due, b.Due)
	case a.Priority != b.Priority:
		return a.Priority < b.Priority
	default:
		return a.ID < b.ID
	}
}

type byPriority struct {
	index
}

func (bp byPriority) Less(i, j int) bool {
	a := bp.index[i]
	b := bp.index[j]
	switch {
	case a.Priority != b.Priority:
		return a.Priority < b.Priority
	case a.Due.Get() != b.Due.Get():
		return earlier(a.Due, b.Due)
	default:
		return a.ID < b.ID
	}
}

// SortByDue sorts []*Task by Due.
func SortByDue(taskList []model.Task) {
	bd := byDue{taskList}
	sort.Sort(bd)
}

// SortByNotification sorts []*Task by Notification.
func SortByNotification(taskList []model.Task) {
	bn := byNotification{taskList}
	sort.Sort(bn)
}

// SortByDefault sorts []*Task by default algorithm.
// ----------------------------
// 0 |Over due      |By Due
// ----------------------------
// 1 |In one week   |By Due
// ----------------------------
// 2 |Priority == 1 |By Due
// ----------------------------
// 3 |Others        |By Due
// ----------------------------
func SortByDefault(taskList []model.Task) {
	lists := [][]model.Task{[]model.Task{}, []model.Task{}, []model.Task{}}
	now := time.Now().Unix()

	for _, task := range taskList {
		switch {
		case !task.Due.EqualZero() && task.Due.Get()-now < 0:
			fallthrough
		case !task.Due.EqualZero() && task.Due.Get()-now < int64(7*24*time.Hour/1e9):
			lists[0] = append(lists[0], task)
		case task.Priority == 1:
			lists[1] = append(lists[1], task)
		default:
			lists[2] = append(lists[2], task)
		}
	}

	n := 0
	for _, l := range lists {
		SortByDue(l)
		for _, i := range l {
			taskList[n] = i
			n++
		}
	}
}

func earlier(a, b model.Time) bool {
	if a.Get() == 0 || b.Get() == 0 {
		return a.Get() > b.Get()
	}
	return a.Get() < b.Get()
}
