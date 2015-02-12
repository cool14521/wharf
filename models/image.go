package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/dockercn/wharf/utils"
)

type Image struct {
	ImageId    string //
	JSON       string //
	Ancestry   string //
	Checksum   string //
	Payload    string //
	URL        string //
	Backend    string //
	Path       string //文件在服务器的存储路径
	Sign       string //
	Size       int64  //
	Uploaded   bool   //
	Checksumed bool   //
	Encrypted  bool   //是否加密
	Created    int64  //
	Updated    int64  //
}

func (image *Image) Get(imageId, sign string) (bool, []byte, error) {
	var k []byte

	if len(sign) == 0 {
		k = []byte(fmt.Sprintf("%s+", GetObjectKey("image", imageId)))
	} else {
		k = []byte(fmt.Sprintf("%s-?%s", GetObjectKey("image", imageId), sign))
	}

	if key, err := LedisDB.HGet([]byte(GetServerKeys("image")), k); err != nil {
		return false, []byte(""), err
	} else if key != nil {
		if image, err := LedisDB.HGet(key, []byte("ImageId")); err != nil {
			return false, []byte(""), err
		} else if image != nil {
			if string(image) != imageId {
				return true, key, fmt.Errorf("存在 Image 数据，但是 ImageID 不相同: %s", string(image))
			}
			return true, key, nil
		}
	}

	return false, []byte(""), nil
}

func (image *Image) GetPushed(imageId, sign string, uploaded, checksumed bool) (bool, error) {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return false, err
	} else if has == false {
		return false, fmt.Errorf("没有在数据库中查询到要更新的 Image 数据")
	} else {
		if results, e := LedisDB.HMget(key, []byte("Checksumed"), []byte("Uploaded")); e != nil {
			return false, e
		} else {

			if utils.BytesToBool(results[0]) != checksumed {
				return false, nil
			}

			if utils.BytesToBool(results[1]) != uploaded {
				return false, nil
			}

			return true, nil
		}
	}

	return false, nil
}

func (image *Image) GetJSON(imageId, sign string, uploaded, checksumed bool) ([]byte, error) {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return []byte(""), err
	} else if has == false {
		return []byte(""), fmt.Errorf("没有在数据库中查询到要更新的 Image 数据")
	} else {
		if results, e := LedisDB.HMget(key, []byte("Checksumed"), []byte("Uploaded"), []byte("JSON")); e != nil {
			return []byte(""), e
		} else {

			if utils.BytesToBool(results[0]) != checksumed {
				return []byte(""), nil
			}

			if utils.BytesToBool(results[1]) != uploaded {
				return []byte(""), nil
			}

			return results[2], nil
		}
	}

	return []byte(""), nil

}

func (image *Image) GetChecksum(imageId, sign string, uploaded, checksumed bool) (string, error) {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return "", err
	} else if has == false {
		return "", fmt.Errorf("没有在数据库中查询到要更新的 Image 数据")
	} else {
		if results, e := LedisDB.HMget(key, []byte("Checksumed"), []byte("Uploaded"), []byte("Checksum")); e != nil {
			return "", e
		} else {

			if utils.BytesToBool(results[0]) != checksumed {
				return "", nil
			}

			if utils.BytesToBool(results[1]) != uploaded {
				return "", nil
			}

			return string(results[2]), nil
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

		image.Uploaded = false
		image.Checksumed = false
		image.Encrypted = false

		image.Size = 0

		image.Updated = time.Now().Unix()
		image.Created = time.Now().Unix()

		if len(sign) > 0 {
			image.Sign = sign
			image.Encrypted = true
		}

		if e := image.Save(key); e != nil {
			return e
		} else {
			if len(sign) == 0 {
				if _, e := LedisDB.HSet([]byte(GetServerKeys("image")), []byte(fmt.Sprintf("%s+", GetObjectKey("image", imageId))), key); e != nil {
					return e
				}
			} else {
				if _, e := LedisDB.HSet([]byte(GetServerKeys("image")), []byte(fmt.Sprintf("%s-?", GetObjectKey("image", imageId), sign)), key); e != nil {
					return e
				}
			}
		}
	} else {
		//更新旧 Image 记录

		image.ImageId = imageId
		image.JSON = json

		image.Updated = time.Now().Unix()

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

func (image *Image) PutPath(imageId, sign, path string) error {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("没有在数据库中查询到要更新的 Image 数据")
	} else {
		image.Path = path

		image.Updated = time.Now().Unix()

		if e := image.Save(key); e != nil {
			return e
		}
	}

	return nil
}

func (image *Image) GetPath(imageId, sign string, uploaded, checksumed bool) (string, error) {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return "", err
	} else if has == false {
		return "", fmt.Errorf("没有在数据库中查询到要更新的 Image 数据")
	} else {
		if results, e := LedisDB.HMget(key, []byte("Checksumed"), []byte("Uploaded"), []byte("Path")); e != nil {
			return "", e
		} else {

			if utils.BytesToBool(results[0]) != checksumed {
				return "", fmt.Errorf("没有找到 Image 的数据")
			}

			if utils.BytesToBool(results[1]) != uploaded {
				return "", fmt.Errorf("没有找到 Image 的数据")
			}

			return string(results[2]), nil
		}
	}

	return "", nil
}

