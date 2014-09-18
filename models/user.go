package models

import (
	"fmt"
	"regexp"
	"time"

	"github.com/dockercn/docker-bucket/utils"
)

type User struct {
	Username string    //
	Password string    //
	Email    string    //Email 可以更换，全局唯一
	Fullname string    //
	Company  string    //
	Location string    //
	Mobile   string    //
	URL      string    //
	Gravatar string    //如果是邮件地址使用 gravatar.org 的 API 显示头像，如果是上传的用户显示头像的地址。
	Actived  bool      //
	Created  time.Time //
	Updated  time.Time //
	Log      string    //
}

func (user *User) Add(username string, passwd string, email string, actived bool) error {
	//检查是否存在用户
	if u, err := LedisDB.HGet([]byte(GetObjectKey("user", username)), []byte("Password")); err != nil {
		return nil
	} else if u != nil {
		//已经存在用户
		return fmt.Errorf("已经 %s 存在用户", username)
	} else {
		//检查用户名合法性，参考实现标准：
		//https://github.com/docker/docker/blob/28f09f06326848f4117baf633ec9fc542108f051/registry/registry.go#L27
		validNamespace := regexp.MustCompile(`^([a-z0-9_]{4,30})$`)
		if !validNamespace.MatchString(username) {
			return fmt.Errorf("用户名必须是 4 - 30 位之间，且只能由 a-z，0-9 和 下划线组成")
		}

		//检查密码合法性
		if len(passwd) < 5 {
			return fmt.Errorf("密码必须等于或大于 5 位字符以上")
		}

		//检查邮箱合法性
		validEmail := regexp.MustCompile(`^[a-z0-9A-Z]+([\-_\.][a-z0-9A-Z]+)*@([a-z0-9A-Z]+(-[a-z0-9A-Z]+)*\.)+[a-zA-Z]+$`)
		if !validEmail.MatchString(email) {
			return fmt.Errorf("Email 格式不合法")
		}

		//设置用户属性
		LedisDB.HSet([]byte(GetObjectKey("user", username)), []byte("Username"), []byte(username))
		LedisDB.HSet([]byte(GetObjectKey("user", username)), []byte("Password"), []byte(passwd))
		LedisDB.HSet([]byte(GetObjectKey("user", username)), []byte("Email"), []byte(email))
		LedisDB.HSet([]byte(GetObjectKey("user", username)), []byte("Actived"), utils.BoolToBytes(actived))

		//设置用户创建的时间
		LedisDB.HSet([]byte(GetObjectKey("user", username)), []byte("Updated"), utils.NowToBytes())
		LedisDB.HSet([]byte(GetObjectKey("user", username)), []byte("Created"), utils.NowToBytes())

		return nil
	}
}

func (this *User) Get(username string, passwd string, actived bool) (bool, error) {
	//读取密码和Actived的值进行判断是否存在用户
	if results, err := LedisDB.HMget([]byte(GetObjectKey("user", username)), []byte("Password"), []byte("Actived")); err != nil {
		return false, err
	} else {
		if password := results[0]; string(password) != passwd {
			return false, nil
		}

		if active := results[1]; utils.BytesToBool(active) != actived {
			return false, nil
		}

		return true, nil
	}
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
