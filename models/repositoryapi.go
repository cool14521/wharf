package models

import (
	"fmt"
	"time"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/utils"
)

func (repo *Repository) DoPut(namespace, repository, json, agent string) error {
	isHas, _, err := repo.Has(namespace, repository)
	if err != nil {
		return err
	}

	if !isHas {
		repo.UUID = string(utils.GeneralKey(fmt.Sprintf("%s:%s", namespace, repository)))
		repo.Created = time.Now().Unix()
	}
	beego.Debug("0")
	repo.Namespace = namespace
	repo.Repository = repository
	repo.JSON = json
	repo.Agent = agent
	repo.Updated = time.Now().Unix()

	// 将put状态设置为后续put操作需要
	repo.Checksumed = false
	repo.Uploaded = false

	err = repo.Save()
	if err != nil {
		return err
	}

	return nil
}

func (repo *Repository) PutTag(imageId, namespace, repository, tag string) error {

	isHas, _, err := repo.Has(namespace, repository)

	if err != nil {
		return err
	}
	if !isHas {
		return fmt.Errorf("没有在数据库中查询到要更新的 Repository 数据")
	}

	image := new(Image)
	isHas, _, err = image.Has(imageId)
	if err != nil {
		return err
	}
	if !isHas {
		return fmt.Errorf("没有查询到 tag 依赖的 image 数据")
	}

	nowTag := new(Tag)
	//isHas, _, err = nowTag.Has(namespace, repository, imageId, tag)
	//		if err != nil {
	//		return fmt.Errorf("查询 tag 数据异常")
	//	}

	//if !isHas {
	nowTag.UUID = string(fmt.Sprintf("%s:%s:%s", namespace, repository, tag))
	//	}

	nowTag.Name = tag
	nowTag.ImageId = imageId
	nowTag.Namespace = namespace
	nowTag.Repository = repository
	err = nowTag.Save()
	if err != nil {
		return fmt.Errorf("保存 tag 数据异常")
	}
	repo.Tags = append(repo.Tags, nowTag.UUID)
	fmt.Println("*********************************", repo.Tags)
	repo.Save()

	return nil
}

func (repo *Repository) PutImages(namespace, repository string) error {

	isHas, _, err := repo.Has(namespace, repository)

	if err != nil {
		return err
	}
	if !isHas {
		return fmt.Errorf("没有在数据库中查询到要更新的 Repository 数据")
	}

	//repo.Checksum
	repo.Checksumed = true
	repo.Uploaded = true
	repo.Updated = time.Now().Unix()

	return nil
}
