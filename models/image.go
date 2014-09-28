package models

import (
	"encoding/json"
	"fmt"
	"reflect"

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
		return "", fmt.Errorf("没有找到 Image 的数据")
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

func (image *Image) PutJSON(imageId, sign, json string) error {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return err
	} else if has == false {
		//新建 Image 记录
		key = utils.GeneralKey(fmt.Sprintf("%s+%s", GetObjectKey("image", imageId), sign))

		image.ImageId = imageId
		image.JSON = json

		if len(sign) > 0 {
			image.Sign = sign
			image.Encrypted = true
		}

		if e := image.Save(key); e != nil {
			return e
		} else {
			if len(sign) == 0 {
				if e := LedisDB.Set([]byte(fmt.Sprintf("%s+", GetObjectKey("image", imageId))), key); e != nil {
					return e
				}
			} else {
				if e := LedisDB.Set([]byte(fmt.Sprintf("%s-?", GetObjectKey("image", imageId), sign)), key); e != nil {
					return e
				}
			}
		}
	} else {
		//更新旧 Image 记录

		image.ImageId = imageId
		image.JSON = json

		if len(sign) > 0 {
			image.Sign = sign
			image.Encrypted = true
		}

		if e := image.Save(key); e != nil {
			return e
		}
	}

	return nil
}

func (image *Image) PutLocation(imageId, sign, location string) error {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("更新 Image 数据时没有找到相应的记录")
	} else {
		image.Location = location

		if e := image.Save(key); e != nil {
			return e
		}
	}

	return nil
}

func (image *Image) GetLocation(imageId, sign string, uploaded, checksumed bool) (string, error) {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return "", err
	} else if has == false {
		return "", fmt.Errorf("没有找到 Image 的数据")
	} else {
		if results, e := LedisDB.HMget(key, []byte("CheckSumed"), []byte("Uploaded"), []byte("Location")); e != nil {
			return "", e
		} else {
			checksum := results[0]
			upload := results[1]
			location := results[2]

			if utils.BytesToBool(checksum) != checksumed {
				return "", fmt.Errorf("没有找到 Image 的数据")
			}

			if utils.BytesToBool(upload) != uploaded {
				return "", fmt.Errorf("没有找到 Image 的数据")
			}

			return string(location), nil
		}
	}

	return "", nil
}

func (image *Image) PutUploaded(imageId, sign string, uploaded bool) error {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("更新 Image 数据时没有找到相应的记录")
	} else {
		image.Uploaded = uploaded

		if e := image.Save(key); e != nil {
			return e
		}
	}

	return nil
}

func (image *Image) PutSize(imageId, sign string, size int64) error {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("更新 Image 数据时没有找到相应的记录")
	} else {
		image.Size = size

		if e := image.Save(key); e != nil {
			return e
		}
	}

	return nil
}

func (image *Image) PutChecksum(imageId, sign, checksum string) error {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("更新 Image 数据时没有找到相应的记录")
	} else {
		image.Checksum = checksum

		if e := image.Save(key); e != nil {
			return e
		}
	}

	return nil
}

func (image *Image) PutPayload(imageId, sign, payload string) error {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("更新 Image 数据时没有找到相应的记录")
	} else {
		image.Payload = payload

		if e := image.Save(key); e != nil {
			return e
		}
	}

	return nil
}

func (image *Image) PutChecksumed(imageId, sign string, checksumed bool) error {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("更新 Image 数据时没有找到相应的记录")
	} else {
		image.CheckSumed = checksumed

		if e := image.Save(key); e != nil {
			return e
		}
	}

	return nil
}

//从 JSON 数据中解析查找是否存在 parent 的数据。
//如果存在 parent 数据，根据 imageId 和 sign 查找 parent 的记录。
//把当前 imageId 加入到 数组中在 Marshal 后保存在 ParentJSON 中。
//如果不存在 parent 数据，则认为当前 imageId 是 root 节点。
func (image *Image) PutAncestry(imageId, sign string) error {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("更新 Image 数据时没有找到相应的记录")
	} else {
		var imageJSONMap map[string]interface{}
		var parents []string
		var images []string

		//从数据库中读取 JSON 数据
		if imageJSON, err := LedisDB.HGet(key, []byte("JSON")); err != nil {
			return err
		} else {
			// JSON 数据解码到 image 对象中
			if err := json.Unmarshal(imageJSON, &imageJSONMap); err != nil {
				return err
			}

			//判断是否存在 parent 数据。
			if value, has := imageJSONMap["parent"]; has == true {
				//从数据库中读取 parent image 的数据
				i := new(Image)
				if h, k, e := i.Get(value.(string), sign); e != nil {
					return e
				} else if h == true {
					//从数据库中获得 parent image 的 key
					//从数据库中读取 parent image 的 ParentJSON 字段数据
					if j, e := LedisDB.HGet(k, []byte("ParentJSON")); e != nil {
						return e
					} else {
						if e := json.Unmarshal(j, &parents); e != nil {
							return e
						}
					}
				}
			}

			images = append(images, imageId)
			parents = append(images, parents...)

			parentJSON, _ := json.Marshal(parents)

			image.ParentJSON = string(parentJSON)

			if e := image.Save(key); e != nil {
				return e
			}
		}
	}

	return nil
}

func (image *Image) GetAncestry(imageId, sign string, uploaded, checksumed bool) ([]byte, error) {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return []byte(""), err
	} else if has == false {
		return []byte(""), fmt.Errorf("没有找到 Image 的数据")
	} else {
		if results, e := LedisDB.HMget(key, []byte("CheckSumed"), []byte("Uploaded"), []byte("ParentJSON")); e != nil {
			return []byte(""), e
		} else {
			checksum := results[0]
			upload := results[1]
			parent := results[2]

			if utils.BytesToBool(checksum) != checksumed {
				return []byte(""), fmt.Errorf("没有找到 Image 的数据")
			}

			if utils.BytesToBool(upload) != uploaded {
				return []byte(""), fmt.Errorf("没有找到 Image 的数据")
			}

			return parent, nil
		}
	}

	return []byte(""), nil
}

func (image *Image) Save(key []byte) error {
	s := reflect.TypeOf(image).Elem()

	//循环处理 Struct 的每一个 Field
	for i := 0; i < s.NumField(); i++ {
		//获取 Field 的 Value
		value := reflect.ValueOf(image).Elem().Field(s.Field(i).Index[0])

		//判断 Field 不为空
		if utils.IsEmptyValue(value) == false {
			switch value.Kind() {
			case reflect.String:
				if _, err := LedisDB.HSet(key, []byte(s.Field(i).Name), []byte(value.String())); err != nil {
					return err
				}
			case reflect.Bool:
				if _, err := LedisDB.HSet(key, []byte(s.Field(i).Name), utils.BoolToBytes(value.Bool())); err != nil {
					return err
				}
			case reflect.Int64:
				if _, err := LedisDB.HSet(key, []byte(s.Field(i).Name), utils.Int64ToBytes(value.Int())); err != nil {
					return err
				}
			default:
				return fmt.Errorf("不支持的数据类型 %s:%s", s.Field(i).Name, value.Kind().String())
			}
		}

	}

	return nil
}
