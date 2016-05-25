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

func earlier(a, b int64) bool {
	if a == 0 || b == 0 {
		return a > b
	}
	return a < b
}

type byDue struct {
	index
}

func (bd byDue) Less(i, j int) bool {
	a := bd.index[i].Due
	b := bd.index[j].Due
	if a == b {
		return bd.index[i].Priority < bd.index[j].Priority
	}
	return earlier(a, b)
}

type byPriority struct {
	index
}

func (bp byPriority) Less(i, j int) bool {
	a := bp.index[i].Priority
	b := bp.index[j].Priority
	if a == b {
		return earlier(bp.index[i].Due, bp.index[j].Due)
	}
	return a < b
}

// SortByDefault sorts []*Task by default algorithm
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
		case i.Due-now < 0:
			fallthrough
		case i.Due-now < int64(7*24*time.Hour/1e9):
			lists[0] = append(lists[0], i)
		case i.Priority == 1:
			lists[1] = append(lists[1], i)
		default:
			lists[2] = append(lists[2], i)
		}
	}
	n := 0
	for _, l := range lists {
		bd := byDue{l}
		sort.Sort(bd)
		for _, i := range bd.index {
			taskList[n] = i
			n++
		}
	}
}
