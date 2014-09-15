package models

import (
	"github.com/dockercn/docker-bucket/utils"
	"strconv"
	"time"
)

type Job struct {
	User         string    //
	Organization string    //如果是 Organization ，存 Organization 的 ID
	Template     string    //运行的模板和 tag
	Name         string    //任务名称，为扩展 docker 命令准备。
	Title        string    //任务名称，高级版本
	YAML         string    //YAML 格式运行定义
	Description  string    //描述
	Engine       string    //运行 Template 的云服务器方面接口信息
	Actived      bool      //
	Created      time.Time //
	Updated      time.Time //
	Log          string    //
}

type Template struct {
	User         string    //
	Organization string    //如果是 Organization ，存 Organization 的 ID
	Name         string    //模板名，为 docker 扩展 template 命令使用，docker tempalte docker.cn/docker/wordpress 运行 docker 用户的 wordpress 模板
	Tag          string    //模板和 docker 使用相同的标签方案，user+name+tag 在数据表中唯一。
	Title        string    //模板名称
	Description  string    //Markdown
	Repositories string    //模板使用的 repository 集合，使用 docker.cn/docker/golang 这样的全路径，多个 repository 之间 ; 分隔
	YAML         string    // yaml 格式模板定义
	Links        string    //保存 JSON 的信息，产生 template 库的 Git 库地址
	Size         int64     //使用的所有 repository 的 size 总和
	Labels       string    //用户设定的标签，和库的 Tag 是不一样
	Icon         string    //
	Privated     bool      //私有模板
	Actived      bool      //
	Created      time.Time //
	Updated      time.Time //
	Log          string    //
}

type Repository struct {
	Id           int64
	Namespace    string    //即用户名或 Organization 的 Name
	Repository   string    //仓库名
	Organization string    //如果是 Organization ，存 Organization 的 ID
	Description  string    //保存 Markdown 格式
	JSON         string    //
	Dockerfile   string    //生产 Repository 的 Dockerfile 文件内容
	Agent        string    //docker 命令产生的 agent 信息
	Links        string    //保存 JSON 的信息，保存官方库的 Link，产生 repository 库的 Git 库地址
	Size         int64     //仓库所有 Image 的大小 byte
	Uploaded     bool      //上传完成标志
	Checksum     string    //
	Checksumed   bool      //Checksum 检查标志
	Labels       string    //用户设定的标签，和库的 Tag 是不一样
	Icon         string    //
	Privated     bool      //私有 Repository
	Actived      bool      //删除标志
	Security     bool      //是否加密
	Created      time.Time //
	Updated      time.Time //
	Log          string    //
}

type Tag struct {
	Name       string    //
	ImageId    string    //
	Repository string    //
	Created    time.Time //
	Updated    time.Time //
	Log        string    //
}

func (repo *Repository) Get(namespace string, repository string, namespaceType string) (bool, error) {
	if namespaceType == "Organization" {
		info, err := LedisDB.HGet([]byte(utils.ToString("#", namespace, "$", repository)), []byte("JSON"))
		if err != nil {
			return false, err
		} else if len(info) <= 0 {
			return false, nil
		} else {
			return true, nil
		}
	} else {
		info, err := LedisDB.HGet([]byte(utils.ToString("@", namespace, "$", repository)), []byte("JSON"))
		if err != nil {
			return false, err
		} else if len(info) <= 0 {
			return false, nil
		} else {
			return true, nil
		}
	}
}

func (repo *Repository) GetPushed(namespace string, repository string, uploaded bool, checksumed bool) (bool, error) {
	return false, nil
}

//多个UP方法可以合并
//通用方法更新Repository的信息
func (repo *Repository) UpdateRepositoryInfo(namespace, repository, namespaceType, infoKey, infoData string) (bool, error) {
	if namespaceType == "Organization" {
		_, infoErr := LedisDB.HSet([]byte(utils.ToString("#", namespace, "$", repository)), []byte(infoKey), []byte(infoData))

		if infoErr != nil {
			return false, infoErr
		}
		return true, nil
	} else {
		_, infoErr := LedisDB.HSet([]byte(utils.ToString("@", namespace, "$", repository)), []byte(infoKey), []byte(infoData))
		if infoErr != nil {
			return false, infoErr
		}
		return true, nil
	}
}

