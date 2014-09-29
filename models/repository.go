package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/astaxie/beego"

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
	Encrypted    bool   //是否加密
	Created      int64  //
	Updated      int64  //
}

type Tag struct {
	Name    string //
	ImageId string //
	Sign    string //
}

func (repo *Repository) Get(username, repository, organization, sign string) (bool, []byte, error) {
	var keys [][]byte

	//查询 User 是否存在，如果不存在 User 的数据，直接返回错误信息。
	//系统先创用户，在由用户创建组织，所以搜索到用户就直接报错。
	user := new(User)
	if has, err := user.Has(username); err != nil {
		return false, []byte(""), err
	} else if has == true && len(organization) == 0 {
		//存在用户数据且组织数据为空
		if len(sign) == 0 {
			//非加密数据库 Key 规则：
			//公开库：@username$repository+
			//私有库：@username$repository-
			keys = append(keys, []byte(fmt.Sprintf("%s%s+", GetObjectKey("user", username), GetObjectKey("repo", repository))))
			keys = append(keys, []byte(fmt.Sprintf("%s%s-", GetObjectKey("user", username), GetObjectKey("repo", repository))))
		} else if len(sign) > 0 {
			//加密数据库必须为私有库：@username$repository-?sign
			keys = append(keys, []byte(fmt.Sprintf("%s%s-?%s", GetObjectKey("user", username), GetObjectKey("repo", repository), sign)))
		}
	} else if has == true && len(organization) > 0 {
		//存在用户数据且组织数据不为空
		//查询组织数据是否存在
		org := new(Organization)
		if h, e := org.Has(organization); e != nil {
			return false, []byte(""), e
		} else if h == false {
			return false, []byte(""), fmt.Errorf("没有找到 %s 组织的数据", organization)
		} else if h == true {
			if len(sign) == 0 {
				//非加密数据库 Key 规则：
				//公开库：@org$repository+
				//私有库：@org$repository-
				keys = append(keys, []byte(fmt.Sprintf("%s%s+", GetObjectKey("org", organization), GetObjectKey("repo", repository))))
				keys = append(keys, []byte(fmt.Sprintf("%s%s-", GetObjectKey("org", organization), GetObjectKey("repo", repository))))
			} else if len(sign) > 0 {
				//加密数据库必须为私有库：@org$repository-?sign
				keys = append(keys, []byte(fmt.Sprintf("%s%s-?%s", GetObjectKey("org", organization), GetObjectKey("repo", repository), sign)))
			}
		}
	}

	//循环 keys 判断是否存在
	for _, value := range keys {
		if exist, err := LedisDB.Exists(value); err != nil {
			return false, []byte(""), err
		} else if exist > 0 {
			var key []byte

			if key, err = LedisDB.Get([]byte(value)); err != nil {
				return false, key, err
			}

			return true, key, nil
		}
	}

	return false, []byte(""), nil
}

