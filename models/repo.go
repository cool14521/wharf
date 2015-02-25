package models

import (
  "fmt"
  "time"

  "github.com/dockercn/wharf/utils"
)

type Repository struct {
  UUID          string   `json:"UUID"`          //全局唯一的索引, LedisDB 中 Repository List 保存全局所有的仓库名列表信息。 LedisDB 独立保存每个 Repository 信息到一个HASH，名字为 {UUID}
  Repository    string   `json:"repository"`    //仓库名称 全局唯一，不可修改
  Namespace     string   `json:"namespace"`     //仓库所有者的名字
  NamespaceType bool     `json:"namespacetype"` // false 为普通用户，true 为组织
  Organization  string   `json:"organization"`  //如果仓库属于一个 team，那么在此记录 team 所属组织
  Tags          []string `json:"tags"`          //保存此仓库所有 tag 的对应 UUID
  Starts        []string `json:"starts"`        //保存此仓库 Start 的 UUID 列表
  Comments      []string `json:"comments"`      //保存此仓库 Comment 的对应 UUID 列表
  Short         string   `json:"short"`         //保存此仓库的的短描述
  Description   string   `json:"description"`   //保存 Markdown 格式
  JSON          string   `json:"json"`          //Docker 客户端上传的 Images 信息，JSON 格式。
  Dockerfile    string   `json:"dockerfile"`    //生产 Repository 的 Dockerfile 文件内容
  Agent         string   `json:"agent"`         //docker 命令产生的 agent 信息
  Links         string   `json:"links"`         //保存 JSON 的信息，保存官方库的 Link，产生 repository 库的 Git 库地址
  Size          int64    `json:"size"`          //仓库所有 Image 的大小 byte
  Download      int64    `json:"download"`      //下载次数
  Uploaded      bool     `json:"uploaded"`      //上传完成标志
  Checksum      string   `json:"checksum"`      //
  Checksumed    bool     `json:"checksumed"`    //Checksum 检查标志
  Icon          string   `json:"icon"`          //
  Sign          string   `json:"sign"`          //
  Privated      bool     `json:"privated"`      //私有 Repository
  Clear         string   `json:"clear"`         //对 Repository 进行了杀毒，杀毒的结果和 status 等信息以 JSON 格式保存
  Cleared       bool     `json:"cleared"`       //对 Repository 是否进行了杀毒处理
  Encrypted     bool     `json:"encrypted"`     //是否加密
  Created       int64    `json:"created"`       //
  Updated       int64    `json:"updated"`       //
  Memo          string   `json:"memo"`          //
}

type Star struct {
  UUID       string `json:"UUID"`       //全局唯一的索引
  User       string `json:"user"`       //用户 UUID，代表哪个用户加的星
  Repository string `json:"repository"` //仓库 UUID，代表给哪个仓库加的星
  Time       int64  `json:"time"`       //代表加星的时间
  Memo       string `json:"memo"`       //
}

type Comment struct {
  UUID       string `json:"UUID"`       //全局唯一的索引
  Comment    string `json:"comment"`    //评论的内容 markdown 格式保存
  User       string `json:"user"`       //用户 UUID，代表哪个用户进行的评论
  Repository string `json:"repository"` //仓库 UUID，代表评论的哪个仓库
  Time       int64  `json:"time"`       //代表评论的时间
  Memo       string `json:"memo"`       //
}

type Privilege struct {
  UUID       string `json:"UUID"`       //全局唯一的索引
  Privilege  bool   `json:"privilege"`  //true 为读写，false为只读
  Team       string `json:"team"`       //此权限所属 Team 的 UUID
  Repository string `json:"repository"` //此权限对应的仓库 UUID
  Memo       string `json:"memo"`       //
}

type Tag struct {
  UUID       string `json:"uuid"`       //
  Name       string `json:"name"`       //
  ImageId    string `json:"imageid"`    //
  Namespace  string `json:"namespace"`  //
  Repository string `json:"repository"` //
  Sign       string `json:"sign"`       //
  Memo       string `json:"memo"`       //
}

