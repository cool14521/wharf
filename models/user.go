package models

import (
	"encoding/json"
	"fmt"
	"github.com/dockercn/wharf/utils"
	"regexp"
	"sort"
	"strconv"
	"time"
)

const (
	ORG_MEMBER = "M"
	ORG_OWNER  = "O"
)

type User struct {
	Username      string `json:"usernamne"`     //
	Password      string `json:"password"`      //
	Repositories  string `json:"repositories"`  //用户的所有 Respository
	Organizations string `json:"organizaitons"` //用户所属的所有组织
	Email         string `json:"email"`         //Email 可以更换，全局唯一
	Fullname      string `json:"fullname"`      //
	Company       string `json:"company"`       //
	Location      string `json:"location"`      //
	Mobile        string `json:"mobile"`        //
	URL           string `json:"url"`           //
	Gravatar      string `json:"gravatar"`      //如果是邮件地址使用 gravatar.org 的 API 显示头像，如果是上传的用户显示头像的地址。
	Created       int64  `json:"created"`       //
	Updated       int64  `json:"updated"`       //
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
				return true, key, fmt.Errorf("已经存在了 Key，但是用户名不相同 %s ", string(name))
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
			return fmt.Errorf("已经存在相同名称的组织 %s", username)
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
		validEmail := regexp.MustCompile("[\\w!#$%&'*+/=?^_`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[a-zA-Z0-9](?:[\\w-]*[\\w])?")
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
		if err := Save(user, key); err != nil {
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

//根据用户名更新用户数据
func (user *User) Update(u map[string]interface{}) (bool, error) {
	//获取用户唯一ID
	if _, key, err := user.Has(user.Username); err != nil {
		return false, err
	} else {
		//遍历map中元素，对存在元素进行更新
		for attr, value := range u {
			if result, ok := value.(string); ok {
				switch attr {
				case "name":

				case "fullname":
					user.Fullname = result
				case "url":
					user.URL = result
				case "gravatar":
					user.Gravatar = result
				case "company":
					user.Company = result
				case "mobile":
					user.Mobile = result
				case "email":
					user.Email = result
				case "newPassword":
					user.Password = result
				}
			}

		}
		//保存 User 对象的数据
		if err := Save(user, key); err != nil {
			return false, err
		}
		return true, nil
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
			//对user进行赋值
			attrs, _ := LedisDB.HKeys(key)
			for _, attr := range attrs {
				attr2string := string(attr)
				switch attr2string {
				case "Company":
					company, _ := LedisDB.HGet(key, []byte("Company"))
					user.Company = string(company)
				case "Created":
					created, _ := LedisDB.HGet(key, []byte("Created"))
					user.Created, _ = strconv.ParseInt(string(created), 0, 64)
				case "Email":
					email, _ := LedisDB.HGet(key, []byte("Email"))
					user.Email = string(email)
				case "Fullname":
					fullName, _ := LedisDB.HGet(key, []byte("Fullname"))
					user.Fullname = string(fullName)
				case "Gravatar":
					gravatar, _ := LedisDB.HGet(key, []byte("Gravatar"))
					user.Gravatar = string(gravatar)
				case "Location":
					location, _ := LedisDB.HGet(key, []byte("Location"))
					user.Location = string(location)
				case "Mobile":
					mobile, _ := LedisDB.HGet(key, []byte("Mobile"))
					user.Mobile = string(mobile)
				case "Organizations":
					organizations, _ := LedisDB.HGet(key, []byte("Organizations"))
					user.Organizations = string(organizations)
				case "Repositories":
					repositories, _ := LedisDB.HGet(key, []byte("Repositories"))
					user.Repositories = string(repositories)
				case "URL":
					url, _ := LedisDB.HGet(key, []byte("URL"))
					user.URL = string(url)
				case "Updated":
					updated, _ := LedisDB.HGet(key, []byte("Updated"))
					user.Updated, _ = strconv.ParseInt(string(updated), 0, 64)
				case "Username":
					username, _ := LedisDB.HGet(key, []byte("Username"))
					user.Username = string(username)
				case "Password":
					password, _ := LedisDB.HGet(key, []byte("Password"))
					user.Password = string(password)
				}
			}
			return true, nil
		}
	} else {
		//没有用户的 Key 存在
		return false, fmt.Errorf("不存在 %s 的用户数据", username)
	}
}

//重置用户的密码
func (user *User) ResetPasswd(username, password string) error {
	//检查用户的 Key 是否存在
	if has, key, err := user.Has(username); err != nil {
		return err
	} else if has == true {
		user.Password = password
		user.Updated = time.Now().Unix()

		if err := Save(user, key); err != nil {
			return err
		}
	} else {
		//没有用户的 Key 存在
		return fmt.Errorf("不存在 %s 的用户数据", username)
	}

	return nil
}

//用户创建镜像仓库后，在 user 的 Repositories 字段保存相应的记录
//repository 字段是 Repository 的全局 Key
func (user *User) AddRepository(username, repository, key string) error {
	var (
		u   []byte
		has bool
		err error
	)

	r := make(map[string]string, 0)

	//检查用户的 Key 是否存在
	if has, u, err = user.Has(username); err != nil {
		return err
	} else if has == true {
		if repo, err := LedisDB.HGet(u, []byte("Repositories")); err != nil {
			return err
		} else if repo != nil {
			if e := json.Unmarshal(repo, &r); e != nil {
				return nil
			}
			if value, exist := r[repository]; exist == true && value == key {
				return fmt.Errorf("已经存在了镜像仓库数据")
			}
		}
	} else if has == false {
		return fmt.Errorf("不存在 %s 的用户数据", username)
	}

	//在 Map 中增加 repository 记录
	r[repository] = key
	//JSON
	repo, _ := json.Marshal(r)

	user.Repositories = string(repo)
	user.Updated = time.Now().Unix()

	if err := Save(user, u); err != nil {
		return err
	}

	return nil
}

//用户删除镜像仓库后，在 user 的 Repositories 中删除相应的记录
func (user *User) RemoveRepository(username, repository string) error {
	var (
		u   []byte
		has bool
		err error
	)

	r := make(map[string]string, 0)

	//检查用户的 Key 是否存在
	if has, u, err = user.Has(username); err != nil {
		return err
	} else if has == true {
		if repo, err := LedisDB.HGet(u, []byte("Repositories")); err != nil {
			return err
		} else if repo != nil {
			if e := json.Unmarshal(repo, &r); e != nil {
				return nil
			}
			if _, exist := r[repository]; exist == false {
				return fmt.Errorf("不存在要删除的镜像仓库数据")
			}
		}
	} else if has == false {
		return fmt.Errorf("不存在 %s 的用户数据", username)
	}

	//在 Map 中删除 repository 记录
	delete(r, repository)
	//JSON
	repo, _ := json.Marshal(r)

	user.Repositories = string(repo)
	user.Updated = time.Now().Unix()

	if err := Save(user, u); err != nil {
		return err
	}

	return nil
}

//向用户添加 Organization 数据
//member 的值为：	ORG_MEMBER 或 ORG_OWNER
func (user *User) AddOrganization(username, org, member string) error {
	var (
		u   []byte
		has bool
		err error
	)

	o := make(map[string]string, 0)

	if has, u, err = user.Has(username); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("不存在 %s 的用户数据", username)
	} else if has == true {
		if organization, err := LedisDB.HGet(u, []byte("Organizations")); err != nil {
			return nil
		} else if organization != nil {
			if e := json.Unmarshal(organization, &o); e != nil {
				return e
			}

			if value, exist := o[org]; exist == true && value == member {
				return fmt.Errorf("已经存在了组织的数据")
			}
		}
	}

	o[org] = member

	os, _ := json.Marshal(o)

	user.Organizations = string(os)
	user.Updated = time.Now().Unix()

	if err := Save(user, u); err != nil {
		return err
	}

	return nil
}

