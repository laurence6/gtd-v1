package model

import (
	"gopkg.in/pg.v4"
	"gopkg.in/pg.v4/orm"
)

// Column flag
const (
	CTaskUserID = 1 << iota

	CTaskID
	CTaskSubject
	CTaskDue
	CTaskPriority
	CTaskReminder
	CTaskNext
	CTaskNote
	CTaskParentTaskID

	CTaskTags

	CTaskSubTaskIDs
)

// Column name string
var (
	ShortColumnNames = map[string]string{
		"UserID":       "user_id",
		"ID":           "id",
		"Subject":      "subject",
		"Due":          "due",
		"Priority":     "priority",
		"Reminder":     "reminder",
		"Next":         "next",
		"Note":         "note",
		"ParentTaskID": "parent_task_id",
	}

	FullColumnNames = map[string]string{
		"UserID":       "task.user_id",
		"ID":           "task.id",
		"Subject":      "task.subject",
		"Due":          "task.due",
		"Priority":     "task.priority",
		"Reminder":     "task.reminder",
		"Next":         "task.next",
		"Note":         "task.note",
		"ParentTaskID": "task.parent_task_id",
	}
)

type taskNormalColumns struct {
	TableName struct{} `sql:"task"`

	UserID string

	ID           int64 `sql:",pk"`
	Subject      string
	Due          Time
	Priority     int
	Reminder     Time
	Next         Time
	Note         string
	ParentTaskID int64 `sql:",null"`
}

type Task struct {
	TableName struct{} `sql:"task"`

	taskNormalColumns

	Tags []string `pg:",array"`

	SubTaskIDs []int64 `sql:"sub_task_ids" pg:",array"`
}

func genNormalColumnList(fullName bool, columns int) []string {
	var columnNames map[string]string
	if fullName {
		columnNames = FullColumnNames
	} else {
		columnNames = ShortColumnNames
	}

	c := make([]string, 0, len(columnNames)) // TODO: pop count

	if columns&CTaskUserID != 0 {
		c = append(c, columnNames["UserID"])
	}
	if columns&CTaskID != 0 {
		c = append(c, columnNames["ID"])
	}
	if columns&CTaskSubject != 0 {
		c = append(c, columnNames["Subject"])
	}
	if columns&CTaskDue != 0 {
		c = append(c, columnNames["Due"])
	}
	if columns&CTaskPriority != 0 {
		c = append(c, columnNames["Priority"])
	}
	if columns&CTaskReminder != 0 {
		c = append(c, columnNames["Reminder"])
	}
	if columns&CTaskNext != 0 {
		c = append(c, columnNames["Next"])
	}
	if columns&CTaskNote != 0 {
		c = append(c, columnNames["Note"])
	}
	if columns&CTaskParentTaskID != 0 {
		c = append(c, columnNames["ParentTaskID"])
	}

	return c
}

func joinTags(q *orm.Query) *orm.Query {
	return q.Column("task_tag.tags").
		Join("LEFT JOIN task_tag ON task.id = task_tag.task_id")
}

func updateTags(tx *pg.Tx, task Task) error {
	if len(task.Tags) == 0 {
		_, err := tx.Exec("DELETE FROM tag WHERE tag.task_id = ?", task.ID)
		if err != nil {
			return err
		}
		return nil
	}

	_, err := tx.Exec("DELETE FROM tag WHERE tag.task_id = ? AND tag.name NOT IN (?)", task.ID, pg.In(task.Tags))
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO tag (SELECT ?, ?, unnest(?::text[])) ON CONFLICT DO NOTHING", task.UserID, task.ID, pg.Array(task.Tags))
	if err != nil {
		return err
	}

	return nil
}

func joinSubTaskIDs(q *orm.Query) *orm.Query {
	return q.Column("task_sub_task_id.sub_task_ids").
		Join("LEFT JOIN task_sub_task_id ON task.id = task_sub_task_id.task_id")
}

func CreateTask(task Task) error {
	tx, err := DBConn.Begin()
	if err != nil {
		return err
	}

	err = tx.Create(&task.taskNormalColumns)
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

func GetTask(userID string, taskID int64, columns int) (Task, error) {
	task := Task{}

	q := DBConn.Model(&task).Column("task.*")

	if columns&CTaskTags != 0 {
		q = joinTags(q)
	}
	if columns&CTaskSubTaskIDs != 0 {
		q = joinSubTaskIDs(q)
	}

	err := q.
		Where("task.id = ? AND task.user_id = ?", taskID, userID).
		Select()
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

func GetTasksByID(userID string, ids []int64, columns int) ([]Task, error) {
	tasks := []Task{}

	q := DBConn.Model(&tasks).Column("task.*")

	if columns&CTaskTags != 0 {
		q = joinTags(q)
	}
	if columns&CTaskSubTaskIDs != 0 {
		q = joinSubTaskIDs(q)
	}

	err := q.
		Where("task.user_id = ? AND task.id IN (?)", userID, pg.In(ids)).
		Select()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func GetTasksByUserID(userID string, columns int) ([]Task, error) {
	tasks := []Task{}

	q := DBConn.Model(&tasks).Column("task.*")

	if columns&CTaskTags != 0 {
		q = joinTags(q)
	}
	if columns&CTaskSubTaskIDs != 0 {
		q = joinSubTaskIDs(q)
	}

	err := q.
		Where("task.user_id = ?", userID).
		Select()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func GetTasksByTag(userID string, tagName string, columns int) ([]Task, error) {
	tasks := []Task{}

	q := DBConn.Model(&tasks).Column("task.*")

	if columns&CTaskTags != 0 {
		q = joinTags(q)
	}
	if columns&CTaskSubTaskIDs != 0 {
		q = joinSubTaskIDs(q)
	}

	err := q.
		Join("INNER JOIN tag ON task.id = tag.task_id").
		Where("task.user_id = ? AND tag.name = ?", userID, tagName).
		Select()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func GetTasksByFunc(userID string, f func(q *orm.Query) *orm.Query, columns int) ([]Task, error) {
	tasks := []Task{}

	q := DBConn.Model(&tasks).Column("task.*")

	if columns&CTaskTags != 0 {
		q = joinTags(q)
	}
	if columns&CTaskSubTaskIDs != 0 {
		q = joinSubTaskIDs(q)
	}

	err := q.
		Where("task.user_id = ?", userID).
		Apply(f).
		Select()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func UpdateTask(task Task, columns int) error {
	tx, err := DBConn.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Model(&task).
		Column(genNormalColumnList(false, columns)...).
		Where("task.id = ? AND task.user_id = ?", task.ID, task.UserID).
		Update()
	if err != nil {
		tx.Rollback()
		return err
	}

	if columns&CTaskTags != 0 {
		err = updateTags(tx, task)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	return err
}

func DeleteTask(task Task) error {
	_, err := DBConn.Model(&task).
		Where("task.id = ? AND task.user_id = ?", task.ID, task.UserID).
		Delete()
	return err
}

func DoneTask(task Task) error {
	err := DBConn.Model(&task).
		Where("task.id = ? AND task.user_id = ?", task.ID, task.UserID).
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
	if !task.Reminder.EqualZero() {
		task.Reminder.Set(task.Reminder.Get() + delta)
	}
	task.Next.Set(task.Next.Get() + delta)

	return UpdateTask(task, CTaskDue|CTaskReminder|CTaskNext)
}