func (repo *Repository) PutJSON(username, repository, organization, sign, json string) error {
	if has, key, err := repo.Get(username, repository, organization, sign); err != nil {
		return err
	} else if has == true {
		//修改数据
		repo.Username = username
		repo.Repository = repository
		repo.Organization = organization
		repo.JSON = json
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

		if len(sign) > 0 {
			repo.Sign = sign
			repo.Encrypted = true
		}

		if e := repo.Save(key); e != nil {
			return e
		} else {
			if len(organization) == 0 {
				//没有 org 为空，根据 sign 的值判断是否为私有
				if len(sign) == 0 {
					if e := LedisDB.Set([]byte(fmt.Sprintf("%s%s+", GetObjectKey("user", username), GetObjectKey("repo", repository))), key); e != nil {
						return e
					}
				} else {
					if e := LedisDB.Set([]byte(fmt.Sprintf("%s%s-?%s", GetObjectKey("user", username), GetObjectKey("repo", repository), sign)), key); e != nil {
						return e
					}
				}
			} else {
				//没有 org 不为空，根据 sign 的值判断是否为私有
				if len(sign) == 0 {
					if e := LedisDB.Set([]byte(fmt.Sprintf("%s%s+", GetObjectKey("org", organization), GetObjectKey("repo", repository))), key); e != nil {
						return e
					}
				} else {
					if e := LedisDB.Set([]byte(fmt.Sprintf("%s%s-?%s", GetObjectKey("org", organization), GetObjectKey("repo", repository), sign)), key); e != nil {
						return e
					}
				}
			}
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

func (repo *Repository) PutAgent(username, repository, organization, sign, agent string) error {
	if has, key, err := repo.Get(username, repository, organization, sign); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("没有在数据库中查询到要更新的 Repository 数据")
	} else if has == true {
		repo.Agent = agent

		if e := repo.Save(key); e != nil {
			return e
		}
	}

	return nil
}

func (repo *Repository) PutTag(username, repository, organization, sign, tag, imageId string) error {
	if has, key, err := repo.Get(username, repository, organization, sign); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("没有在数据库中查询到要更新的 Repository 数据")
	} else if has == true {
		//Tags 字段保存 Tag 结构的数组再 JSON 编码
		tags := make([]Tag, 0)

		//获取记录中得 Tags 数据，解压到 tags 数组中
		if ts, err := LedisDB.HGet(key, []byte("Tags")); err != nil {
			return err
		} else if ts != nil {
			beego.Debug(fmt.Sprintf("[Tags] %s", string(ts)))
			if err := json.Unmarshal(ts, &tags); err != nil {
				return err
			}
		}

		updated := false

		//循环数组，如果已经存在了 Tag 标签，更新相应的数据
		for _, t := range tags {
			if t.Name == tag {
				t.ImageId = imageId
				t.Sign = sign
				updated = true
			}
		}

		if updated == false {
			t := new(Tag)
			t.Name = tag
			t.ImageId = imageId
			t.Sign = sign

			tags = append(tags, *t)
		}

		tagJSON, _ := json.Marshal(tags)

		repo.Tags = string(tagJSON)

		if e := repo.Save(key); e != nil {
			return e
		}
	}

	return nil
}

func (repo *Repository) PutUploaded(username, repository, organization, sign string, uploaded bool) error {
	if has, key, err := repo.Get(username, repository, organization, sign); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("没有在数据库中查询到 Repository 数据")
	} else if has == true {
		//TODO 循环检查所有的 Image 是不是 Uploaded
		repo.Uploaded = uploaded

		if e := repo.Save(key); e != nil {
			return e
		}
	}

	return nil
}

func (repo *Repository) PutChecksumed(username, repository, organization, sign string, checksumed bool) error {
	if has, key, err := repo.Get(username, repository, organization, sign); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("没有在数据库中查询到 Repository 数据")
	} else if has == true {
		//TODO 循环检查所有的 Image 是不是 Checksumed
		repo.Checksumed = checksumed

		if e := repo.Save(key); e != nil {
			return e
		}
	}

	return nil
}

func (repo *Repository) PutSize(username, repository, organization, sign string) error {
	//TODO 根据所有的 Image 的总和更新 Repository 的 Size 属性
	return nil
}

func (repo *Repository) GetJSON(username, repository, organization, sign string, uploaded, checksumed bool) ([]byte, error) {
	if has, key, err := repo.Get(username, repository, organization, sign); err != nil {
		return []byte(""), err
	} else if has == false {
		return []byte(""), fmt.Errorf("没有在数据库中查询到 Repository 数据")
	} else if has == true {
		if results, e := LedisDB.HMget(key, []byte("CheckSumed"), []byte("Uploaded"), []byte("JSON")); e != nil {
			return []byte(""), e
		} else {
			checksum := results[0]
			upload := results[1]
			json := results[2]

			if utils.BytesToBool(checksum) != checksumed {
				return []byte(""), fmt.Errorf("没有找到 Image 的数据")
			}

			if utils.BytesToBool(upload) != uploaded {
				return []byte(""), fmt.Errorf("没有找到 Image 的数据")
			}

			return json, nil
		}

	}

	return []byte(""), nil
}

func (repo *Repository) GetTags(username, repository, organization, sign string, uploaded, checksumed bool) ([]byte, error) {
	if has, key, err := repo.Get(username, repository, organization, sign); err != nil {
		return []byte(""), err
	} else if has == false {
		return []byte(""), fmt.Errorf("没有在数据库中查询到 Repository 数据")
	} else if has == true {
		if results, e := LedisDB.HMget(key, []byte("CheckSumed"), []byte("Uploaded"), []byte("Tags")); e != nil {
			return []byte(""), e
		} else {
			checksum := results[0]
			upload := results[1]
			tagsJSON := results[2]

			if utils.BytesToBool(checksum) != checksumed {
				return []byte(""), fmt.Errorf("没有找到 Image 的数据")
			}

			if utils.BytesToBool(upload) != uploaded {
				return []byte(""), fmt.Errorf("没有找到 Image 的数据")
			}

			//循环 JSON 对象的值，生成返回的数据
			results := make(map[string]string)
			tags := make([]Tag, 0)

			if err := json.Unmarshal(tagsJSON, &tags); err != nil {
				return []byte(""), err
			}

			for _, tag := range tags {
				results[tag.Name] = tag.ImageId
			}

			result, _ := json.Marshal(results)

			return result, nil
		}
	}
	return []byte(""), nil
}