//从用户中删除 Organization 数据
func (user *User) RemoveOrganization(username, org string) error {
	var (
		u   []byte
		has bool
		err error
	)

	o := make(map[string]string, 0)

	if has, u, err = user.Has(username); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("不存在 %s 的用户数据", username)
	} else if has == true {
		if organization, err := LedisDB.HGet(u, []byte("Organizations")); err != nil {
			return nil
		} else if organization != nil {
			if e := json.Unmarshal(organization, &o); e != nil {
				return e
			}
			if _, exist := o[org]; exist == false {
				return fmt.Errorf("不存在要移除的用户数据")
			}
		}
	}

	delete(o, org)

	os, _ := json.Marshal(o)

	user.Organizations = string(os)
	user.Updated = time.Now().Unix()

	if err := Save(user, u); err != nil {
		return err
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
		if n, err := LedisDB.HGet(key, []byte("Name")); err != nil {
			return false, []byte(""), err
		} else if n != nil {
			if string(n) != name {
				return true, key, fmt.Errorf("已经存在了 Key，但是组织名称不相同 %s", string(n))
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
			return fmt.Errorf("已经存在相同名称的用户 %s", name)
		}

		//检查用户是否存在
		if has, _, err := u.Has(user); err != nil {
			return err
		} else if has == false {
			return fmt.Errorf("不存在用户的数据 %s", user)
		}

		key := utils.GeneralKey(name)

		//检查组织名合法性，参考实现标准：
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

		if err := Save(org, key); err != nil {
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
	var (
		o   []byte
		has bool
		err error
	)

	users := make(map[string]string, 0)

	if has, o, err = org.Has(name); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("不存在组织的数据 %s", name)
	}

	u := new(User)
	if has, _, err = u.Has(user); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("不存在用户的数据 %S", user)
	}

	if us, err := LedisDB.HGet(o, []byte("Users")); err != nil {
		return nil
	} else if us != nil {
		if e := json.Unmarshal(us, &users); e != nil {
			return e
		}

		if value, exist := users[user]; exist == true && value == member {
			return fmt.Errorf("已经存在了用户的数据")
		}
	}

	users[user] = member

	us, _ := json.Marshal(users)

	org.Users = string(us)
	org.Updated = time.Now().Unix()

	if err = Save(org, o); err != nil {
		return err
	}

	return nil
}

