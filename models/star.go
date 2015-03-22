package models

type Star struct {
	Id     string   `json:"id"`     //
	User   string   `json:"user"`   //
	Object string   `json:"object"` //
	Time   int64    `json:"time"`   //
	Memo   []string `json:"memo"`   //
}

func (star *Star) Save() error {
	if err := Save(star, []byte(star.Id)); err != nil {
		return err
	}

	return nil
}
