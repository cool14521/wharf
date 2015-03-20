package models

type Admin struct {
	Id       string   `json:"id"`       //
	Username string   `json:"username"` //
	Password string   `json:"password"` //
	Email    string   `json:"email"`    //
	Created  int64    `json:"created"`  //
	Updated  int64    `json:"updated"`  //
	Memo     []string `json:"memo"`     //
}

func (admin *Admin) Save() error {
	if err := Save(admin, []byte(admin.Id)); err != nil {
		return err
	}

	if _, err := LedisDB.HSet([]byte(GLOBAL_ADMIN_INDEX), []byte(admin.Username), []byte(admin.Id)); err != nil {
		return err
	}

	return nil
}
