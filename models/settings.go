package models

import (
	"time"

	"github.com/dockercn/wharf/utils"
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
	ACTION_ADD_REPO
	ACTION_UPDATE_REPO
	ACTION_DEL_REPO
	ACTION_ADD_COMMENT
	ACTION_DEL_COMMENT
	ACTION_ADD_ORG
	ACTION_DEL_ORG
	ACTION_ADD_MEMBER
	ACTION_DEL_MEMBER
	ACTION_ADD_STAR
	ACTION_DEL_STAR
)

type Log struct {
	UUID       string `json:"UUID"`       //
	Action     int64  `json:"action"`     //
	ActionUUID string `json:"actionuuid"` //
	Level      int64  `json:"level"`      //
	Content    string `json:"content"`    //
	Created    int64  `json:"created"`    //
}

type EmailMessage struct {
	UUID     string   `json:"UUID"`     //
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
	UUID     string   `json:"UUID"`    //
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
	UUID    string   `json:"UUID"`    //
	Server  int64    `json:"server"`  //
	Name    string   `json:"name"`    //
	Content string   `json:"content"` //
	Created int64    `json:"created"` //
	Updated int64    `json:"updated"` //
	Memo    []string `json:"memo"`    //
}

func (l *Log) Has(uuid string) (bool, []byte, error) {
	if len(uuid) <= 0 {
		return false, nil, nil
	}

	err := Get(l, []byte(uuid))

	return true, []byte(uuid), err
}

func (l *Log) Save() error {
	if err := Save(l, []byte(l.UUID)); err != nil {
		return err
	}

	if _, err := LedisDB.HSet([]byte(GLOBAL_LOG_INDEX), []byte(l.UUID), []byte(l.UUID)); err != nil {
		return err
	}

	return nil
}

func (user *User) Log(action, level int64, actionUUID string, content []byte) error {
	log := Log{Action: action, ActionUUID: actionUUID, Level: level, Content: string(content), Created: time.Now().Unix()}
	log.UUID = string(utils.GeneralKey(actionUUID))

	if err := log.Save(); err != nil {
		return err
	}

	user.Memo = append(user.Memo, log.UUID)

	if err := user.Save(); err != nil {
		return err
	}

	return nil
}

func (admin *Admin) Log(action, level int64, actionUUID string, content []byte) error {
	log := Log{Action: action, ActionUUID: actionUUID, Level: level, Content: string(content), Created: time.Now().Unix()}
	log.UUID = string(utils.GeneralKey(actionUUID))

	if err := log.Save(); err != nil {
		return err
	}

	admin.Memo = append(admin.Memo, log.UUID)

	if err := admin.Save(); err != nil {
		return err
	}

	return nil
}

func (org *Organization) Log(action, level int64, actionUUID string, content []byte) error {
	log := Log{Action: action, ActionUUID: actionUUID, Level: level, Content: string(content), Created: time.Now().Unix()}
	log.UUID = string(utils.GeneralKey(actionUUID))

	if err := log.Save(); err != nil {
		return err
	}

	org.Memo = append(org.Memo, log.UUID)

	if err := org.Save(); err != nil {
		return err
	}

	return nil
}

func (team *Team) Log(action, level int64, actionUUID string, content []byte) error {
	log := Log{Action: action, ActionUUID: actionUUID, Level: level, Content: string(content), Created: time.Now().Unix()}
	log.UUID = string(utils.GeneralKey(actionUUID))

	if err := log.Save(); err != nil {
		return err
	}

	team.Memo = append(team.Memo, log.UUID)

	if err := team.Save(); err != nil {
		return err
	}

	return nil
}

func (repo *Repository) Log(action, level int64, actionUUID string, content []byte) error {
	log := Log{Action: action, ActionUUID: actionUUID, Level: level, Content: string(content), Created: time.Now().Unix()}
	log.UUID = string(utils.GeneralKey(actionUUID))

	if err := log.Save(); err != nil {
		return err
	}

	repo.Memo = append(repo.Memo, log.UUID)

	if err := repo.Save(); err != nil {
		return err
	}

	return nil
}

func (compose *Compose) Log(action, level int64, actionUUID string, content []byte) error {
	log := Log{Action: action, ActionUUID: actionUUID, Level: level, Content: string(content), Created: time.Now().Unix()}
	log.UUID = string(utils.GeneralKey(actionUUID))

	if err := log.Save(); err != nil {
		return err
	}

	compose.Memo = append(compose.Memo, log.UUID)

	if err := compose.Save(); err != nil {
		return err
	}

	return nil
}

func (image *Image) Log(action, level int64, actionUUID string, content []byte) error {
	log := Log{Action: action, ActionUUID: actionUUID, Level: level, Content: string(content), Created: time.Now().Unix()}
	log.UUID = string(utils.GeneralKey(actionUUID))

	if err := log.Save(); err != nil {
		return err
	}

	image.Memo = append(image.Memo, log.UUID)

	if err := image.Save(); err != nil {
		return err
	}

	return nil
}

func (star *Star) Log(action, level int64, actionUUID string, content []byte) error {
	log := Log{Action: action, ActionUUID: actionUUID, Level: level, Content: string(content), Created: time.Now().Unix()}
	log.UUID = string(utils.GeneralKey(actionUUID))

	if err := log.Save(); err != nil {
		return err
	}

	star.Memo = append(star.Memo, log.UUID)

	if err := star.Save(); err != nil {
		return err
	}

	return nil
}

func (comment *Comment) Log(action, level int64, actionUUID string, content []byte) error {
	log := Log{Action: action, ActionUUID: actionUUID, Level: level, Content: string(content), Created: time.Now().Unix()}
	log.UUID = string(utils.GeneralKey(actionUUID))

	if err := log.Save(); err != nil {
		return err
	}

	comment.Memo = append(comment.Memo, log.UUID)

	if err := comment.Save(); err != nil {
		return err
	}

	return nil
}

func (p *Privilege) Log(action, level int64, actionUUID string, content []byte) error {
	log := Log{Action: action, ActionUUID: actionUUID, Level: level, Content: string(content), Created: time.Now().Unix()}
	log.UUID = string(utils.GeneralKey(actionUUID))

	if err := log.Save(); err != nil {
		return err
	}

	p.Memo = append(p.Memo, log.UUID)

	if err := p.Save(); err != nil {
		return err
	}

	return nil
}
