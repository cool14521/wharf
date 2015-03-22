package models

import (
  "fmt"
)

type Tag struct {
  Id         string   `json:"id"`         //
  Name       string   `json:"name"`       //
  ImageId    string   `json:"imageid"`    //
  Namespace  string   `json:"namespace"`  //
  Repository string   `json:"repository"` //
  Sign       string   `json:"sign"`       //
  Manifest   string   `json:"manifest"`   //
  Memo       []string `json:"memo"`       //
}

func (t *Tag) Has(namespace, repository, tag string) (bool, []byte, error) {
  id, err := GetByGobalId("tag", fmt.Sprintf("%s:%s:%s", namespace, repository, tag))
  if err != nil {
    return false, nil, err
  }

  if len(id) <= 0 {
    return false, nil, nil
  }

  err = Get(t, id)

  return true, id, err
}

func (t *Tag) Save() error {
  if err := Save(t, []byte(t.Id)); err != nil {
    return err
  }

  if _, err := LedisDB.HSet([]byte(GLOBAL_TAG_INDEX), []byte(fmt.Sprintf("%s:%s:%s:%s", t.Namespace, t.Repository, t.ImageId, t.Name)), []byte(t.Id)); err != nil {
    return err
  }

  return nil
}

func (t *Tag) GetById(id string) error {
  return Get(t, []byte(id))
}