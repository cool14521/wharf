package models

type Admin struct {
	UUID     string   `json:"UUID"`     //
	Username string   `json:"username"` //
	Password string   `json:"password"` //
	Email    string   `json:"email"`    //
	Created  int64    `json:"created"`  //
	Updated  int64    `json:"updated"`  //
	Memo     []string `json:"memo"`     //
}

func (admin *Admin) Save() error {
	if err := Save(admin, []byte(admin.UUID)); err != nil {
		return err
	}

	if _, err := LedisDB.HSet([]byte(GLOBAL_ADMIN_INDEX), []byte(admin.Username), []byte(admin.UUID)); err != nil {
		return err
	}

	return nil
}
