package gtd

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
)

// Marshal serializes TaskPool to json objects and writes to Writer.
func (tp *TaskPool) Marshal(w io.Writer) error {
	for _, i := range tp.tp {
		if i.ParentTask == nil {
			err := i.marshalJSON(w)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Unmarshal reads json objects from Reader and deserializes them to a TaskPool.
func (tp *TaskPool) Unmarshal(b []byte) error {
	r := bytes.NewReader(b)
	decoder := json.NewDecoder(r)
	for {
		mt := marshalTask{}

		err := decoder.Decode(&mt)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		task := mt.Task
		if mt.ParentTask != 0 {
			task.ParentTask = tp.Get(mt.ParentTask)
			task.ParentTask.AddSubTask(&task)
		}

		tp.tp[task.ID] = &task
	}
	return nil
}

type marshalTask struct {
	Task

	ParentTask int64
	SubTasks   []int64
}

func (task *Task) marshalJSON(w io.Writer) error {
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

	for _, i := range task.SubTasks {
		err := i.marshalJSON(w)
		if err != nil {
			return err
		}
	}

	return nil
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
