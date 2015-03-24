package models

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dockercn/wharf/utils"
)

type Repository struct {
	Id          string   `json:"id"`          //
	Repository  string   `json:"repository"`  //
	Namespace   string   `json:"namespace"`   //
	Tags        []string `json:"tags"`        //
	Starts      []string `json:"starts"`      //
	Comments    []string `json:"comments"`    //
	Short       string   `json:"short"`       //
	Description string   `json:"description"` //
	JSON        string   `json:"json"`        //
	Dockerfile  string   `json:"dockerfile"`  //
	Agent       string   `json:"agent"`       //
	Links       string   `json:"links"`       //
	Size        int64    `json:"size"`        //
	Download    int64    `json:"download"`    //
	Uploaded    bool     `json:"uploaded"`    //
	Checksum    string   `json:"checksum"`    //
	Checksumed  bool     `json:"checksumed"`  //
	Icon        string   `json:"icon"`        //
	Sign        string   `json:"sign"`        //
	Privated    bool     `json:"privated"`    //
	Permissions []string `json:"permissions"` //
	Clear       string   `json:"clear"`       //
	Cleared     bool     `json:"cleared"`     //
	Encrypted   bool     `json:"encrypted"`   //
	Version     int64    `json:"version"`     //
	Created     int64    `json:"created"`     //
	Updated     int64    `json:"updated"`     //
	Memo        []string `json:"memo"`        //
}

func (r *Repository) Has(namespace, repository string) (bool, []byte, error) {
	id, err := GetByGobalId("repository", fmt.Sprintf("%s:%s", namespace, repository))

	if err != nil {
		return false, nil, err
	}

	if len(id) <= 0 {
		return false, nil, nil
	}
	err = Get(r, id)

	return true, id, err
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

func (r *Repository) Save() error {
	if err := Save(r, []byte(r.Id)); err != nil {
		return err
	}

	if _, err := LedisDB.HSet([]byte(GLOBAL_REPOSITORY_INDEX), []byte(fmt.Sprintf("%s:%s", r.Namespace, r.Repository)), []byte(r.Id)); err != nil {
		return err
	}

	return nil
}

func (r *Repository) Remove() error {
	if _, err := LedisDB.HSet([]byte(fmt.Sprintf("%s_remove", GLOBAL_REPOSITORY_INDEX)), []byte(fmt.Sprintf("%s:%s", r.Namespace, r.Repository)), []byte(r.Id)); err != nil {
		return err
	}

	if _, err := LedisDB.HDel([]byte(GLOBAL_REPOSITORY_INDEX), []byte(fmt.Sprintf("%s:%s", r.Namespace, r.Repository))); err != nil {
		return err
	}

	return nil
}

func (r *Repository) Put(namespace, repository, json, agent string, version int64) error {
	if has, _, err := r.Has(namespace, repository); err != nil {
		return err
	} else if has == false {
		r.Id = string(utils.GeneralKey(fmt.Sprintf("%s:%s", namespace, repository)))
		r.Created = time.Now().UnixNano() / int64(time.Millisecond)
	}

	r.Namespace, r.Repository, r.JSON, r.Agent, r.Version = namespace, repository, json, agent, version

	r.Updated = time.Now().UnixNano() / int64(time.Millisecond)
	r.Checksumed, r.Uploaded, r.Cleared, r.Encrypted = false, false, false, false
	r.Size, r.Download = 0, 0

	if err := r.Save(); err != nil {
		return err
	}

	return nil
}

func (repository *Repository) Get(id string) error {
	if err := Get(repository, []byte(id)); err != nil {
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
	t.Id = string(fmt.Sprintf("%s:%s:%s", namespace, repository, tag))
	t.Name, t.ImageId, t.Namespace, t.Repository = tag, imageId, namespace, repository

	if err := t.Save(); err != nil {
		return err
	}

	has := false
	for _, value := range r.Tags {
		if value == t.Id {
			has = true
		}
	}
	if !has {
		r.Tags = append(r.Tags, t.Id)
	}
	if err := r.Save(); err != nil {
		return err
	}

	return nil
}

func (r *Repository) PutJSONFromManifests(image map[string]string, namespace, repository string) error {
	if has, _, err := r.Has(namespace, repository); err != nil {
		return err
	} else if has == false {
		r.Id = string(utils.GeneralKey(fmt.Sprintf("%s:%s", namespace, repository)))
		r.Created = time.Now().UnixNano() / int64(time.Millisecond)
		r.JSON = ""
	}

	r.Namespace, r.Repository, r.Version = namespace, repository, APIVERSION_V2

	r.Updated = time.Now().UnixNano() / int64(time.Millisecond)
	r.Checksumed, r.Uploaded, r.Cleared, r.Encrypted = true, true, true, false
	r.Size, r.Download = 0, 0

	if len(r.JSON) == 0 {
		if data, err := json.Marshal([]map[string]string{image}); err != nil {
			return err
		} else {
			r.JSON = string(data)
		}

	} else {
		var ids []map[string]string

		if err := json.Unmarshal([]byte(r.JSON), &ids); err != nil {
			return err
		}

		has := false
		for _, v := range ids {
			if v["id"] == image["id"] {
				has = true
			}
		}

		if has == false {
			ids = append(ids, image)
		}

		if data, err := json.Marshal(ids); err != nil {
			return err
		} else {
			r.JSON = string(data)
		}
	}

	if err := r.Save(); err != nil {
		return err
	}

	log.Println("[REGISTRY API V2] Convert Manifests To JSON: ", r.JSON)

	return nil
}

func (r *Repository) PutTagFromManifests(image, namespace, repository, tag, manifests string) error {
	if has, _, err := r.Has(namespace, repository); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("Repository not found")
	}

	t := new(Tag)
	t.Id = string(fmt.Sprintf("%s:%s:%s", namespace, repository, tag))
	t.Name, t.ImageId, t.Namespace, t.Repository, t.Manifest = tag, image, namespace, repository, manifests

	if err := t.Save(); err != nil {
		return err
	}

	has := false
	for _, v := range r.Tags {
		if v == t.Id {
			has = true
		}
	}

	if has == false {
		r.Tags = append(r.Tags, t.Id)
	}

	if err := r.Save(); err != nil {
		return err
	}

	log.Println("[REGISTRY API V2] Tag: ", t)
	log.Println("[REGISTRY API V2] Repository Tags: ", r.Tags)

	return nil
}
