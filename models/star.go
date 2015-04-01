package models

import (
	"time"

	"github.com/containerops/wharf/utils"
)

type Star struct {
	Id     string   `json:"id"`     //
	User   string   `json:"user"`   //
	Object string   `json:"object"` //
	Time   int64    `json:"time"`   //
	Memo   []string `json:"memo"`   //
}

func (star *Star) Save() error {
	if err := Save(star, []byte(star.Id)); err != nil {
		return err
	}

	return nil
}

func (star *Star) Log(action, level, t int64, actionId string, content []byte) error {
	log := Log{Action: action, ActionId: actionId, Level: level, Type: t, Content: string(content), Created: time.Now().UnixNano() / int64(time.Millisecond)}
	log.Id = string(utils.GeneralKey(actionId))

	if err := log.Save(); err != nil {
		return err
	}

	star.Memo = append(star.Memo, log.Id)

	if err := star.Save(); err != nil {
		return err
	}

	return nil
}
