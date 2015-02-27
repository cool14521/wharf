package models

import (
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
