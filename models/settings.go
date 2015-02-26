package models

import (
	"fmt"
	"time"
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
	Content    []byte `json:"content"`    //
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
	UUID    string   `json:"UUID"`    //
	Name    string   `json:"name"`    //
	Host    string   `json:"host"`    //
	Port    int64    `json:"port"`    //
	User    string   `json:"user"`    //
	Passwd  string   `json:"passwd"`  //
	API     string   `json:"api"`     //
	Created int64    `json:"created"` //
	Updated int64    `json:"updated"` //
	Memo    []string `json:"memo"`    //
}

type EmailTemplate struct {
	UUID    string    `json:"UUID"`    //
	Server  int64     `json:"server"`  //
	Name    string    `json:"name"`    //
	Content string    `json:"content"` //
	Created time.Time `json:"created"` //
	Updated time.Time `json:"updated"` //
	Memo    []string  `json:"memo"`    //
}

func (u *User) Log(action, level int64, actionId string, content []byte) error {
	if uuid, err := GetUUID("log", fmt.Sprintf("%d-%d-%d", action, actionId)); err != nil {
		return err
	} else {
		log := Log{UUID: string(uuid), Created: time.Now().Unix()}
		log.Action = action
		log.ActionUUID = actionId

		if err := Save(log, []byte(uuid)); err != nil {
			return err
		}

		u.Memo = append(u.Memo, string(uuid))

		if err := u.Save(); err != nil {
			return err
		}
	}

	return nil
}

func (o *Organization) Log(action, level int64, actionId string, content []byte) error {
	if uuid, err := GetUUID("log", fmt.Sprintf("%d-%d-%d", action, actionId)); err != nil {
		return err
	} else {
		log := Log{UUID: string(uuid), Created: time.Now().Unix()}
		log.Action = action
		log.ActionUUID = actionId

		if err := Save(log, []byte(uuid)); err != nil {
			return err
		}

		o.Memo = append(o.Memo, string(uuid))

		if err := o.Save(); err != nil {
			return err
		}
	}

	return nil
}
