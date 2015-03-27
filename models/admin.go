package models

import (
	"time"

	"github.com/dockercn/wharf/utils"
)

type Admin struct {
	Id       string   `json:"id"`       //
	Username string   `json:"username"` //
	Password string   `json:"password"` //
	Email    string   `json:"email"`    //
	Created  int64    `json:"created"`  //
	Updated  int64    `json:"updated"`  //
	Memo     []string `json:"memo"`     //
}

func (admin *Admin) Save() error {
	if err := Save(admin, []byte(admin.Id)); err != nil {
		return err
	}

	if _, err := LedisDB.HSet([]byte(GLOBAL_ADMIN_INDEX), []byte(admin.Username), []byte(admin.Id)); err != nil {
		return err
	}

	return nil
}

func (admin *Admin) Log(action, level, t int64, actionId string, content []byte) error {
	log := Log{Action: action, ActionId: actionId, Level: level, Type: t, Content: string(content), Created: time.Now().UnixNano() / int64(time.Millisecond)}
	log.Id = string(utils.GeneralKey(actionId))

	if err := log.Save(); err != nil {
		return err
	}

	admin.Memo = append(admin.Memo, log.Id)

	if err := admin.Save(); err != nil {
		return err
	}

	return nil
}