//从组织移除用户
func (org *Organization) RemoveUser(name, user string) error {
	var (
		o   []byte
		has bool
		err error
	)

	users := make(map[string]string, 0)

	if has, o, err = org.Has(name); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("不存在组织的数据 %s", name)
	}

	u := new(User)
	if has, _, err = u.Has(user); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("不存在用户的数据 %S", user)
	}

	if us, err := LedisDB.HGet(o, []byte("Users")); err != nil {
		return err
	} else if us != nil {
		if e := json.Unmarshal(us, &users); e != nil {
			return e
		}

		if _, exist := users[user]; exist == false {
			return fmt.Errorf("在组织中不存在要移除的用户数据")
		}
	}

	delete(users, user)

	us, _ := json.Marshal(users)

	org.Users = string(us)
	org.Updated = time.Now().Unix()

	if err = Save(org, o); err != nil {
		return err
	}

	return nil
}

//向组织添加镜像仓库
func (org *Organization) AddRepository(name, repository, key string) error {
	var (
		o   []byte
		has bool
		err error
	)

	repos := make(map[string]string, 0)

	if has, o, err = org.Has(name); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("不存在组织的数据 %s", name)
	}

	repo := new(Repository)
	if has, err = repo.Has(repository); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("不存在镜像仓库的数据: %s", repository)
	}

	if r, err := LedisDB.HGet(o, []byte("Repositories")); err != nil {
		return err
	} else if r != nil {
		if e := json.Unmarshal(r, &repos); e != nil {
			return e
		}

		if value, exist := repos[repository]; exist == true && value == key {
			return fmt.Errorf("在组织中已经存在要添加的镜像仓库数据")
		}
	}

	repos[repository] = key

	rs, _ := json.Marshal(repos)

	org.Repositories = string(rs)
	org.Updated = time.Now().Unix()

	if err = Save(org, o); err != nil {
		return err
	}

	return nil
}

