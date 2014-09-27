package models

import (
	"fmt"

	"github.com/dockercn/docker-bucket/utils"
)

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
	Encrypted  bool   //是否加密
	Created    int64  //
	Updated    int64  //
}

func (image *Image) Get(imageId, sign string) (bool, []byte, error) {
	if len(sign) == 0 {

		if exist, err := LedisDB.Exists([]byte(GetObjectKey("image", imageId))); err != nil {
			return false, []byte(""), err
		} else if exist > 0 {
			if key, e := LedisDB.Get([]byte(GetObjectKey("image", imageId))); e != nil {
				return false, []byte(""), e
			} else {
				return true, key, nil
			}
		}

	} else {

		if exist, err := LedisDB.Exists([]byte(fmt.Sprintf("%s-?%s", []byte(GetObjectKey("image", imageId)), sign))); err != nil {
			return false, []byte(""), err
		} else if exist > 0 {
			if key, e := LedisDB.Get([]byte(fmt.Sprintf("%s-?%s", []byte(GetObjectKey("image", imageId)), sign))); e != nil {
				return false, []byte(""), e
			} else {
				return true, key, nil
			}
		}
	}

	return false, []byte(""), nil
}

func (image *Image) GetPushed(imageId, sign string, uploaded, checksumed bool) (bool, error) {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return false, err
	} else if has == false {
		return false, nil
	} else {
		if results, e := LedisDB.HMget(key, []byte("CheckSumed"), []byte("Uploaded")); e != nil {
			return false, e
		} else {
			checksum := results[0]
			upload := results[1]

			if utils.BytesToBool(checksum) != checksumed {
				return false, nil
			}

			if utils.BytesToBool(upload) != uploaded {
				return false, nil
			}

			return true, nil
		}
	}

	return false, nil
}

func (image *Image) GetJSON(imageId, sign string, uploaded, checksumed bool) (string, error) {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return "", err
	} else if has == false {
		return "", nil
	} else {
		if results, e := LedisDB.HMget(key, []byte("CheckSumed"), []byte("Uploaded"), []byte("JSON")); e != nil {
			return "", e
		} else {
			checksum := results[0]
			upload := results[1]
			json := results[2]

			if utils.BytesToBool(checksum) != checksumed {
				return "", nil
			}

			if utils.BytesToBool(upload) != uploaded {
				return "", nil
			}

			return string(json), nil
		}
	}

	return "", nil

}

func (image *Image) GetChecksum(imageId, sign string, uploaded, checksumed bool) (string, error) {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return "", err
	} else if has == false {
		return "", nil
	} else {
		if results, e := LedisDB.HMget(key, []byte("CheckSumed"), []byte("Uploaded"), []byte("Checksum")); e != nil {
			return "", e
		} else {
			checksum := results[0]
			upload := results[1]
			c := results[2]

			if utils.BytesToBool(checksum) != checksumed {
				return "", nil
			}

			if utils.BytesToBool(upload) != uploaded {
				return "", nil
			}

			return string(c), nil
		}
	}

	return "", nil

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
