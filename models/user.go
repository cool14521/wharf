package models

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/dockercn/docker-bucket/utils"
)

const (
	ORG_MEMBER = "M"
	ORG_OWNER  = "O"
)

type User struct {
	Username      string //
	Password      string //
	Repositories  string //用户的所有 Respository
	Organizations string //用户所属的所有组织
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
func (user *User) Has(username string) (bool, []byte, error) {
	if key, err := LedisDB.HGet([]byte(GetServerKeys("user")), []byte(GetObjectKey("user", username))); err != nil {
		return false, []byte(""), err
	} else if key != nil {
		if name, err := LedisDB.HGet(key, []byte("Username")); err != nil {
			return false, []byte(""), err
		} else if name != nil {
			//已经存在了用户的 Key，接着判断用户是否相同
			if string(name) != username {
				return true, key, fmt.Errorf("已经存在了 Key，但是用户名不相同")
			}

			return true, key, nil
		} else {
			return false, []byte(""), nil
		}
	} else {
		return false, []byte(""), nil
	}
}

//创建用户数据，如果存在返回错误信息。
func (user *User) Put(username string, passwd string, email string) error {
	//检查用户的 Key 是否存在
	if has, _, err := user.Has(username); err != nil {
		return err
	} else if has == true {
		//已经存在用户
		return fmt.Errorf("已经存在用户 %s", username)
	} else {
		//检查用户名和 Organization 的 Name 是不是冲突
		org := new(Organization)
		if h, _, e := org.Has(username); e != nil {
			return err
		} else if h == true {
			return fmt.Errorf("已经存在相同的组织 %s", username)
		}

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

		//生成随机的 Object Key
		key := utils.GeneralKey(username)

		user.Username = username
		user.Password = passwd
		user.Email = email

		user.Updated = time.Now().Unix()
		user.Created = time.Now().Unix()

		//保存 User 对象的数据
		if err := user.Save(key); err != nil {
			return err
		} else {
			//在全局 @user 数据中保存 key 信息
			if _, err := LedisDB.HSet([]byte(GetServerKeys("user")), []byte(GetObjectKey("user", username)), key); err != nil {
				return err
			}

			return nil
		}
	}
}

//根据用户名和密码获取用户
func (user *User) Get(username, passwd string) (bool, error) {
	//检查用户的 Key 是否存在
	if has, key, err := user.Has(username); err != nil {
		return false, err
	} else if has == true {

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

//用户创建镜像仓库后，在 user 的 Repositories 字段保存相应的记录
func (user *User) AddRepository(username, repository, key string) error {
	return nil
}

//用户删除镜像仓库后，在 user 的 Repositories 中删除相应的记录
func (user *User) RemoveRepository(username, repository string) error {
	return nil
}

//重置用户的密码
func (user *User) ResetPasswd(username, password string) error {
	//检查用户的 Key 是否存在
	if has, key, err := user.Has(username); err != nil {
		return err
	} else if has == true {
		user.Password = password

		if err := user.Save(key); err != nil {
			return err
		}
	} else {
		//没有用户的 Key 存在
		return fmt.Errorf("不存在用户 %s", username)
	}

	return nil
}

//向用户添加 Organization 数据
func (user *User) AddOrganization(username, org, member string) error {
	return nil
}

//从用户中删除 Organization 数据
func (user *User) RemoveOrganization(username, org string) error {
	return nil
}

//循环 User 的所有 Property ，保存数据
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

//在全局 org 存储的的 Hash 中查询到 org 的 key，然后根据 key 再使用 Exists 方法查询是否存在数据
func (org *Organization) Has(name string) (bool, []byte, error) {
	if key, err := LedisDB.HGet([]byte(GetServerKeys("org")), []byte(GetObjectKey("org", name))); err != nil {
		return false, []byte(""), err
	} else if key != nil {
		if exist, err := LedisDB.Exists(key); err != nil || exist == 0 {
			return false, []byte(""), err
		} else {
			return true, key, nil
		}
	} else {
		return false, []byte(""), nil
	}
}

//创建用户数据，如果存在返回错误信息。
func (org *Organization) Put(user, name, description string) error {
	if has, _, err := org.Has(name); err != nil {
		return err
	} else if has == true {
		return fmt.Errorf("组织 %s 已经存在", name)
	} else {
		//检查用户的命名空间是否冲突
		u := new(User)
		if has, _, err := u.Has(name); err != nil {
			return err
		} else if has == true {
			return fmt.Errorf("已经存在相同的用户 %s", name)
		}

		//检查用户是否存在
		if has, _, err := u.Has(user); err != nil {
			return err
		} else if has == false {
			return fmt.Errorf("不存在用户数据")
		}

		key := utils.GeneralKey(name)

		//检查用户名合法性，参考实现标准：
		//https://github.com/docker/docker/blob/28f09f06326848f4117baf633ec9fc542108f051/registry/registry.go#L27
		validNamespace := regexp.MustCompile(`^([a-z0-9_]{4,30})$`)
		if !validNamespace.MatchString(name) {
			return fmt.Errorf("组织名必须是 4 - 30 位之间，且只能由 a-z，0-9 和 下划线组成")
		}

		org.Owner = user
		org.Name = name
		org.Description = description

		org.Updated = time.Now().Unix()
		org.Created = time.Now().Unix()

		if err := org.Save(key); err != nil {
			return err
		} else {
			//保存成功后在全局变量 #org 中保存 Key 的信息。
			if _, err := LedisDB.HSet([]byte(GetServerKeys("org")), []byte(GetObjectKey("org", name)), key); err != nil {
				return err
			}

			//向组织添加 Owner 用户
			if e := org.AddUser(name, user, ORG_OWNER); e != nil {
				return e
			}

			//向用户添加组织的数据
			if e := u.AddOrganization(user, name, ORG_MEMBER); e != nil {
				return e
			}

			return nil
		}
	}
}

//向组织添加用户，member 参数的值为 [OWNER/MEMBER] 两种
func (org *Organization) AddUser(name, user, member string) error {
	return nil
}

//从组织移除用户
func (org *Organization) RemoveUser(name, user string) error {
	return nil
}

//向组织添加镜像仓库
func (org *Organization) AddRepository(name, repository, key string) error {
	return nil
}

//从组织移除镜像仓库
func (org *Organization) RemoveRepository(name, repository string) error {
	return nil
}

//为用户@镜像仓库添加读写权限
func (org *Organization) AddPrivilege(name, user, repository, key string) error {
	return nil
}

//为用户@镜像仓库移除读写权限
func (org *Organization) RemovePrivilege(name, user, repository string) error {
	return nil
}

//循环 Org 的所有 Property ，保存数据
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