func (r *Repository) Has(namespace, repository string) (bool, []byte, error) {

  UUID, err := GetUUID("repository", fmt.Sprintf("%s:%s", namespace, repository))

  if err != nil {
    return false, nil, err
  }

  if len(UUID) <= 0 {
    return false, nil, nil
  }
  err = Get(r, UUID)

  return true, UUID, err
}

func (r *Repository) Save() error {
  if err := Save(r, []byte(r.UUID)); err != nil {
    return err
  }

  if _, err := LedisDB.HSet([]byte(GLOBAL_REPOSITORY_INDEX), []byte(fmt.Sprintf("%s:%s", r.Namespace, r.Repository)), []byte(r.UUID)); err != nil {
    return err
  }

  return nil
}

func (r *Repository) Remove() error {
  if _, err := LedisDB.HSet([]byte(fmt.Sprintf("%s_remove", GLOBAL_REPOSITORY_INDEX)), []byte(fmt.Sprintf("%s:%s", r.Namespace, r.Repository)), []byte(r.UUID)); err != nil {
    return err
  }

  if _, err := LedisDB.HDel([]byte(GLOBAL_REPOSITORY_INDEX), []byte(fmt.Sprintf("%s:%s", r.Namespace, r.Repository))); err != nil {
    return err
  }

  return nil
}

func (r *Repository) Put(namespace, repository, json, agent string) error {
  if has, _, err := r.Has(namespace, repository); err != nil {
    return err
  } else if has == false {
    r.UUID = string(utils.GeneralKey(fmt.Sprintf("%s:%s", namespace, repository)))
    r.Created = time.Now().Unix()
  }

  r.Namespace, r.Repository, r.JSON, agent = namespace, repository, json, agent

  r.Updated = time.Now().Unix()
  r.Checksumed = false
  r.Uploaded = false

  if err := r.Save(); err != nil {
    return err
  }

  return nil
}

func (r *Repository) PutTag(imageId, namespace, repository, tag string) error {
  if has, _, err := r.Has(namespace, repository); err != nil {
    return err
  } else if has == false {
    return fmt.Errorf("Repository not found")
  }

  image := new(Image)

  if has, _, err := image.Has(imageId); err != nil {
    return err
  } else if has == false {
    return fmt.Errorf("Tag's image not found")
  }

  t := new(Tag)
  t.UUID = string(fmt.Sprintf("%s:%s:%s", namespace, repository, tag))
  t.Name, t.ImageId, t.Namespace, t.Repository = tag, imageId, namespace, repository

  if err := t.Save(); err != nil {
    return err
  }

  r.Tags = append(r.Tags, t.UUID)

  if err := r.Save(); err != nil {
    return err
  }

  return nil
}

func (r *Repository) PutImages(namespace, repository string) error {
  if has, _, err := r.Has(namespace, repository); err != nil {
    return err
  } else if has == false {
    return fmt.Errorf("Repository not found")
  }

  r.Checksumed, r.Uploaded, r.Updated = true, true, time.Now().Unix()

  if err := r.Save(); err != nil {
    return err
  }

  return nil
}

func (p *Privilege) Get(UUID string) error {
  if err := Get(p, []byte(UUID)); err != nil {
    return err
  }

  return nil
}

func (t *Tag) Has(namespace, repository, image, tag string) (bool, []byte, error) {
  UUID, err := GetUUID("tag", fmt.Sprintf("%s:%s:%s:%s", namespace, repository, image, tag))
  if err != nil {
    return false, nil, err
  }

  if len(UUID) <= 0 {
    return false, nil, nil
  }

  err = Get(t, UUID)

  return true, UUID, err
}

func (t *Tag) Save() error {
  if err := Save(t, []byte(t.UUID)); err != nil {
    return err
  }

  if _, err := LedisDB.HSet([]byte(GLOBAL_TAG_INDEX), []byte(fmt.Sprintf("%s:%s:%s:%s", t.Namespace, t.Repository, t.ImageId, t.Name)), []byte(t.UUID)); err != nil {
    return err
  }

  return nil
}

func (t *Tag) GetByUUID(uuid string) error {
  return Get(t, []byte(uuid))
}
