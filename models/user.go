package models

import (
	"fmt"
	"regexp"
)

type User struct {
	UUID          string   `json:"UUID"`          //全局唯一的索引, LedisDB UserList 保存全局所有的用户UUID列表信息，LedisDB独立保存每个用户信息到一个hash,名字为{UUID}
	Username      string   `json:"username"`      //用于保存用户的登录名,全局唯一
	Password      string   `json:"password"`      //保存用户MD5后的密码
	Email         string   `json:"email"`         //保存用户注册的密码，bucket 项目是否有必要
	Fullname      string   `json:"fullname"`      //保存用户全名，还是昵称
	Company       string   `json:"company"`       //用户所属公司
	Location      string   `json:"location"`      //用户所在地
	Mobile        string   `json:"mobile"`        //用户电话
	URL           string   `json:"url"`           //用户主页URL
	Gravatar      string   `json:"gravatar"`      //用户头像地址 如果是邮件地址，使用gravatar.org 进行解析
	Created       int64    `json:"created"`       //用户创建时间
	Updated       int64    `json:"updated"`       //用户信息更新时间
	Repositories  []string `json:"repositories"`  //用户具备所有权的respository对应UUID信息,最新添加的Repository在最前面
	Organizations []string `json:"organizations"` //用户所有的组织对应UUID信息，最新添加的在最前面
	Teams         []string `json:"teams"`         //用户所有的Team对应UUID信息，最新添加的在最前面
	Starts        []string `json:"starts"`        //用户加星的Respository对应UUID信息,最新添加的在最前面
	Comments      []string `json:"comments"`      //和用户相关的所有评论对应UUID信息，包括自己发的评论和别人评论相关自己的，最新的评论在最前面
}

func (user *User) Has(username string) (bool, []byte, error) {

	UUID, err := GetUUID("user", username)

	if err != nil {
		return false, nil, err
	}

	if len(UUID) <= 0 {
		return false, nil, nil
	}

	err = Get(user, UUID)

	return true, UUID, err
}

func (user *User) Save() error {
	//https://github.com/docker/docker/blob/28f09f06326848f4117baf633ec9fc542108f051/registry/registry.go#L27
	validNamespace := regexp.MustCompile(`^([a-z0-9_]{4,30})$`)
	if !validNamespace.MatchString(user.Username) {
		return fmt.Errorf("Username must be 4 - 30, include a-z, 0-9 and '_'")
	}

	if len(user.Password) < 5 {
		return fmt.Errorf("Password length should be more than 5")
	}

	validEmail := regexp.MustCompile("[\\w!#$%&'*+/=?^_`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[a-zA-Z0-9](?:[\\w-]*[\\w])?")
	if !validEmail.MatchString(user.Email) {
		return fmt.Errorf("Email illegal")
	}

	if err := Save(user, []byte(user.UUID)); err != nil {
		return err
	}

	if _, err := LedisDB.HSet([]byte(GLOBAL_USER_INDEX), []byte(user.Username), []byte(user.UUID)); err != nil {
		return err
	}

	return nil
}

func (user *User) Remove() error {
	if _, err := LedisDB.HSet([]byte(fmt.Sprintf("%s_remove", GLOBAL_USER_INDEX)), []byte(user.Username), []byte(user.UUID)); err != nil {
		return err
	}

	if _, err := LedisDB.HDel([]byte(GLOBAL_USER_INDEX), []byte(user.Username)); err != nil {
		return err
	}

	return nil
}

func (user *User) Get(username, password string) error {
	if exist, UUID, err := user.Has(username); err != nil {
		return err
	} else if exist == false && err == nil {
		return fmt.Errorf("User is not exist: %s", username)
	} else if exist == true && err == nil {
		if err := Get(user, UUID); err != nil {
			return err
		} else {
			if user.Password != password {
				return fmt.Errorf("User password error.")
			} else {
				return nil
			}
		}
	}

	return nil
}

func (user *User) Orgs(username string) (map[string]string, error) {
	result := map[string]string{}

	if exist, _, err := user.Has(username); err != nil {
		return nil, err
	} else if exist == false && err == nil {
		return nil, fmt.Errorf("User is not exist: %s", username)
	} else if exist == true && err == nil {
		for _, uuid := range user.Organizations {
			var org Organization

			if err := org.Get(uuid); err == nil {
				result[org.Organization] = org.UUID
			}
		}
	}

	return result, nil
}
