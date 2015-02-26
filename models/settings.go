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

type EmailMessage struct {
	UUID     string   `json:"UUID"` //
	User     string   `json:""`     //
	Server   string   `json:""`     //
	Template string   `json:""`     //
	Object   string   `json:""`     //
	Message  string   `json:""`     //
	Status   string   `json:""`     //
	Count    int64    `json:""`     //
	Created  int64    `json:""`     //
	Updated  int64    `json:""`     //
	Memo     []string `json:""`     //
}

type EmailServer struct {
	UUID    string   `json:"UUID"` //
	Name    string   `json:""`     //
	Host    string   `json:""`     //
	Port    int64    `json:""`     //
	User    string   `json:""`     //
	Passwd  string   `json:""`     //
	API     string   `json:""`     //
	Created int64    `json:""`     //
	Updated int64    `json:""`     //
	Memo    []string `json:""`     //
}

type EmailTemplate struct {
	UUID    string    `json:"UUID"` //
	Server  int64     `json:""`     //
	Name    string    `json:""`     //
	Content string    `json:""`     //
	Created time.Time `json:""`     //
	Updated time.Time `json:""`     //
	Memo    []string  `json:""`     //
}
