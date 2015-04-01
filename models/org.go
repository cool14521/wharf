package models

import (
	"fmt"
	"time"

	"github.com/containerops/wharf/utils"
)

type Organization struct {
	Id           string   `json:"id"`           //
	Name         string   `json:"name"`         //
	Username     string   `json:"username"`     //
	Description  string   `json:"description"`  //
	Repositories []string `json:"repositories"` //
	Teams        []string `json:"teams"`        //
	Created      int64    `json:"created"`      //
	Updated      int64    `json:"updated"`      //
	Memo         []string `json:"memo"`         //
}

func (org *Organization) Has(name string) (bool, []byte, error) {
	id, err := GetByGobalId("organization", name)
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

func (org *Organization) Remove() error {
	if _, err := LedisDB.HSet([]byte(fmt.Sprintf("%s_remove", GLOBAL_ORGANIZATION_INDEX)), []byte(org.Name), []byte(org.Id)); err != nil {
		return err
	}

	if _, err := LedisDB.HDel([]byte(GLOBAL_ORGANIZATION_INDEX), []byte(org.Id)); err != nil {
		return err
	}

	return nil
}

func (org *Organization) Log(action, level, t int64, actionId string, content []byte) error {
	log := Log{Action: action, ActionId: actionId, Level: level, Type: t, Content: string(content), Created: time.Now().UnixNano() / int64(time.Millisecond)}
	log.Id = string(utils.GeneralKey(actionId))

	if err := log.Save(); err != nil {
		return err
	}

	org.Memo = append(org.Memo, log.Id)

	if err := org.Save(); err != nil {
		return err
	}

	return nil
}
