package models

type Image struct {
	ImageId    string //
	JSON       string //
	ParentJSON string //
	Checksum   string //
	Payload    string //
	URL        string //
	Backend    string //
	Location   string //文件在服务器的存储路径
	Sign       string //
	Size       int64  //
	Uploaded   bool   //
	CheckSumed bool   //
	Privated   bool   //
	Encrypted  bool   //是否加密
	Created    int64  //
	Updated    int64  //
}

func (image *Image) Get(imageId string) (bool, error) {
	return false, nil
}

func (image *Image) GetPushed(imageId string, uploaded, checksumed bool) (bool, error) {
	return false, nil
}

func (image *Image) UpdateJSON(json string) (bool, error) {
	return false, nil
}

func (image *Image) Insert(imageId, json string) (bool, error) {
	return true, nil
}

func (image *Image) UpdateChecksum(checksum string) (bool, error) {
	return true, nil
}

func (image *Image) UpdatePayload(payload string) (bool, error) {
	return true, nil
}

func (image *Image) UpdateSize(size int64) (bool, error) {
	return true, nil
}

func (image *Image) UpdateUploaded(uploaded bool) (bool, error) {
	return true, nil
}

func (image *Image) UpdateChecksumed(checksumed bool) (bool, error) {
	return true, nil
}

func (image *Image) UpdateParentJSON() (bool, error) {
	return true, nil
}
