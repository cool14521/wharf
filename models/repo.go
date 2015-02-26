package models

import (
	"fmt"
	"time"

	"github.com/dockercn/wharf/utils"
)

type Repository struct {
	UUID          string   `json:"UUID"`          //
	Repository    string   `json:"repository"`    //
	Namespace     string   `json:"namespace"`     //
	NamespaceType bool     `json:"namespacetype"` //
	Organization  string   `json:"organization"`  //
	Tags          []string `json:"tags"`          //
	Starts        []string `json:"starts"`        //
	Comments      []string `json:"comments"`      //
	Short         string   `json:"short"`         //
	Description   string   `json:"description"`   //
	JSON          string   `json:"json"`          //
	Dockerfile    string   `json:"dockerfile"`    //
	Agent         string   `json:"agent"`         //
	Links         string   `json:"links"`         //
	Size          int64    `json:"size"`          //
	Download      int64    `json:"download"`      //
	Uploaded      bool     `json:"uploaded"`      //
	Checksum      string   `json:"checksum"`      //
	Checksumed    bool     `json:"checksumed"`    //
	Icon          string   `json:"icon"`          //
	Sign          string   `json:"sign"`          //
	Privated      bool     `json:"privated"`      //
	Clear         string   `json:"clear"`         //
	Cleared       bool     `json:"cleared"`       //
	Encrypted     bool     `json:"encrypted"`     //
	Created       int64    `json:"created"`       //
	Updated       int64    `json:"updated"`       //
	Memo          []string `json:"memo"`          //
}

type Star struct {
	UUID       string   `json:"UUID"`       //
	User       string   `json:"user"`       //
	Repository string   `json:"repository"` //
	Time       int64    `json:"time"`       //
	Memo       []string `json:"memo"`       //
}

type Comment struct {
	UUID       string   `json:"UUID"`       //
	Comment    string   `json:"comment"`    //
	User       string   `json:"user"`       //
	Repository string   `json:"repository"` //
	Time       int64    `json:"time"`       //
	Memo       []string `json:"memo"`       //
}

type Privilege struct {
	UUID       string   `json:"UUID"`       //
	Privilege  bool     `json:"privilege"`  //
	Team       string   `json:"team"`       //
	Repository string   `json:"repository"` //
	Memo       []string `json:"memo"`       //
}

type Tag struct {
	UUID       string   `json:"uuid"`       //
	Name       string   `json:"name"`       //
	ImageId    string   `json:"imageid"`    //
	Namespace  string   `json:"namespace"`  //
	Repository string   `json:"repository"` //
	Sign       string   `json:"sign"`       //
	Memo       []string `json:"memo"`       //
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
