package models

import (
  "fmt"
)

type Compose struct {
  UUID          string   `json:"UUID"`          //全局唯一的索引, LedisDB 中 Compose List 保存全局所有的仓库名列表信息。 LedisDB 独立保存每个 Compose 信息到一个 HASH，名字为 {UUID}
  Compose       string   `json:"compose"`       //仓库名称 全局唯一，不可修改
  Namespace     string   `json:"namespace"`     //仓库所有者的名字
  NamespaceType bool     `json:"namespacetype"` // false 为普通用户，true 为组织
  Organization  string   `json:"organization"`  //如果仓库属于一个 team，那么在此记录 team 所属组织
  Tags          []string `json:"tags"`          //保存此 Compose 所有 tag 的对应 UUID
  Starts        []string `json:"starts"`        //保存此 Compose Start的UUID列表
  Comments      []string `json:"comments"`      //保存此 Compose Comment的对应UUID列表
  Short         string   `json:"short"`         //此仓库的的短描述
  Description   string   `json:"description"`   //保存 Markdown 格式
  YAML          string   `json:"yaml"`          //保存 Compose 的 YAML 内容
  Download      int64    `json:"download"`      //下载次数
  Icon          string   `json:"icon"`          //
  Privated      bool     `json:"privated"`      //私有 Compose
  Created       int64    `json:"created"`       //
  Updated       int64    `json:"updated"`       //
  Memo          string   `json:"memo"`          //
}

func (c *Compose) Has(namespace, compose string) (bool, []byte, error) {

  UUID, err := GetUUID("compose", fmt.Sprintf("%s:%s", namespace, compose))

  if err != nil {
    return false, nil, err
  }

  if len(UUID) <= 0 {
    return false, nil, nil
  }

  err = Get(c, UUID)

  return true, UUID, err
}

func (c *Compose) Save() error {
  if err := Save(c, []byte(c.UUID)); err != nil {
    return err
  }

  if _, err := LedisDB.HSet([]byte(GLOBAL_COMPOSE_INDEX), []byte(fmt.Sprintf("%s:%s", c.Namespace, c.Compose)), []byte(c.UUID)); err != nil {
    return err
  }

  return nil
}
