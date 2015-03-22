package models

import (
	"fmt"
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
