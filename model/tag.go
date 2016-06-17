package model

type Tag struct {
	TableName struct{} `sql:"tag"`

	UserID string `sql:",pk"`
	TaskID int64  `sql:",pk"`
	Name   string `sql:",pk"`
}

func GetTagsByUserID(userID string) ([]Tag, error) {
	tags := []Tag{}

	err := DBConn.Model(&tags).
		Column("name").
		Where("tag.user_id = ?", userID).
		Group("name").
		Select()
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func GetTagsByTaskID(taskID int64) ([]Tag, error) {
	tags := []Tag{}

	err := DBConn.Model(&tags).
		Where("tag.task_id = ?", taskID).
		Select()
	if err != nil {
		return nil, err
	}

	return tags, nil
}