//从组织移除镜像仓库
func (org *Organization) RemoveRepository(name, repository string) error {
	var (
		o   []byte
		has bool
		err error
	)

	repos := make(map[string]string, 0)

	if has, o, err = org.Has(name); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("不存在组织的数据 %s", name)
	}

	repo := new(Repository)
	if has, err = repo.Has(repository); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("不存在镜像仓库的数据: %s", repository)
	}

	if r, err := LedisDB.HGet(o, []byte("Repositories")); err != nil {
		return err
	} else if r != nil {
		if e := json.Unmarshal(r, &repos); e != nil {
			return e
		}

		if _, exist := repos[repository]; exist == false {
			return fmt.Errorf("在组织中不存在要删除的仓库数据")
		}
	}

	delete(repos, repository)

	rs, _ := json.Marshal(repos)

	org.Repositories = string(rs)
	org.Updated = time.Now().Unix()

	if err = Save(org, o); err != nil {
		return err
	}

	return nil
}

//为用户@镜像仓库添加读写权限
func (org *Organization) AddPrivilege(name, user, repository string) error {
	var (
		o   []byte
		has bool
		err error
	)

	privileges := make(map[string][]string, 0)

	if has, o, err = org.Has(name); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("不存在组织的数据 %s", name)
	}

	repo := new(Repository)
	if has, err = repo.Has(repository); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("不存在镜像仓库的数据: %s", repository)
	}

	u := new(User)
	if has, _, err = u.Has(user); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("不存在用户的数据 %S", user)
	}

	if p, err := LedisDB.HGet(o, []byte("Privileges")); err != nil {
		return err
	} else if p != nil {
		if e := json.Unmarshal(p, &privileges); e != nil {
			return e
		}

		if value, exist := privileges[repository]; exist == true {
			if sort.SearchStrings(value, user) > -1 {
				return fmt.Errorf(" %s 镜像仓库已经存在了 % 的用户权限记录", repository, user)
			}

			value = append(value, user)
			privileges[repository] = value

			ps, _ := json.Marshal(privileges)

			org.Privileges = string(ps)
			org.Updated = time.Now().Unix()

			if err = Save(org, o); err != nil {
				return err
			}

		}
	}

	return nil
}

//为用户@镜像仓库移除读写权限
func (org *Organization) RemovePrivilege(name, user, repository string) error {
	var (
		o   []byte
		has bool
		err error
	)

	privileges := make(map[string][]string, 0)

	if has, o, err = org.Has(name); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("不存在组织的数据 %s", name)
	}

	repo := new(Repository)
	if has, err = repo.Has(repository); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("不存在镜像仓库的数据: %s", repository)
	}

	u := new(User)
	if has, _, err = u.Has(user); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("不存在用户的数据 %S", user)
	}

	if p, err := LedisDB.HGet(o, []byte("Privileges")); err != nil {
		return err
	} else if p != nil {
		if e := json.Unmarshal(p, &privileges); e != nil {
			return e
		}

		if value, exist := privileges[repository]; exist == true {
			var i int
			if i = sort.SearchStrings(value, user); i == -1 {
				return fmt.Errorf(" %s 镜像仓库已经不存在了 % 的用户权限记录", repository, user)
			}

			value = append(value[:i], value[i+1:]...)
			privileges[repository] = value

			ps, _ := json.Marshal(privileges)

			org.Privileges = string(ps)
			org.Updated = time.Now().Unix()

			if err = Save(org, o); err != nil {
				return err
			}

		}
	}

	return nil
}