func (repo *Repository) UpdateJSON(namespace, repository, namespaceType, json string) (bool, error) {
	if namespaceType == "Organization" {
		_, jsonInfoErr := LedisDB.HSet([]byte(utils.ToString("#", namespace, "$", repository)), []byte("JSON"), []byte(json))

		if jsonInfoErr != nil {
			return false, jsonInfoErr
		}
		return true, nil
	} else {
		_, jsonInfoErr := LedisDB.HSet([]byte(utils.ToString("@", namespace, "$", repository)), []byte("JSON"), []byte(json))
		if jsonInfoErr != nil {
			return false, jsonInfoErr
		}
		return true, nil
	}
}

func (repo *Repository) UpdateUploaded(uploaded bool) (bool, error) {
	return false, nil
}

func (repo *Repository) UpdateChecksumed(checksumed bool) (bool, error) {
	return false, nil
}

func (repo *Repository) UpdateSize(size int64) (bool, error) {
	return false, nil
}

func (repo *Repository) Insert(namespace, repository, namespaceType, json string, privated bool) (bool, error) {
	if namespaceType == "Organization" {
		_, jsonInfoErr := LedisDB.HSet([]byte(utils.ToString("#", namespace, "$", repository)), []byte("JSON"), []byte(json))
		_, privatedInfoErr := LedisDB.HSet([]byte(utils.ToString("#", namespace, "$", repository)), []byte("Privated"), []byte(strconv.FormatBool(privated)))

		if jsonInfoErr != nil {
			return false, jsonInfoErr
		}
		if privatedInfoErr != nil {
			return false, privatedInfoErr
		}
		return true, nil
	} else {
		_, jsonInfoErr := LedisDB.HSet([]byte(utils.ToString("@", namespace, "$", repository)), []byte("JSON"), []byte(json))
		_, privatedInfoErr := LedisDB.HSet([]byte(utils.ToString("@", namespace, "$", repository)), []byte("Privated"), []byte(strconv.FormatBool(privated)))

		if jsonInfoErr != nil {
			return false, jsonInfoErr
		}
		if privatedInfoErr != nil {
			return false, privatedInfoErr
		}
		return true, nil
	}
}

//func (tag *Tag) Insert(name string, imageId string, repository int64) (bool, error) {
//	return false, nil
//}

func (tag *Tag) Insert(namespace, repository, namespaceType, tagName, imageId string) (bool, error) {
	if namespaceType == "Organization" {
		info, err := LedisDB.HSet([]byte(utils.ToString("#", namespace, "$", repository, "%", tagName)), []byte("ImageId"), []byte(imageId))
		if info <= 0 || err != nil {
			return false, err
		} else {
			return true, nil
		}
	} else {
		info, err := LedisDB.HSet([]byte(utils.ToString("@", namespace, "$", repository, "%", tagName)), []byte("ImageId"), []byte(imageId))
		if info <= 0 || err != nil {
			return false, err
		} else {
			return true, nil
		}
	}
}

func (tag *Tag) UpdateImageId(namespace, repository, namespaceType, tagName, imageId string) (bool, error) {
	if namespaceType == "Organization" {
		info, err := LedisDB.HSet([]byte(utils.ToString("#", namespace, "$", repository, "%", tagName)), []byte("ImageId"), []byte(imageId))
		if info <= 0 || err != nil {
			return false, err
		} else {
			return true, nil
		}
	} else {
		info, err := LedisDB.HSet([]byte(utils.ToString("@", namespace, "$", repository, "%", tagName)), []byte("ImageId"), []byte(imageId))
		if info <= 0 || err != nil {
			return false, err
		} else {
			return true, nil
		}
	}
}

func (tag *Tag) Get(namespace, repository, namespaceType, tagName string) (bool, error) {
	if namespaceType == "Organization" {
		info, err := LedisDB.HGet([]byte(utils.ToString("#", namespace, "$", repository, "%", tagName)), []byte("Name"))
		if len(info) <= 0 || err != nil {
			return false, err
		} else {
			return true, nil
		}
	} else {
		info, err := LedisDB.HGet([]byte(utils.ToString("@", namespace, "$", repository, "%", tagName)), []byte("Name"))
		if len(info) <= 0 || err != nil {
			return false, err
		} else {
			return true, nil
		}
	}
}

func (tag *Tag) GetImagesJSON(repository int64) ([]byte, error) {
	return []byte(""), nil
}
