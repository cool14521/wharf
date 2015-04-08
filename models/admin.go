package models

type Admin interface {
	CreateAdmin(username, password, email string) error
}
