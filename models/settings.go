package models

import (
	"time"

	"github.com/dockercn/wharf/utils"
)

const (
	APIVERSION_V1 = iota
	APIVERSION_V2
)

const (
	TYPE_WEBV1 = iota
	TYPE_WEBV2
	TYPE_APIV1
	TYPE_APIV2
)

const (
	LEVELEMERGENCY = iota
	LevelALERT
	LEVELCRITICAL
	LEVELERROR
	LEVELWARNING
	LEVELNOTICE
	LEVELINFORMATIONAL
	LEVELDEBUG
)

const (
	ACTION_SIGNUP = iota
	ACTION_SIGNIN
	ACTION_SINGOUT
	ACTION_UPDATE_PROFILE
	ACTION_UPDATE_PASSWORD
	ACTION_ADD_REPO
	ACTION_GET_REPO
	ACTION_UPDATE_REPO
	ACTION_PUT_REPO_IMAGES
	ACTION_PUT_TAG
	ACTION_PUT_IMAGES_JSON
	ACTION_PUT_IMAGES_LAYER
	ACTION_PUT_IMAGES_CHECKSUM
	ACTION_REMOVE_REPO
	ACTION_ADD_COMMENT
	ACTION_REMOVE_COMMENT
	ACTION_ADD_ORG
	ACTION_UPDATE_ORG
	ACTION_REMOVE_ORG
	ACTION_ADD_TEAM
	ACTION_REMOVE_TEAM
	ACTION_ADD_PRIVILEGE
	ACTION_REMOVE_PRIVILEGE
	ACTION_ADD_STAR
	ACTION_REMOVE_STAR
)

type Log struct {
	Id       string `json:"id"`       //
	Action   int64  `json:"action"`   //
	ActionId string `json:"actionid"` //
	Level    int64  `json:"level"`    //
	Type     int64  `json:"type"`     //
	Content  string `json:"content"`  //
	Created  int64  `json:"created"`  //
}

type EmailMessage struct {
	Id       string   `json:"id"`       //
	User     string   `json:"user"`     //
	Server   string   `json:"server"`   //
	Template string   `json:"template"` //
	Object   string   `json:"object"`   //
	Message  string   `json:"message"`  //
	Status   string   `json:"status"`   //
	Count    int64    `json:"count"`    //
	Created  int64    `json:"created"`  //
	Updated  int64    `json:"updated"`  //
	Memo     []string `json:"memo"`     //
}

type EmailServer struct {
	Id       string   `json:"id"`      //
	Name     string   `json:"name"`    //
	Host     string   `json:"host"`    //
	Port     int64    `json:"port"`    //
	User     string   `json:"user"`    //
	Password string   `json:"passwd"`  //
	API      string   `json:"api"`     //
	Created  int64    `json:"created"` //
	Updated  int64    `json:"updated"` //
	Memo     []string `json:"memo"`    //
}

type EmailTemplate struct {
	Id      string   `json:"id"`      //
	Server  int64    `json:"server"`  //
	Name    string   `json:"name"`    //
	Content string   `json:"content"` //
	Created int64    `json:"created"` //
	Updated int64    `json:"updated"` //
	Memo    []string `json:"memo"`    //
}

func (l *Log) Has(id string) (bool, []byte, error) {
	if len(id) <= 0 {
		return false, nil, nil
	}

	err := Get(l, []byte(id))

	return true, []byte(id), err
}

func (l *Log) Save() error {
	if err := Save(l, []byte(l.Id)); err != nil {
		return err
	}

	if _, err := LedisDB.HSet([]byte(GLOBAL_LOG_INDEX), []byte(l.Id), []byte(l.Id)); err != nil {
		return err
	}

	return nil
}

func (user *User) Log(action, level, t int64, actionID string, content []byte) error {
	log := Log{Action: action, ActionId: actionID, Level: level, Type: t, Content: string(content), Created: time.Now().UnixNano() / int64(time.Millisecond)}
	log.Id = string(utils.GeneralKey(actionID))

	if err := log.Save(); err != nil {
		return err
	}

	user.Memo = append(user.Memo, log.Id)

	if err := user.Save(); err != nil {
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

func (repo *Repository) Log(action, level, t int64, actionId string, content []byte) error {
	log := Log{Action: action, ActionId: actionId, Level: level, Type: t, Content: string(content), Created: time.Now().UnixNano() / int64(time.Millisecond)}
	log.Id = string(utils.GeneralKey(actionId))

	if err := log.Save(); err != nil {
		return err
	}

	repo.Memo = append(repo.Memo, log.Id)

	if err := repo.Save(); err != nil {
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

func (image *Image) Log(action, level, t int64, actionId string, content []byte) error {
	log := Log{Action: action, ActionId: actionId, Level: level, Type: t, Content: string(content), Created: time.Now().UnixNano() / int64(time.Millisecond)}
	log.Id = string(utils.GeneralKey(actionId))

	if err := log.Save(); err != nil {
		return err
	}

	image.Memo = append(image.Memo, log.Id)

	if err := image.Save(); err != nil {
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

func (p *Privilege) Log(action, level, t int64, actionId string, content []byte) error {
	log := Log{Action: action, ActionId: actionId, Level: level, Type: t, Content: string(content), Created: time.Now().UnixNano() / int64(time.Millisecond)}
	log.Id = string(utils.GeneralKey(actionId))

	if err := log.Save(); err != nil {
		return err
	}

	p.Memo = append(p.Memo, log.Id)

	if err := p.Save(); err != nil {
		return err
	}

	return nil
}