func (image *Image) PutUploaded(imageId, sign string, uploaded bool) error {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("没有在数据库中查询到要更新的 Image 数据")
	} else {
		image.Uploaded = uploaded

		image.Updated = time.Now().Unix()

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
		return fmt.Errorf("没有在数据库中查询到要更新的 Image 数据")
	} else {
		image.Size = size

		image.Updated = time.Now().Unix()

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
		return fmt.Errorf("没有在数据库中查询到要更新的 Image 数据")
	} else {
		image.Checksum = checksum

		image.Updated = time.Now().Unix()

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
		return fmt.Errorf("没有在数据库中查询到要更新的 Image 数据")
	} else {
		image.Payload = payload

		image.Updated = time.Now().Unix()

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
		return fmt.Errorf("没有在数据库中查询到要更新的 Image 数据")
	} else {
		image.Checksumed = checksumed

		image.Updated = time.Now().Unix()

		if e := image.Save(key); e != nil {
			return e
		}
	}

	return nil
}

//从 JSON 数据中解析查找是否存在 parent 的数据。
//如果存在 parent 数据，根据 imageId 和 sign 查找 parent 的记录。
//把当前 imageId 加入到 数组中在 Marshal 后保存在 Ancestry 中。
//如果不存在 parent 数据，则认为当前 imageId 是 root 节点。
func (image *Image) PutAncestry(imageId, sign string) error {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("没有在数据库中查询到要更新的 Image 数据")
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
					//从数据库中读取 parent image 的 Ancestry 字段数据
					if j, e := LedisDB.HGet(k, []byte("Ancestry")); e != nil {
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

			image.Ancestry = string(parentJSON)

			image.Updated = time.Now().Unix()

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
		return []byte(""), fmt.Errorf("没有在数据库中查询到要更新的 Image 数据")
	} else {
		if results, e := LedisDB.HMget(key, []byte("Checksumed"), []byte("Uploaded"), []byte("Ancestry")); e != nil {
			return []byte(""), e
		} else {

			if utils.BytesToBool(results[0]) != checksumed {
				return []byte(""), fmt.Errorf("没有找到 Image 的数据")
			}

			if utils.BytesToBool(results[1]) != uploaded {
				return []byte(""), fmt.Errorf("没有找到 Image 的数据")
			}

			return results[2], nil
		}
	}

	return []byte(""), nil
}

func (image *Image) GetSize(imageId, sign string, uploaded, checksumed bool) (int64, error) {
	if has, key, err := image.Get(imageId, sign); err != nil {
		return 0, err
	} else if has == false {
		return 0, fmt.Errorf("没有在数据库中查询到要更新的 Image 数据")
	} else {
		if results, e := LedisDB.HMget(key, []byte("Checksumed"), []byte("Uploaded"), []byte("Size")); e != nil {
			return 0, e
		} else {

			if utils.BytesToBool(results[0]) != checksumed {
				return 0, fmt.Errorf("没有找到 Image 的数据")
			}

			if utils.BytesToBool(results[1]) != uploaded {
				return 0, fmt.Errorf("没有找到 Image 的数据")
			}

			return utils.BytesToInt64(results[2]), nil
		}
	}

	return 0, nil
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
