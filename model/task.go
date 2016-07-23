package model

import (
	"gopkg.in/pg.v4"
	"gopkg.in/pg.v4/orm"
)

type Task struct {
	TableName struct{} `sql:"task"`

	UserID string

	ID           int64 `sql:",pk"`
	Subject      string
	Due          Time
	Priority     int
	Notification Time
	Next         Time
	Note         string

	Tags Tags

	ParentTaskID int64 `sql:",null"`
	ParentTask   *Task
	SubTasks     []Task `pg:",fk:ParentTask"`
}

func updateTags(tx *pg.Tx, task Task) error {
	if len(task.Tags) == 0 {
		_, err := tx.Exec("DELETE FROM tag WHERE tag.task_id = ?;", task.ID)
		if err != nil {
			return err
		}
		return nil
	}

	tagNames := make([]string, len(task.Tags))
	for n := 0; n < len(task.Tags); n++ {
		task.Tags[n].UserID = task.UserID

		task.Tags[n].TaskID = task.ID

		tagNames[n] = task.Tags[n].Name
	}

	_, err := tx.Exec("DELETE FROM tag WHERE tag.task_id = ? AND tag.name NOT IN (?);", task.ID, pg.In(tagNames))
	if err != nil {
		return err
	}

	_, err = tx.Model(&task.Tags).OnConflict("DO NOTHING").Create()
	if err != nil {
		return err
	}

	return nil
}

func CreateTask(task Task) error {
	tx, err := DBConn.Begin()
	if err != nil {
		return err
	}

	err = tx.Create(&task)
	if err != nil {
		tx.Rollback()
		return err
	}

	if len(task.Tags) > 0 {
		err = updateTags(tx, task)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	return err
}

func GetTask(userID string, taskID int64, columns ...string) (Task, error) {
	task := Task{
		UserID: userID,
		ID:     taskID,
	}

	err := DBConn.Model(&task).
		Column(columns...).
		Where("task.id = ?", taskID).
		Where("task.user_id = ?", userID).
		Select()
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

func GetTasksByUserID(userID string, columns ...string) ([]Task, error) {
	tasks := []Task{}

	err := DBConn.Model(&tasks).
		Column(columns...).
		Where("task.user_id = ?", userID).
		Select()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func GetTasksByTag(userID, tagName string, columns ...string) ([]Task, error) {
	tasks := []Task{}

	err := DBConn.Model(&tasks).
		Column(columns...).
		Join("JOIN tag ON task.id = tag.task_id").
		Where("task.user_id = ?", userID).
		Where("tag.name = ?", tagName).
		Select()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func GetTasksByFunc(userID string, f func(q *orm.Query) *orm.Query, columns ...string) ([]Task, error) {
	tasks := []Task{}

	err := DBConn.Model(&tasks).
		Column(columns...).
		Where("task.user_id = ?", userID).
		Apply(f).
		Select()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func UpdateTask(task Task, columns ...string) error {
	tx, err := DBConn.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Model(&task).
		Column(columns...).
		Where("task.id = ?", task.ID).
		Where("task.user_id = ?", task.UserID).
		Update()
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, column := range columns {
		if column == "Tags" {
			err = updateTags(tx, task)
			if err != nil {
				tx.Rollback()
				return err
			}
			break
		}
	}

	err = tx.Commit()
	return err
}

func DeleteTask(task Task) error {
	_, err := DBConn.Model(&task).
		Where("task.id = ?", task.ID).
		Where("task.user_id = ?", task.UserID).
		Delete()
	return err
}

func DoneTask(task Task) error {
	err := DBConn.Model(&task).
		Where("task.id = ?", task.ID).
		Where("task.user_id = ?", task.UserID).
		Select()
	if err != nil {
		return err
	}

	if task.Next.EqualZero() {
		return DeleteTask(task)
	}

	if task.Due.EqualZero() {
		return nil
	}

	delta := task.Next.Get()/86400*86400 - task.Due.Get()/86400*86400

	task.Due.Set(task.Next.Get())
	if !task.Notification.EqualZero() {
		task.Notification.Set(task.Notification.Get() + delta)
	}
	task.Next.Set(task.Next.Get() + delta)

	return UpdateTask(task, "due", "notification", "next")
}
