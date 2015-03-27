package models

import (
	"time"

	"github.com/dockercn/wharf/utils"
)

type Comment struct {
	Id      string   `json:"id"`      //
	Comment string   `json:"comment"` //
	User    string   `json:"user"`    //
	Object  string   `json:"object"`  //
	Created int64    `json:"created"` //
	Updated int64    `json:"updated"` //
	Memo    []string `json:"memo"`    //
}

func (comment *Comment) Save() error {
	if err := Save(comment, []byte(comment.Id)); err != nil {
		return err
	}

	return nil
}

func (comment *Comment) Log(action, level, t int64, actionId string, content []byte) error {
	log := Log{Action: action, ActionId: actionId, Level: level, Type: t, Content: string(content), Created: time.Now().UnixNano() / int64(time.Millisecond)}
	log.Id = string(utils.GeneralKey(actionId))

	if err := log.Save(); err != nil {
		return err
	}

	comment.Memo = append(comment.Memo, log.Id)

	if err := comment.Save(); err != nil {
		return err
	}

	return nil
}
