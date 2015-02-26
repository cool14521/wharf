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
	Repositories []string `json:"repositories"` //用户具备所有权的respository对应UUID信息,最新添加的Repository在最前面
}

type Team struct {
	UUID              string       `json:"UUID"`           //全局唯一的索引
	Team              string       `json:"team"`           //全局唯一,Team名称，不可修改，如果可以修改就要加code
	Organization      string       `json:"organization"`   //此Team属于哪个组织
	Username          string       `json:"username"`       //此Team属于哪个用户
	Description       string       `json:"description"`    //
	Users             []string     `json:"users"`          //已经加入此Team的所有User对应UUID
	TeamPrivileges    []string     `json:"teamprivileges"` //已经加入此Team的所有User对应的权限UUID,一个Team有统一的读写权限，权限不到个人
	Repositories      []string     `json:"repositories"`   //此Team所有Repository的UUID
	RepositoryObjects []Repository `json:"repositoryobjects"`
}

func (organization *Organization) Has(organizationName string) (bool, []byte, error) {
	UUID, err := GetUUID("organization", organizationName)
	if err != nil {
		return false, nil, err
	}
	if len(UUID) <= 0 {
		return false, nil, nil
	}

	err = Get(organization, UUID)

	return true, UUID, err
}

func (organization *Organization) Save() error {
	if err := Save(organization, []byte(organization.UUID)); err != nil {
		return err
	}

	if _, err := LedisDB.HSet([]byte(GLOBAL_ORGANIZATION_INDEX), []byte(organization.Organization), []byte(organization.UUID)); err != nil {
		return err
	}

	return nil
}

func (organization *Organization) Get(UUID string) error {
	if err := Get(organization, []byte(UUID)); err != nil {
		return err
	}

	return nil
}

func (organization *Organization) Remove() error {
	if _, err := LedisDB.HSet([]byte(fmt.Sprintf("%s_remove", GLOBAL_ORGANIZATION_INDEX)), []byte(organization.Organization), []byte(organization.UUID)); err != nil {
		return err
	}

	if _, err := LedisDB.HDel([]byte(GLOBAL_ORGANIZATION_INDEX), []byte(organization.UUID)); err != nil {
		return err
	}

	return nil
}

func (team *Team) Has(teamName string) (bool, []byte, error) {
	UUID, err := GetUUID("team", teamName)
	if err != nil {
		return false, nil, err
	}

	if len(UUID) <= 0 {
		return false, nil, nil
	}

	err = Get(team, UUID)

	return true, UUID, err
}

func (team *Team) Save() error {
	if err := Save(team, []byte(team.UUID)); err != nil {
		return err
	}

	if _, err := LedisDB.HSet([]byte(GLOBAL_TEAM_INDEX), []byte(team.Team), []byte(team.UUID)); err != nil {
		return err
	}

	return nil
}

func (team *Team) Get(UUID string) error {
	if err := Get(team, []byte(UUID)); err != nil {
		return err
	}

	return nil
}

func (team *Team) Remove() error {
	if _, err := LedisDB.HSet([]byte(fmt.Sprintf("%s_remove", GLOBAL_TEAM_INDEX)), []byte(team.Team), []byte(team.UUID)); err != nil {
		return err
	}

	if _, err := LedisDB.HDel([]byte(GLOBAL_TEAM_INDEX), []byte(team.UUID)); err != nil {
		return err
	}

	return nil
}
