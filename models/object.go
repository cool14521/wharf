package models

type Object interface {
	Create() error
	Read() error
	Update() error
	Delete() error
}

type Model interface {
	Object
	Has(name string) (bool, string, error)
	GetById(id string) error
	GetByName(name string) error
	Log(action, actionLevel, actionType int64, actionId string, content []string) error
}
