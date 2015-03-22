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
	Id             string   `json:"id"`             //
	Name           string   `json:"team"`           //
	Organization   string   `json:"organization"`   //
	Username       string   `json:"username"`       //
	Description    string   `json:"description"`    //
	Users          []string `json:"users"`          //
	TeamPrivileges []string `json:"teamprivileges"` //
	Repositories   []string `json:"repositories"`   //
	Memo           []string `json:"memo"`           //
}

func (org *Organization) Has(name string) (bool, []byte, error) {
	id, err := GetId("organization", name)
	if err != nil {
		return false, nil, err
	}
	if len(id) <= 0 {
		return false, nil, nil
	}

	err = Get(org, id)

	return true, id, err
}

func (org *Organization) Save() error {
	if err := Save(org, []byte(org.Id)); err != nil {
		return err
	}

	if _, err := LedisDB.HSet([]byte(GLOBAL_ORGANIZATION_INDEX), []byte(org.Name), []byte(org.Id)); err != nil {
		return err
	}

	return nil
}

func (org *Organization) GetById(id string) error {
	if err := Get(org, []byte(id)); err != nil {
		return err
	}

	return nil
}

func (org *Organization) GetByName(name string) error {
	if exists, _, err := org.Has(name); err != nil {
		return err
	} else if exists == false {
    return fmt.Errorf("Orgnization has not found")
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

func (team *Team) Has(org, name string) (bool, []byte, error) {
	id, err := GetId("team", fmt.Sprintf("%s-%s", org, name))
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

	if _, err := LedisDB.HSet([]byte(GLOBAL_TEAM_INDEX), []byte(fmt.Sprintf("%s-%s", team.Organization, team.Name)), []byte(team.Id)); err != nil {
		return err
	}

	return nil
}

func (team *Team) GetById(Id string) error {
	if err := Get(team, []byte(Id)); err != nil {
		return err
	}

	return nil
}

func (team *Team) GetByName(org, name string) error {
  if exists, _, err := team.Has(org, name); err != nil {
    return err
  } else if exists == false {
    return fmt.Errorf("Team has not found")
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
