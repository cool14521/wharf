package models

type Object interface {
	Create() error
	Read() (Object, error)
	Update() error
	Delete() error
}

type Model interface {
	Has(name string) (bool, string, error)
	GetById(id string) error
	GetByName(name string) error
	Log(action, actionLevel, actionType int64, actionId string, content []string) error
}
