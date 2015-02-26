package models

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
	ActionUUID int64  `json:"actionuuid"` //
	Level      int64  `json:"level"`      //
	Created    int64  `json:"created"`    //
}
