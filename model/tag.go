package model

type Tag struct {
	TableName struct{} `sql:"tag"`

	UserID string `sql:",pk"`
	TaskID int64  `sql:",pk"`
	Name   string `sql:",pk"`
}

type Tags []Tag

func (tags Tags) String() string {
	if len(tags) == 0 {
		return ""
	}
	if len(tags) == 1 {
		return tags[0].Name
	}

	l := len(tags) - 1
	for n := 0; n < len(tags); n++ {
		l += len(tags[n].Name)
	}

	b := make([]byte, l)
	bp := copy(b, tags[0].Name)
	for _, tag := range tags[1:] {
		bp += copy(b[bp:], ",")
		bp += copy(b[bp:], tag.Name)
	}
	return string(b)
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
