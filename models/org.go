package models

import (
	"fmt"
)

type Organization struct {
	Id           string   `json:"id"`           //
	Name         string   `json:"organization"` //
	Username     string   `json:"username"`     //
	Description  string   `json:"description"`  //
	Repositories []string `json:"repositories"` //
	Created      int64    `json:"created"`      //
	Updated      int64    `json:"updated"`      //
	Teams        []string `json:"teams"`        //
	Memo         []string `json:"memo"`         //
}

type Team struct {
	Id                string       `json:"id"`             //
	Name              string       `json:"team"`           //
	Organization      string       `json:"organization"`   //
	Username          string       `json:"username"`       //
	Description       string       `json:"description"`    //
	Users             []string     `json:"users"`          //
	TeamPrivileges    []string     `json:"teamprivileges"` //
	Repositories      []string     `json:"repositories"`   //
	Memo              []string     `json:"memo"`           //
	RepositoryObjects []Repository `json:"repositoryobjects"`
	UserObjects       []User       `json:"userobjects"`
}

func (organization *Organization) Has(name string) (bool, []byte, error) {
	id, err := GetId("organization", name)
	if err != nil {
		return false, nil, err
	}
	if len(id) <= 0 {
		return false, nil, nil
	}

	err = Get(organization, id)

	return true, id, err
}

func (organization *Organization) Save() error {
	if err := Save(organization, []byte(organization.Id)); err != nil {
		return err
	}

	if _, err := LedisDB.HSet([]byte(GLOBAL_ORGANIZATION_INDEX), []byte(organization.Name), []byte(organization.Id)); err != nil {
		return err
	}

	return nil
}

func (organization *Organization) Get(id string) error {
	if err := Get(organization, []byte(id)); err != nil {
		return err
	}

	return nil
}

func (organization *Organization) Remove() error {
	if _, err := LedisDB.HSet([]byte(fmt.Sprintf("%s_remove", GLOBAL_ORGANIZATION_INDEX)), []byte(organization.Name), []byte(organization.Id)); err != nil {
		return err
	}

	if _, err := LedisDB.HDel([]byte(GLOBAL_ORGANIZATION_INDEX), []byte(organization.Id)); err != nil {
		return err
	}

	return nil
}

func (team *Team) Has(name string) (bool, []byte, error) {
	id, err := GetId("team", name)
	if err != nil {
		return false, nil, err
	}

	if len(id) <= 0 {
		return false, nil, nil
	}

	err = Get(team, id)

	return true, id, err
}

func (team *Team) Save() error {
	if err := Save(team, []byte(team.Id)); err != nil {
		return err
	}

	if _, err := LedisDB.HSet([]byte(GLOBAL_TEAM_INDEX), []byte(team.Name), []byte(team.Id)); err != nil {
		return err
	}

	return nil
}

func (team *Team) Get(Id string) error {
	if err := Get(team, []byte(Id)); err != nil {
		return err
	}

	return nil
}

func (team *Team) Remove() error {
	if _, err := LedisDB.HSet([]byte(fmt.Sprintf("%s_remove", GLOBAL_TEAM_INDEX)), []byte(team.Name), []byte(team.Id)); err != nil {
		return err
	}

	if _, err := LedisDB.HDel([]byte(GLOBAL_TEAM_INDEX), []byte(team.Id)); err != nil {
		return err
	}

	return nil
}
