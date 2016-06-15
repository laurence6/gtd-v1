package gtd

import (
	"sort"
	"time"
)

type index []*Task

func (index index) Len() int {
	return len(index)
}

func (index index) Swap(i, j int) {
	index[i], index[j] = index[j], index[i]
}

func earlier(a, b *Time) bool {
	if a.sec == 0 || b.sec == 0 {
		return a.sec > b.sec
	}
	return a.sec < b.sec
}

type byDue struct {
	index
}

func (bd byDue) Less(i, j int) bool {
	a := bd.index[i]
	b := bd.index[j]
	switch {
	case a.Due.sec != b.Due.sec:
		return earlier(a.Due, b.Due)
	case a.Priority != b.Priority:
		return a.Priority < b.Priority
	case a.Start != b.Start:
		return a.Start < b.Start
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
	case a.Notification.sec != b.Notification.sec:
		return earlier(a.Notification, b.Notification)
	case a.Due.sec != b.Due.sec:
		return earlier(a.Due, b.Due)
	case a.Priority != b.Priority:
		return a.Priority < b.Priority
	case a.Start != b.Start:
		return a.Start < b.Start
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
	case a.Due.sec != b.Due.sec:
		return earlier(a.Due, b.Due)
	case a.Start != b.Start:
		return a.Start < b.Start
	default:
		return a.ID < b.ID
	}
}

// SortByDue sorts []*Task by Due.
func SortByDue(taskList []*Task) {
	bd := byDue{taskList}
	sort.Sort(bd)
}

// SortByNotification sorts []*Task by Notification.
func SortByNotification(taskList []*Task) {
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
func SortByDefault(taskList []*Task) {
	lists := [][]*Task{[]*Task{}, []*Task{}, []*Task{}}
	now := time.Now().Unix()

	for _, i := range taskList {
		switch {
		case !i.Due.EqualZero() && i.Due.sec-now < 0:
			fallthrough
		case !i.Due.EqualZero() && i.Due.sec-now < int64(7*24*time.Hour/1e9):
			lists[0] = append(lists[0], i)
		case i.Priority == 1:
			lists[1] = append(lists[1], i)
		default:
			lists[2] = append(lists[2], i)
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
