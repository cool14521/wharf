package models

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dockercn/docker-bucket/utils"
)

const (
	FROM_CLI = "cli"
	FROM_WEB = "web"
)

const (
	PUBLIC  = "public"
	PRIVATE = "private"
)

const (
	ACTION_SIGNIN         = "signin"
	ACTION_SINGOUT        = "signout"
	ACTION_UPDATE_PROFILE = "update_profile"
	ACTION_ADD_REPO       = "add_repository"
	ACTION_UPDATE_REPO    = "update_repository"
	ACTION_DEL_REPO       = "del_repository"
	ACTION_ADD_COMMENT    = "add_comment"
	ACTION_DEL_COMMENT    = "del_comment"
	ACTION_ADD_ORG        = "add_org"
	ACTION_DEL_ORG        = "del_org"
	ACTION_ADD_MEMBER     = "add_member"
	ACTION_DEL_MEMBER     = "del_member"
	ACTION_ADD_STAR       = "add_star"
	ACTION_DEL_STAR       = "del_star"
)

type User struct {
	Username string    //
	Password string    //
	Email    string    //Email 可以更换，全局唯一
	Quota    int64     //可以创建 Repository 的数量
	Size     int64     //所有 Repository 存储总和，单位 G
	Fullname string    //
	Company  string    //
	Location string    //
	Mobile   string    //
	URL      string    //
	Gravatar string    //如果是邮件地址使用 gravatar.org 的 API 显示头像，如果是上传的用户显示头像的地址。
	Actived  bool      //默认创建的用户是不激活
	Created  time.Time //
	Updated  time.Time //
	Log      string    //
}

func (user *User) Add(username string, passwd string, email string, actived bool) error {
	if u, err := LedisDB.HGet([]byte(GetObjectKey("user", username)), []byte("Password")); err != nil {
		return nil
	} else if u != nil {
		return errors.New("已经存在用户!")
	} else {
		LedisDB.HSet([]byte(utils.ToString("@", username)), []byte("Username"), []byte(username))
		LedisDB.HSet([]byte(utils.ToString("@", username)), []byte("Password"), []byte(passwd))
		LedisDB.HSet([]byte(utils.ToString("@", username)), []byte("Email"), []byte(email))

		rst := make([]byte, 0)
		LedisDB.HSet([]byte(utils.ToString("@", username)), []byte("Actived"), strconv.AppendBool(rst, actived))

		return nil
	}
}

func (this *User) GetUserInfo(username string, userInfoKey string) (userInfo string, errorInfo error) {
	info, err := LedisDB.HGet([]byte(utils.ToString("@", username)), []byte(userInfoKey))
	return string(info), err
}

func (this *User) Get(username string, passwd string, defValue bool) (isAuth bool, errorInfo error) {
	userPassword, userPasswordErr := LedisDB.HGet([]byte(utils.ToString("@", username)), []byte("Password"))
	if userPasswordErr == nil && passwd == string(userPassword) {
		return true, nil
	} else {
		return false, userPasswordErr
	}
}

func (this *User) History(index int64, repositoryId int64, log string) {
	fmt.Println(index)
}
func (this *User) SetToken(username, token string) (errorInfo error) {
	//UserToken保存位置需要讨论-fivestarsky
	LedisDB.HSet([]byte(utils.ToString("@", username)), []byte("Token"), []byte(token))
	return nil
}

type Organization struct {
	Owner   string    //用户的 ID，每个组织都由用户创建，Owner 默认是拥有所有 Repository 的读写权限
	Name    string    //
	Email   string    //未来的 Billing 邮件地址
	Quota   int64     //可以创建 Repository 的数量
	Size    int64     //所有 Repository 存储总和，单位 G
	Actived bool      //组织创建后就是默认激活的
	Created time.Time //
	Updated time.Time //
	Log     string    //
}

//TODO 组织和用户之间的对应关系 struct
