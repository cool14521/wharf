package models

import (
	"fmt"
)

type Organization struct {
	UUID         string   `json:"UUID"`         //全局唯一的索引
	Organization string   `json:"organization"` //全局唯一，组织名称,不可修改，如果可以修改就要加code
	Username     string   `json:"username"`     //创建组织的用户名
	Description  string   `json:"description"`  //组织说明，保存markdown 格式 方便展现
	Created      int64    `json:"created"`      //组织创建时间
	Updated      int64    `json:"updated"`      //组织信息修改时间
	Teams        []string `json:"teams"`        //当前组织下所有的Team的UUID
}

func (organization *Organization) Has(organizationName string) (isHas bool, UUID []byte, err error) {
	UUID, err = GetUUID("organization", organizationName)
	if err != nil {
		return false, nil, err
	}
	if len(UUID) <= 0 {
		return false, nil, nil
	}
	err = Get(organization, UUID)
	return true, UUID, err
}

func (organization *Organization) Save() (err error) {
	err = Save(organization, []byte(organization.UUID))
	if err != nil {
		return err
	}
	_, err = LedisDB.HSet([]byte(GLOBAL_ORGANIZATION_INDEX), []byte(organization.Organization), []byte(organization.UUID))
	if err != nil {
		return err
	}
	return nil
}

func (organization *Organization) Get(UUID string) (err error) {
	err = Get(organization, []byte(UUID))
	if err != nil {
		return err
	}
	return nil
}

func (organization *Organization) Remove() (err error) {
	_, err = LedisDB.HSet([]byte(fmt.Sprintf("%s_remove", GLOBAL_ORGANIZATION_INDEX)), []byte(organization.Organization), []byte(organization.UUID))
	if err != nil {
		return err
	}
	_, err = LedisDB.HDel([]byte(GLOBAL_ORGANIZATION_INDEX), []byte(organization.UUID))
	if err != nil {
		return err
	}
	return nil
}

type Team struct {
	UUID           string   `json:"UUID"`         //全局唯一的索引
	Team           string   `json:"team"`         //全局唯一,Team名称，不可修改，如果可以修改就要加code
	Organization   string   `json:"organization"` //此Team属于哪个组织
	Username       string   `json:"username"`     //此Team属于哪个用户
	Description    string   `json:"description"`
	Users          []string `json:"users"`          //已经加入此Team的所有User对应UUID
	TeamPrivileges []string `json:"teamprivileges"` //已经加入此Team的所有User对应的权限UUID,一个Team有统一的读写权限，权限不到个人
	Repositories   []string `json:"repositories"`   //此Team所有Repository的UUID

}

func (team *Team) Has(teamName string) (isHas bool, UUID []byte, err error) {
	UUID, err = GetUUID("team", teamName)
	if err != nil {
		return false, nil, err
	}
	if len(UUID) <= 0 {
		return false, nil, nil
	}
	err = Get(team, UUID)
	return true, UUID, err
}

func (team *Team) Save() (err error) {
	err = Save(team, []byte(team.UUID))
	if err != nil {
		return err
	}
	_, err = LedisDB.HSet([]byte(GLOBAL_TEAM_INDEX), []byte(team.Team), []byte(team.UUID))
	if err != nil {
		return err
	}
	return nil
}

func (team *Team) Get(UUID string) (err error) {
	err = Get(team, []byte(UUID))
	if err != nil {
		return err
	}
	return nil

}

func (team *Team) Remove() (err error) {
	_, err = LedisDB.HSet([]byte(fmt.Sprintf("%s_remove", GLOBAL_TEAM_INDEX)), []byte(team.Team), []byte(team.UUID))
	if err != nil {
		return err
	}
	_, err = LedisDB.HDel([]byte(GLOBAL_TEAM_INDEX), []byte(team.UUID))
	if err != nil {
		return err
	}
	return nil
}

type Start struct {
	UUID       string `json:"UUID"`       //全局唯一的索引
	User       string `json:"user"`       //用户UUID，代表哪个用户加的星
	Repository string `json:"repository"` //仓库UUID，代表给哪个仓库加的星
	Time       int64  `json:"time"`       //代表加星的时间
}
type Comment struct {
	UUID       string `json:"UUID"`       //全局唯一的索引
	Comment    string `json:"comment"`    //评论的内容 markdown 格式保存
	User       string `json:"user"`       //用户UUID，代表哪个用户进行的评论
	Repository string `json:"repository"` //仓库UUID，代表评论的哪个仓库
	Time       int64  `json:"time"`       //代表评论的时间
}
type Privilege struct {
	UUID       string `json:"UUID"`       //全局唯一的索引
	Privilege  bool   `json:"privilege"`  //true 为读写，false为只读
	Team       string `json:"team"`       //此权限所属Team的UUID
	Repository string `json:"repository"` //此权限对应的仓库UUID
}

func (privilege *Privilege) Get(UUID string) (err error) {
	err = Get(privilege, []byte(UUID))
	if err != nil {
		return err
	}
	return nil

}

type Tag struct {
	UUID       string
	Name       string //
	ImageId    string //
	Namespace  string
	Repository string
	Sign       string //
}

func (tag *Tag) Has(namespace, repository, imageName, tagName string) (isHas bool, UUID []byte, err error) {
	UUID, err = GetUUID("tag", fmt.Sprintf("%s:%s:%s:%s", namespace, repository, imageName, tagName))
	if err != nil {
		return false, nil, err
	}
	if len(UUID) <= 0 {
		return false, nil, nil
	}
	err = Get(tag, UUID)
	return true, UUID, err
}

func (tag *Tag) Save() (err error) {
	err = Save(tag, []byte(tag.UUID))
	if err != nil {
		return err
	}
	_, err = LedisDB.HSet([]byte(GLOBAL_TAG_INDEX), []byte(fmt.Sprintf("%s:%s:%s:%s", tag.Namespace, tag.Repository, tag.ImageId, tag.Name)), []byte(tag.UUID))
	if err != nil {
		return err
	}
	return nil
}

func (tag *Tag) GetByUUID(uuid string) (err error) {
	err = Get(tag, []byte(uuid))
	return err
}
