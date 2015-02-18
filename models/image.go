package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/utils"
)

type Image struct {
	UUID       string
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

func (image *Image) Has(imageName string) (isHas bool, UUID []byte, err error) {
	UUID, err = GetUUID("image", imageName)
	if err != nil {
		return false, nil, err
	}

	if len(UUID) <= 0 {
		return false, nil, nil
	}
	err = Get(image, UUID)

	return true, UUID, err
}

func (image *Image) Save() (err error) {
	err = Save(image, []byte(image.UUID))
	if err != nil {
		return err
	}
	_, err = LedisDB.HSet([]byte(GLOBAL_IMAGE_INDEX), []byte(image.ImageId), []byte(image.UUID))
	if err != nil {
		return err
	}
	return nil
}

func (image *Image) Get(UUID string) (err error) {
	err = Get(image, []byte(UUID))
	if err != nil {
		return err
	}
	return nil
}

func (image *Image) Remove() (err error) {
	_, err = LedisDB.HSet([]byte(fmt.Sprintf("%s_remove", GLOBAL_IMAGE_INDEX)), []byte(image.ImageId), []byte(image.UUID))
	if err != nil {
		return err
	}
	_, err = LedisDB.HDel([]byte(GLOBAL_IMAGE_INDEX), []byte(image.UUID))
	if err != nil {
		return err
	}
	return nil
}

func (image *Image) IsPushed(imageId string) (isPushed bool, err error) {

	isHas, _, err := image.Has(imageId)

	if err != nil {
		return isHas, fmt.Errorf("查找 Image 错误")
	}

	if !isHas {
		return isHas, nil
	}

	if image.Checksumed && image.Uploaded {
		return true, nil
	}

	return false, nil

}

//API 获取 image 的 JSON 使用
func (image *Image) GetJSON(imageId string) ([]byte, error) {
	isHas, _, err := image.Has(imageId)

	if err != nil {
		return nil, fmt.Errorf("查找 Image 错误")
	}

	if !isHas {
		return nil, fmt.Errorf("仓库不存在")
	}

	if !image.Checksumed || !image.Uploaded {
		return nil, fmt.Errorf("仓库没有上传完毕JSON无效")
	}

	return []byte(image.JSON), nil

}

func (image *Image) GetChecksum(imageId string) ([]byte, error) {
	isHas, _, err := image.Has(imageId)

	if err != nil {
		return nil, fmt.Errorf("查找 Image 错误")
	}

	if !isHas {
		return nil, fmt.Errorf("仓库不存在")
	}

	if !image.Checksumed || !image.Uploaded {
		return nil, fmt.Errorf("仓库没有上传完毕JSON无效")
	}

	return []byte(image.Checksum), nil
}

func (image *Image) PutJSON(imageId, json string) error {
	beego.Error("PutJSON")
	isHas, _, err := image.Has(imageId)
	if err != nil {
		return err
	}

	if !isHas {
		image.UUID = string(utils.GeneralKey(imageId))
		image.Created = time.Now().Unix()

	}

	image.ImageId = imageId
	image.JSON = json

	image.Uploaded = false
	image.Checksumed = false
	image.Encrypted = false

	image.Size = 0

	image.Updated = time.Now().Unix()

	err = image.Save()
	if err != nil {
		return err
	}

	//------------------------------------------------------------------------

	return nil
}

func (image *Image) PutLayer(imageId string, path string, uploaded bool, size int64) error {
	beego.Error("PutLayer")
	isHas, _, err := image.Has(imageId)
	if err != nil {
		return err
	}

	if !isHas {
		return fmt.Errorf("不存在put layer对应的image记录")
	}
	image.Path = path
	image.Uploaded = uploaded
	image.Size = size
	image.Updated = time.Now().Unix()

	err = image.Save()
	if err != nil {
		return err
	}

	//------------------------------------------------------------------------

	return nil

}

func (image *Image) PutChecksum(imageId string, checksum string, checksumed bool, payload string) error {
	beego.Error("PutChecksum")
	isHas, _, err := image.Has(imageId)
	if err != nil {
		return err
	}

	if !isHas {
		return fmt.Errorf("不存在put checksum对应的image记录")
	}
	err = image.PutAncestry(imageId)
	if err != nil {
		return fmt.Errorf("Ancestry计算错误", err.Error())
	}

	image.Checksum = checksum
	image.Checksumed = checksumed
	image.Payload = payload
	image.Updated = time.Now().Unix()

	err = image.Save()
	if err != nil {
		return err
	}

	//------------------------------------------------------------------------

	return nil

}

//从 JSON 数据中解析查找是否存在 parent 的数据。
//如果存在 parent 数据，根据 imageId 和 sign 查找 parent 的记录。
//把当前 imageId 加入到 数组中在 Marshal 后保存在 Ancestry 中。
//如果不存在 parent 数据，则认为当前 imageId 是 root 节点。
func (image *Image) PutAncestry(imageId string) error {
	beego.Error("PutAncestry")
	isHas, _, err := image.Has(imageId)
	if err != nil {
		return err
	}

	if !isHas {
		return fmt.Errorf("不存在put checksum对应的image记录")
	}
	var imageJSONMap map[string]interface{}
	var imageAncestry []string
	if err := json.Unmarshal([]byte(image.JSON), &imageJSONMap); err != nil {
		return err
	}
	//判断是否存在 parent 数据。
	if value, has := imageJSONMap["parent"]; has == true {
		//从数据库中读取 parent image 的数据
		parentImage := new(Image)
		//	beego.Error("Image.Has UUID:::~~~~~~~~~~~!!!!!!!!", value.(string))
		parentIsHas, _, err := parentImage.Has(value.(string))
		if err != nil {
			return err
		}
		//		beego.Error("Image.Has parentImage.Ancestry@@@@@@@@@@@@@@@@@", parentImage.Ancestry)
		if !parentIsHas {
			return fmt.Errorf("不存在parent image记录:::", value.(string))
		}
		var parentAncestry []string
		json.Unmarshal([]byte(parentImage.Ancestry), &parentAncestry)
		imageAncestry = append(imageAncestry, imageId)
		imageAncestry = append(imageAncestry, parentAncestry...)
		beego.Error("存在Parent::::::::::Image.Has parentImage.Ancestry@@@@@@@@@@@@@@@@@", imageAncestry)
	} else {

		beego.Error("NO存在Parent::::::::::计算前Image.Has parentImage.Ancestry@@@@@@@@@@@@@@@@@", imageAncestry)

		imageAncestry = append(imageAncestry, imageId)
		beego.Error("NO存在Parent::::::::::计算后Image.Has parentImage.Ancestry@@@@@@@@@@@@@@@@@", imageAncestry)
	}

	ancestryJSON, _ := json.Marshal(imageAncestry)
	image.Ancestry = string(ancestryJSON)
	image.Save()

	//	beego.Error("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&image.Ancestry::::::", image.Ancestry)

	return nil
}
