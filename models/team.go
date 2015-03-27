package models

import (
	"fmt"
	"time"

	"github.com/dockercn/wharf/utils"
)

type Team struct {
	Id           string   `json:"id"`           //
	Name         string   `json:"name"`         //
	Organization string   `json:"organization"` //
	Username     string   `json:"username"`     //
	Description  string   `json:"description"`  //
	Users        []string `json:"users"`        //
	Permissions  []string `json:"permissions"`  //
	Repositories []string `json:"repositories"` //
	Memo         []string `json:"memo"`         //
}

func (team *Team) Has(org, name string) (bool, []byte, error) {
	id, err := GetByGobalId("team", fmt.Sprintf("%s-%s", org, name))
	if err != nil {
		return false, nil, err
	}

	if len(id) <= 0 {
		return false, nil, nil
	}

	err = Get(team, id)

	return true, id, err
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

func (team *Team) Save() error {
	if err := Save(team, []byte(team.Id)); err != nil {
		return err
	}

	if _, err := LedisDB.HSet([]byte(GLOBAL_TEAM_INDEX), []byte(fmt.Sprintf("%s-%s", team.Organization, team.Name)), []byte(team.Id)); err != nil {
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

func (team *Team) Log(action, level, t int64, actionId string, content []byte) error {
	log := Log{Action: action, ActionId: actionId, Level: level, Type: t, Content: string(content), Created: time.Now().UnixNano() / int64(time.Millisecond)}
	log.Id = string(utils.GeneralKey(actionId))

	if err := log.Save(); err != nil {
		return err
	}

	team.Memo = append(team.Memo, log.Id)

	if err := team.Save(); err != nil {
		return err
	}

	return nil
}
