package models

type Object interface {
	//Basic CRUD Method
	Create() error
	Read() error
	Update() error
	Delete() error
	//Basic Object Method
	Has(name string) (bool, string, error)
	GetById(id string) error
	GetByName(name string) error
	Log(action, actionLevel, actionType int64, actionId string, content []string) error
}
