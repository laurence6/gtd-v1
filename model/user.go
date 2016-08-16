package model

type User struct {
	ID       string `sql:",pk"`
	Password string
}

func CreateUser(userID string, password string) error {
	user := User{
		ID:       userID,
		Password: password,
	}

	err := DBConn.Create(&user)
	if err != nil {
		return err
	}

	return nil
}

func GetUser(userID string) (User, error) {
	user := User{
		ID: userID,
	}

	err := DBConn.Select(&user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func ChangeUserID(oldUserID, newUserID string) error {
	user := User{
		ID: newUserID,
	}

	_, err := DBConn.Model(&user).
		Column("id").
		Where("user.id = ?", oldUserID).
		Update()
	return err
}

func ChangePassword(userID string, password string) error {
	user := User{
		ID:       userID,
		Password: password,
	}

	_, err := DBConn.Model(&user).
		Column("password").
		Update()
	return err
}

func DeleteUser(userID string) error {
	user := User{
		ID: userID,
	}

	err := DBConn.Delete(&user)
	return err
}
