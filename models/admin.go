package models

type Admin struct {
	UUID     string `json:"UUID"`     //
	Username string `json:"username"` //
	Password string `json:"password"` //
	Email    string `json:"email"`    //
	Created  int64  `json:"created"`  //
	Updated  int64  `json:"updated"`  //
	Memo     string `json:"memo"`     //
}
