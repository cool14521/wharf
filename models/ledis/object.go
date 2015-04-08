package ledis

type Object struct {
	Id      string   `json:"id"`      //
	Created int64    `json:"created"` //
	Updated int64    `json:"updated"` //
	Memo    []string `json:"memo"`    //
}

func (o *Object) Create() error {

}

func (o *Object) Read() (Object, error) {

}

func (o *Object) Update() error {

}

func (o *Object) Delete() error {

}
