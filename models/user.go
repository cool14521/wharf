package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/dockercn/docker-bucket/utils"
)

type User struct {
	Username      string    //
	Password      string    //
	Repositories  string    //
	Organizations string    //
	Email         string    //Email 可以更换，全局唯一
	Fullname      string    //
	Company       string    //
	Location      string    //
	Mobile        string    //
	URL           string    //
	Gravatar      string    //如果是邮件地址使用 gravatar.org 的 API 显示头像，如果是上传的用户显示头像的地址。
	Actived       bool      //
	Created       time.Time //
	Updated       time.Time //
	Logs          string    //用户日志信息
}

func (user *User) Has(username string) (bool, error) {
	if user, err := LedisDB.Exists([]byte(GetObjectKey("user", username))); err != nil {
		return false, err
	} else if user > 0 {
		return true, nil
	}

	return false, nil
}

func (user *User) Add(username string, passwd string, email string, actived bool) error {
	//检查用户的 Key 是否存在
	if has, err := user.Has(username); err != nil {
		return err
	} else if has == true {
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

func (user *User) Get(username string, passwd string, actived bool) (bool, error) {
	//检查用户的 Key 是否存在
	if has, err := user.Has(username); err != nil {
		return false, err
	} else if has == true {

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

	} else {
		//没有用户的 Key 存在
		return false, nil
	}
}

func (user *User) Log(username string, log string) error {
	var logs []string

	//判断是否存在用户
	if has, err := user.Has(username); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("没有找到用户 %s ", username)
	}

	//获取所有的 Key，判断 Logs 是不是在其中，以判断是否存在 Log 数据。
	if keys, err := LedisDB.HKeys([]byte(GetObjectKey("user", username))); err != nil {
		return err
	} else {
		has := false
		for _, key := range keys {
			if string(key) == "Logs" {
				has = true
			}
		}

		//如果已经存在记录，将原记录 decode 到 logs 数组。
		if has == true {
			//获取已有的
			if l, err := LedisDB.HGet([]byte(GetObjectKey("user", username)), []byte("Logs")); err != nil {
				//没有找到数据会返回 Error
				return err
			} else {
				//解码 Log 数据的数组
				if e := json.Unmarshal(l, logs); e != nil {
					return e
				}
			}
		}

		//向数组追加 Log 记录
		logs = append(logs, fmt.Sprintf("%d %s %s", time.Now().Unix, GetObjectKey("user", username), log))
		//Encode Log 数组，写入数据库
		if bytes, e := json.Marshal(logs); e != nil {
			return e
		} else {
			if _, e := LedisDB.HSet([]byte(GetObjectKey("user", username)), []byte("Logs"), bytes); e != nil {
				return e
			}

			return nil
		}
	}

}

type Organization struct {
	Owner        string    //用户的 Key，每个组织都由用户创建，Owner 默认是拥有所有 Repository 的读写权限
	Name         string    //
	Repositories string    //
	Privileges   string    //
	Users        string    //
	Actived      bool      //组织创建后就是默认激活的
	Created      time.Time //
	Updated      time.Time //
	Logs         string    //
}

func (org *Organization) Has(name string) (bool, error) {
	if org, err := LedisDB.Exists([]byte(GetObjectKey("org", name))); err != nil {
		return false, err
	} else if org > 0 {
		return true, nil
	}

	return false, nil
}

func (org *Organization) Get(name string, actived bool) (bool, error) {
	if has, err := org.Has(name); err != nil {
		return false, err
	} else if has == true {
		if active, err := LedisDB.HGet([]byte(GetObjectKey("org", name)), []byte("Actived")); err != nil {
			return false, err
		} else if utils.BytesToBool(active) == actived {
			return true, nil
		}

		return false, nil
	} else {
		//没有用户的 Key 存在
		return false, nil
	}
}

func (org *Organization) Log(name string, log string) error {
	var logs []string

	//判断是否存在用户
	if has, err := org.Has(name); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("没有找到组织 %s ", name)
	}

	//获取所有的 Key，判断 Logs 是不是在其中，以判断是否存在 Log 数据。
	if keys, err := LedisDB.HKeys([]byte(GetObjectKey("org", name))); err != nil {
		return err
	} else {
		has := false
		for _, key := range keys {
			if string(key) == "Logs" {
				has = true
			}
		}

		//如果已经存在记录，将原记录 decode 到 logs 数组。
		if has == true {
			//获取已有的
			if l, err := LedisDB.HGet([]byte(GetObjectKey("org", name)), []byte("Logs")); err != nil {
				//没有找到数据会返回 Error
				return err
			} else {
				//解码 Log 数据的数组
				if e := json.Unmarshal(l, logs); e != nil {
					return e
				}
			}
		}

		//向数组追加 Log 记录
		logs = append(logs, fmt.Sprintf("%d %s %s", time.Now().Unix, GetObjectKey("org", name), log))
		//Encode Log 数组，写入数据库
		if bytes, e := json.Marshal(logs); e != nil {
			return e
		} else {
			if _, e := LedisDB.HSet([]byte(GetObjectKey("org", name)), []byte("Logs"), bytes); e != nil {
				return e
			}

			return nil
		}
	}

}
