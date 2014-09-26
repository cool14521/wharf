package models

import (
	"fmt"
	"reflect"
	"time"

	"github.com/dockercn/docker-bucket/utils"
)

type Job struct {
	User         string //
	Organization string //如果是 Organization ，存 Organization 的 ID
	Template     string //运行的模板和 tag
	Name         string //任务名称，为扩展 docker 命令准备。
	Title        string //任务名称，高级版本
	YAML         string //YAML 格式运行定义
	Description  string //描述
	Engine       string //运行 Template 的云服务器方面接口信息
	Actived      bool   //
	Created      int64  //
	Updated      int64  //
}

type Template struct {
	User         string //
	Organization string //如果是 Organization ，存 Organization 的 ID
	Name         string //模板名，为 docker 扩展 template 命令使用，docker tempalte docker.cn/docker/wordpress 运行 docker 用户的 wordpress 模板
	Tag          string //模板和 docker 使用相同的标签方案，user+name+tag 在数据表中唯一。
	Title        string //模板名称
	Description  string //Markdown
	Repositories string //模板使用的 repository 集合，使用 docker.cn/docker/golang 这样的全路径，多个 repository 之间 ; 分隔
	YAML         string // yaml 格式模板定义
	Links        string //保存 JSON 的信息，产生 template 库的 Git 库地址
	Size         int64  //使用的所有 repository 的 size 总和
	Labels       string //用户设定的标签，和库的 Tag 是不一样
	Icon         string //
	Privated     bool   //私有模板
	Actived      bool   //
	Created      int64  //
	Updated      int64  //
}

type Repository struct {
	Username     string //即用户名或 Organization 的 Name
	Repository   string //仓库名
	Organization string //如果是 Organization ，存 Organization 的 Key
	Description  string //保存 Markdown 格式
	JSON         string //Docker 客户端上传的 Images 信息，JSON 格式。
	Data         string //包含的 Image 信息集合，由 Image 的 ID 组成。
	Dockerfile   string //生产 Repository 的 Dockerfile 文件内容
	Agent        string //docker 命令产生的 agent 信息
	Links        string //保存 JSON 的信息，保存官方库的 Link，产生 repository 库的 Git 库地址
	Size         int64  //仓库所有 Image 的大小 byte
	Uploaded     bool   //上传完成标志
	Checksum     string //
	Checksumed   bool   //Checksum 检查标志
	Labels       string //用户设定的标签，和库的 Tag 是不一样
	Tags         string //
	Icon         string //
	Sign         string //
	Privated     bool   //私有 Repository
	Clear        string //对 Repository 进行了杀毒，杀毒的结果和 status 等信息以 JSON 格式保存
	Cleared      bool   //对 Repository 是否进行了杀毒处理
	Actived      bool   //删除标志
	Encrypted    bool   //是否加密
	Token        string //
	Created      int64  //
	Updated      int64  //
}

func (repo *Repository) Get(username, repository, organization, sign string, active bool) (bool, []byte, error) {
	var keys [][]byte

	//查询 User 是否存在，如果不存在 User 的数据，直接返回错误信息
	user := new(User)
	if has, err := user.Has(username); err != nil {
		return false, []byte(""), err
	} else if has == true && len(organization) == 0 {
		//个人用户的 repository 条件，存在用户数据且组织数据为空
		if len(sign) == 0 {
			keys = append(keys, []byte(fmt.Sprintf("%s%s+", GetObjectKey("user", username), GetObjectKey("repo", repository))))
			keys = append(keys, []byte(fmt.Sprintf("%s%s-", GetObjectKey("user", username), GetObjectKey("repo", repository))))
		} else if len(sign) > 0 {
			keys = append(keys, []byte(fmt.Sprintf("%s%s-?%s", GetObjectKey("user", username), GetObjectKey("repo", repository), sign)))
		}
	} else if has == true && len(organization) > 0 {
		//组织用户的 repository 条件，存在用户数据且组织数据不为空
		//查询组织数据是否存在
		org := new(Organization)
		if h, e := org.Has(organization); e != nil {
			return false, []byte(""), e
		} else if h == false {
			return false, []byte(""), fmt.Errorf("没有找到 %s 组织的数据", organization)
		} else if h == true {
			if len(sign) == 0 {
				keys = append(keys, []byte(fmt.Sprintf("%s%s+", GetObjectKey("org", organization), GetObjectKey("repo", repository))))
				keys = append(keys, []byte(fmt.Sprintf("%s%s-", GetObjectKey("org", organization), GetObjectKey("repo", repository))))
			} else if len(sign) > 0 {
				keys = append(keys, []byte(fmt.Sprintf("%s%s-?%s", GetObjectKey("org", organization), GetObjectKey("repo", repository), sign)))
			}
		}

		for _, value := range keys {
			if exist, err := LedisDB.Exists(value); err != nil {
				return false, []byte(""), err
			} else if exist > 0 {
				var key []byte

				if key, err = LedisDB.Get([]byte(value)); err != nil {
					return false, key, err
				}

				if result, err := LedisDB.HGet(key, []byte("Actived")); err != nil {
					return false, key, err
				} else if utils.BytesToBool(result) == active {
					return true, key, nil
				}
			}
		}
	}

	return false, []byte(""), nil
}

func (repo *Repository) Add(username, repository, organization, sign, json string) error {
	if has, key, err := repo.Get(username, repository, organization, sign, true); err != nil {
		return err
	} else if has == true {
		//修改数据
		repo.Username = username
		repo.Repository = repository
		repo.Organization = organization
		repo.JSON = json
		repo.Actived = true
		repo.Updated = time.Now().Unix()

		if len(sign) > 0 {
			repo.Sign = sign
			repo.Encrypted = true
		}

		if e := repo.Save(key); e != nil {
			return e
		}

	} else if has == false {
		//第一次创建数据
		key = utils.GeneralKey(fmt.Sprintf("%s%s+", GetObjectKey("user", username), GetObjectKey("repo", repository)))

		repo.Username = username
		repo.Repository = repository
		repo.Organization = organization
		repo.JSON = json

		repo.Updated = time.Now().Unix()
		repo.Created = time.Now().Unix()

		repo.Privated = false
		repo.Checksumed = false
		repo.Uploaded = false
		repo.Cleared = false
		repo.Encrypted = false
		repo.Actived = true

		if len(sign) > 0 {
			repo.Sign = sign
			repo.Encrypted = true
		}

		if e := repo.Save(key); e != nil {
			return e
		}
	}

	return nil
}

func (repo *Repository) Save(key []byte) error {
	s := reflect.TypeOf(repo).Elem()

	//循环处理 Struct 的每一个 Field
	for i := 0; i < s.NumField(); i++ {
		//获取 Field 的 Value
		value := reflect.ValueOf(repo).Elem().Field(s.Field(i).Index[0])

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

func (repo *Repository) SetAgent(namespace, repository, organization, agent string) error {
	return nil
}

func (repo *Repository) SetToken(namespace, repository, organization, token string) error {
	return nil
}

func (repo *Repository) SetTag(namespace, repository, organization, tag, image string) error {
	return nil
}

func (repo *Repository) SetJSON(namespace, repository, organization, json string) error {
	return nil
}

func (repo *Repository) Setloaded(uploaded bool) error {
	return nil
}

func (repo *Repository) SetChecksumed(checksumed bool) error {
	return nil
}

func (repo *Repository) SetSize(size int64) error {
	return nil
}
