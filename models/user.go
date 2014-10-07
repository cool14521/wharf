package models

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/dockercn/docker-bucket/utils"
)

type User struct {
	Username      string //
	Password      string //
	Repositories  string //
	Organizations string //
	Email         string //Email 可以更换，全局唯一
	Fullname      string //
	Company       string //
	Location      string //
	Mobile        string //
	URL           string //
	Gravatar      string //如果是邮件地址使用 gravatar.org 的 API 显示头像，如果是上传的用户显示头像的地址。
	Created       int64  //
	Updated       int64  //
}

//在全局 user 存储的的 Hash 中查询到 user 的 key，然后根据 key 再使用 Exists 方法查询是否存在数据
func (user *User) Has(username string) (bool, error) {
	if key, err := LedisDB.HGet([]byte(GetServerKeys("user")), []byte(GetObjectKey("user", username))); err != nil {
		return false, err
	} else {
		if exist, err := LedisDB.Exists(key); err != nil || exist == 0 {
			return false, err
		} else {
			return true, nil
		}
	}

}

func (user *User) Put(username string, passwd string, email string) error {
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

		key := utils.GeneralKey(username)

		user.Username = username
		user.Password = passwd
		user.Email = email

		user.Updated = time.Now().Unix()
		user.Created = time.Now().Unix()

		if err := user.Save(key); err != nil {
			return err
		} else {
			if err := LedisDB.Set([]byte(GetObjectKey("user", username)), key); err != nil {
				return err
			}

			return nil
		}
	}
}

func (user *User) Save(key []byte) error {
	s := reflect.TypeOf(user).Elem()

	//循环处理 Struct 的每一个 Field
	for i := 0; i < s.NumField(); i++ {
		//获取 Field 的 Value
		value := reflect.ValueOf(user).Elem().Field(s.Field(i).Index[0])

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

func (user *User) Get(username, passwd string) (bool, error) {
	//检查用户的 Key 是否存在
	if has, err := user.Has(username); err != nil {
		return false, err
	} else if has == true {
		var key []byte

		//获取用户对象的 Key
		if key, err = LedisDB.Get([]byte(GetObjectKey("user", username))); err != nil {
			return false, err
		}

		//读取密码的值进行判断是否密码相同
		if password, err := LedisDB.HGet(key, []byte("Password")); err != nil {
			return false, err
		} else {
			if string(password) != passwd {
				return false, nil
			}
			return true, nil
		}
	} else {
		//没有用户的 Key 存在
		return false, nil
	}
}

type Organization struct {
	Owner        string //用户的 Key，每个组织都由用户创建，Owner 默认是拥有所有 Repository 的读写权限
	Name         string //
	Description  string //保存 Markdown 格式
	Repositories string //
	Privileges   string //
	Users        string //
	Created      int64  //
	Updated      int64  //
}

func (org *Organization) Has(name string) (bool, error) {
	if org, err := LedisDB.Exists([]byte(GetObjectKey("org", name))); err != nil {
		return false, err
	} else if org > 0 {
		return true, nil
	}

	return false, nil
}

func (org *Organization) Add(user, name, description string) error {
	if has, err := org.Has(name); err != nil {
		return err
	} else if has == true {
		return fmt.Errorf("组织 %s 已经存在", name)
	} else {
		//检查用户的命名空间是否冲突
		u := new(User)
		if has, err := u.Has(name); err != nil {
			return err
		} else if has == true {
			return fmt.Errorf("%s 和用户名称冲突", name)
		}

		key := utils.GeneralKey(name)

		org.Owner = user
		org.Name = name
		org.Description = description

		org.Updated = time.Now().Unix()
		org.Created = time.Now().Unix()

		if err := org.Save(key); err != nil {
			return err
		} else {
			if err := LedisDB.Set([]byte(GetObjectKey("org", name)), key); err != nil {
				return err
			}

			return nil
		}
	}
}

func (org *Organization) Save(key []byte) error {
	s := reflect.TypeOf(org).Elem()

	//循环处理 Struct 的每一个 Field
	for i := 0; i < s.NumField(); i++ {
		//获取 Field 的 Value
		value := reflect.ValueOf(org).Elem().Field(s.Field(i).Index[0])

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
