package model

func GetTagsByUserID(userID string) ([]string, error) {
	tags := []string{}

	_, err := DBConn.Query(&tags, "SELECT tag.name FROM tag WHERE tag.user_id = ? GROUP BY tag.name", userID)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func GetTagsByTaskID(taskID int64) ([]string, error) {
	tags := []string{}

	_, err := DBConn.Query(&tags, "SELECT tag.name FROM tag WHERE tag.task_id = ?", taskID)
	if err != nil {
		return nil, err
	}

	return tags, nil
}
