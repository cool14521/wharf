package ledis

type Object struct {
	Id      string   `json:"id"`      //
	Created int64    `json:"created"` //
	Updated int64    `json:"updated"` //
	Memo    []string `json:"memo"`    //
}

func (o *Object) Create() error {
	return nil
}

func (o *Object) Read() (Object, error) {
	return nil, nil
}

func (o *Object) Update() error {
	return nil
}

func (o *Object) Delete() error {
	return nil
}
