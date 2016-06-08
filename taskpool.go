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
	id := now.UnixNano() // May be duplicate
	start := now.Unix()

	task := &Task{}
	task.ID = id
	task.Start = start
	task.Due = &Time{}
	task.Notification = &Time{}
	task.Next = &Time{}
	return task
}

// TaskPool contains many *Task, using Task.ID as key. It is not thread safe.
type TaskPool struct {
	tp    map[int64]*Task
	hooks []func()
}

// NewTaskPool returns a *TaskPool.
func NewTaskPool() *TaskPool {
	return &TaskPool{
		map[int64]*Task{},
		[]func(){},
	}
}

// NewTask creates a *Task and stores it into TaskPool.
func (tp *TaskPool) NewTask() (*Task, error) {
	for n := 0; n < 3; n++ {
		task := newTask()
		if !tp.Has(task.ID) {
			tp.tp[task.ID] = task
			return task, nil
		}
	}
	return nil, ErrDupTaskID
}

// NewSubTask creates a sub *Task and stores it into TaskPool.
func (tp *TaskPool) NewSubTask(task *Task) (*Task, error) {
	if !tp.Has(task.ID) {
		return nil, ErrTaskNotFound
	}

	for n := 0; n < 3; n++ {
		subTask := newTask()
		if !tp.Has(subTask.ID) {
			task.AddSubTask(subTask)
			subTask.ParentTask = task
			tp.tp[subTask.ID] = subTask
			return subTask, nil
		}
	}
	return nil, ErrDupTaskID
}

// Get returns a *Task using id as key.
func (tp *TaskPool) Get(id int64) *Task {
	if !tp.Has(id) {
		return nil
	}
	return tp.tp[id]
}

// GetAll returns all *Task in TaskPool.
func (tp *TaskPool) GetAll() []*Task {
	taskList := []*Task{}
	for _, i := range tp.tp {
		taskList = append(taskList, i)
	}
	return taskList
}

// Has returns if TaskPool has this id.
func (tp *TaskPool) Has(id int64) bool {
	if _, ok := tp.tp[id]; ok {
		return true
	}
	return false
}

// Delete removes a *Task from TaskPool and its parent's SubTasks, then recursively deletes its subtasks.
func (tp *TaskPool) Delete(task *Task) error {
	if !tp.Has(task.ID) {
		return ErrTaskNotFound
	}
	for n := len(task.SubTasks); n > 0; n-- {
		err := tp.Delete(task.SubTasks[0])
		if err != nil {
			return err
		}
	}
	delete(tp.tp, task.ID)
	if task.ParentTask != nil {
		task.ParentTask.DeleteSubTask(task)
		task.ParentTask = nil
	}
	return nil
}

// Done deletes the task if task.Next == 0, or it add Due, Notification, Next by Next-Start and set Start = Next.
func (tp *TaskPool) Done(task *Task) error {
	if !tp.Has(task.ID) {
		return ErrTaskNotFound
	}
	if task.Next.EqualZero() {
		return tp.Delete(task)
	}
	for n := len(task.SubTasks); n > 0; n-- {
		err := tp.Done(task.SubTasks[0])
		if err != nil {
			return err
		}
	}
	delta := task.Next.Get()/86400*86400 - task.Start/86400*86400
	task.Start = task.Next.Get()
	if !task.Due.EqualZero() {
		task.Due.Set(task.Due.Get() + delta)
	}
	if !task.Notification.EqualZero() {
		task.Notification.Set(task.Notification.Get() + delta)
	}
	task.Next.Set(task.Next.Get() + delta)
	return nil
}

// FindFunc is used by Find to find Task.
type FindFunc func(*Task) bool

// Find finds a corresponding Task.
func (tp *TaskPool) Find(f FindFunc) (*Task, error) {
	for _, i := range tp.tp {
		if f(i) {
			return i, nil
		}
	}
	return nil, ErrTaskNotFound
}

// FindAll finds all corresponding Task.
func (tp *TaskPool) FindAll(f FindFunc) ([]*Task, error) {
	tasks := []*Task{}
	for _, i := range tp.tp {
		if f(i) {
			tasks = append(tasks, i)
		}
	}
	if len(tasks) > 0 {
		return tasks, nil
	}
	return nil, ErrTaskNotFound
}

// HookFunc adds hook function.
func (tp *TaskPool) HookFunc(f func()) {
	tp.hooks = append(tp.hooks, f)
}

// Changed is called when TaskPool or Task in it is changed. It calls hook functions one by one in another go routine.
func (tp *TaskPool) Changed() {
	go func() {
		for _, f := range tp.hooks {
			f()
		}
	}()
}
