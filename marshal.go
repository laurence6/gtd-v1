package gtd

import (
	"encoding/json"
	"io"
	"strconv"
)

type marshalTask struct {
	Task

	ParentTask int64
	SubTasks   []int64
}

// Marshal serializes Task to a json object and writes to Writer.
func (task *Task) Marshal(w io.Writer) error {
	mt := marshalTask{}
	mt.Task = *task
	if task.ParentTask != nil {
		mt.ParentTask = task.ParentTask.ID
	}
	for _, i := range task.SubTasks {
		mt.SubTasks = append(mt.SubTasks, i.ID)
	}
	mt.Task.ParentTask = nil
	mt.Task.SubTasks = nil

	b, err := json.Marshal(mt)
	if err != nil {
		return err
	}
	w.Write(b)
	w.Write([]byte("\n"))

	for _, i := range task.SubTasks {
		err := i.Marshal(w)
		if err != nil {
			return err
		}
	}
	return nil
}

// Marshal serializes TaskPool to json objects and writes to Writer.
func (tp *TaskPool) Marshal(w io.Writer) error {
	for _, i := range tp.tp {
		if i.ParentTask == nil {
			err := i.Marshal(w)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// UnmarshalTaskPool reads json objects from Reader and deserializes them to a TaskPool.
func UnmarshalTaskPool(r io.Reader) (*TaskPool, error) {
	tp := NewTaskPool()
	decoder := json.NewDecoder(r)
	for {
		mt := marshalTask{}
		if err := decoder.Decode(&mt); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		task := mt.Task
		if mt.ParentTask != 0 {
			task.ParentTask = tp.Get(mt.ParentTask)
			task.ParentTask.AddSubTask(&task)
		}
		tp.tp[task.ID] = &task
	}
	return tp, nil
}

// MarshalJSON marshals Time to json.
func (t *Time) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(t.sec, 10)), nil
}

// UnmarshalJSON unmarshals json to Time.
func (t *Time) UnmarshalJSON(b []byte) error {
	sec, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}
	t.Set(sec)
	return nil
}
