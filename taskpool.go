package gtd

import (
	"errors"
	"time"
)

// Errors
var (
	ErrTaskNotFound = errors.New("Cannot find this task")
	ErrDupTaskID    = errors.New("Duplicate task ID")
)

func newTask() *Task {
	now := time.Now()
	id := now.UnixNano() // FIXME: May be duplicate
	start := now.Unix()

	task := &Task{}
	task.ID = id
	task.Start = start
	task.Due = &Time{}
	task.Next = &Time{}
	task.Notification = &Time{}
	return task
}

// TaskPool contains many *Task, using Task.ID as key
type TaskPool map[int64]*Task

// NewTask creates a *Task and stores it into TaskPool
func (tp TaskPool) NewTask() (*Task, error) {
	for n := 0; n < 5; n++ {
		task := newTask()
		if _, ok := tp[task.ID]; !ok {
			tp[task.ID] = task
			return task, nil
		}
	}
	return nil, ErrDupTaskID
}

// NewSubTask creates a sub *Task and stores it into TaskPool
func (tp TaskPool) NewSubTask(task *Task) (*Task, error) {
	if _, ok := tp[task.ID]; !ok {
		return nil, ErrTaskNotFound
	}

	for n := 0; n < 5; n++ {
		subTask := newTask()
		if _, ok := tp[subTask.ID]; !ok {
			task.AddSubTask(subTask)
			subTask.ParentTask = task
			tp[subTask.ID] = subTask
			return subTask, nil
		}
	}
	return nil, ErrDupTaskID
}

// Delete removes a *Task from TaskPool and its parent's SubTasks, then recursively deletes its subtasks
func (tp TaskPool) Delete(task *Task) error {
	if _, ok := tp[task.ID]; !ok {
		return ErrTaskNotFound
	}
	delete(tp, task.ID)
	if task.ParentTask != nil {
		task.ParentTask.DeleteSubTask(task)
		task.ParentTask = nil
	}
	for n := len(task.SubTasks); n > 0; n-- {
		err := tp.Delete(task.SubTasks[0])
		if err != nil {
			return err
		}
	}
	return nil
}

// FindFunc is used by Find to find Task
type FindFunc func(*Task) bool

// Find finds a corresponding Task
func (tp TaskPool) Find(f FindFunc) (*Task, error) {
	for _, i := range tp {
		if f(i) {
			return i, nil
		}
	}
	return nil, ErrTaskNotFound
}

// FindAll finds all corresponding Task
func (tp TaskPool) FindAll(f FindFunc) ([]*Task, error) {
	tasks := []*Task{}
	for _, i := range tp {
		if f(i) {
			tasks = append(tasks, i)
		}
	}
	if len(tasks) > 0 {
		return tasks, nil
	}
	return nil, ErrTaskNotFound
}
