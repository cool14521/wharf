package models

type Admin struct {
	Object
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (a *Admin) Has(name string) (bool, string, error) {
	return false, "", nil
}

func (a *Admin) GetById(id string) error {
	return nil
}

func (a *Admin) GetByName(name string) error {
	return nil
}

func (a *Admin) Log(action, actionLevel, actionType int64, actionId string, content []string) error {
	return nil
}

func (a *Admin) CreateAdmin(username, password, email string) error {

}
