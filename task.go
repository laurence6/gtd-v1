package gtd

// Task represents a task.
type Task struct {
	ID    int64
	Start int64

	Subject      string
	Due          *Time
	Priority     int
	Notification *Time
	Next         *Time
	Note         string

	ParentTask *Task
	SubTasks   []*Task
}

// AddSubTask appends *Task to SubTasks.
func (task *Task) AddSubTask(subTask *Task) {
	task.SubTasks = append(task.SubTasks, subTask)
}

// DeleteSubTask deletes *Task from SubTasks.
func (task *Task) DeleteSubTask(subTask *Task) {
	for n, i := range task.SubTasks {
		if i == subTask {
			task.SubTasks = append(task.SubTasks[:n], task.SubTasks[n+1:]...)
			break
		}
	}
}
