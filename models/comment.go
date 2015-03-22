package models

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
