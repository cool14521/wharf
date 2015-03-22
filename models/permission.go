package models

import (
	"fmt"
)

type Permission struct {
	Id           string   `json:"id"`           //
	Write        bool     `json:"write"`        //
	User         string   `json:"user"`         //
	Organization string   `json:"organization"` //
	Team         string   `json:"team"`         //
	Object       string   `json:"object"`       //
	Memo         []string `json:"memo"`         //
}

func (p *Permission) Get(id string) error {
	if err := Get(p, []byte(id)); err != nil {
		return err
	}

	return nil
}

func (p *Permission) Save() error {
	if err := Save(p, []byte(p.Id)); err != nil {
		return err
	}

	if _, err := LedisDB.HSet([]byte(GLOBAL_PRIVILEGE_INDEX), []byte(fmt.Sprintf("%s:%s:%s", p.Write, p.Team, p.Object)), []byte(p.Id)); err != nil {
		return err
	}

	return nil
}
