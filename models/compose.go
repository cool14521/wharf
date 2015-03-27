package models

import (
	"fmt"
	"time"

	"github.com/dockercn/wharf/utils"
)

type Compose struct {
	Id            string   `json:"id"`            //
	Compose       string   `json:"compose"`       //
	Namespace     string   `json:"namespace"`     //
	NamespaceType bool     `json:"namespacetype"` //
	Organization  string   `json:"organization"`  //
	Tags          []string `json:"tags"`          //
	Starts        []string `json:"starts"`        //
	Comments      []string `json:"comments"`      //
	Short         string   `json:"short"`         //
	Description   string   `json:"description"`   //
	YAML          string   `json:"yaml"`          //
	Download      int64    `json:"download"`      //
	Icon          string   `json:"icon"`          //
	Privated      bool     `json:"privated"`      //
	Permissions   []string `json:"permissions"`   //
	Created       int64    `json:"created"`       //
	Updated       int64    `json:"updated"`       //
	Memo          []string `json:"memo"`          //
}

func (c *Compose) Has(namespace, compose string) (bool, []byte, error) {
	id, err := GetByGobalId("compose", fmt.Sprintf("%s:%s", namespace, compose))

	if err != nil {
		return false, nil, err
	}

	if len(id) <= 0 {
		return false, nil, nil
	}

	err = Get(c, id)

	return true, id, err
}

func (c *Compose) Save() error {
	if err := Save(c, []byte(c.Id)); err != nil {
		return err
	}

	if _, err := LedisDB.HSet([]byte(GLOBAL_COMPOSE_INDEX), []byte(fmt.Sprintf("%s:%s", c.Namespace, c.Compose)), []byte(c.Id)); err != nil {
		return err
	}

	return nil
}

func (compose *Compose) Log(action, level, t int64, actionId string, content []byte) error {
	log := Log{Action: action, ActionId: actionId, Level: level, Type: t, Content: string(content), Created: time.Now().UnixNano() / int64(time.Millisecond)}
	log.Id = string(utils.GeneralKey(actionId))

	if err := log.Save(); err != nil {
		return err
	}

	compose.Memo = append(compose.Memo, log.Id)

	if err := compose.Save(); err != nil {
		return err
	}

	return nil
}
