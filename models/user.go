package models

import (
	"fmt"
	"regexp"
	"time"

	"github.com/dockercn/wharf/utils"
)

type User struct {
	Id                string   `json:"id"`                //
	Username          string   `json:"username"`          //
	Password          string   `json:"password"`          //
	Email             string   `json:"email"`             //
	Fullname          string   `json:"fullname"`          //
	Company           string   `json:"company"`           //
	Location          string   `json:"location"`          //
	Mobile            string   `json:"mobile"`            //
	URL               string   `json:"url"`               //
	Gravatar          string   `json:"gravatar"`          //
	Created           int64    `json:"created"`           //
	Updated           int64    `json:"updated"`           //
	Repositories      []string `json:"repositories"`      //
	Organizations     []string `json:"organizations"`     // Owner's Organizations
	Teams             []string `json:"teams"`             // Owner's Teams
	JoinOrganizations []string `json:"joinorganizations"` // Join's Organizations
	JoinTeams         []string `json:"jointeams"`         // Join's Teams
	Starts            []string `json:"starts"`            //
	Comments          []string `json:"comments"`          //
	Memo              []string `json:"memo"`              //
}

func (user *User) Has(username string) (bool, []byte, error) {
	id, err := GetByGobalId("user", username)

	if err != nil {
		return false, nil, err
	}

	if len(id) <= 0 {
		return false, nil, nil
	}

	err = Get(user, id)

	return true, id, err
}

func (user *User) GetById(id string) error {
	if err := Get(user, []byte(id)); err != nil {
		return err
	}
	return nil
}

func (user *User) Get(username, password string) error {
	if exist, id, err := user.Has(username); err != nil {
		return err
	} else if exist == false && err == nil {
		return fmt.Errorf("User is not exist: %s", username)
	} else if exist == true && err == nil {
		if err := Get(user, id); err != nil {
			return err
		} else {
			if user.Password != password {
				return fmt.Errorf("User password error.")
			} else {
				return nil
			}
		}
	}
	return nil
}

func (user *User) Save() error {
	validNamespace := regexp.MustCompile(`^([a-z0-9_]{4,30})$`)
	if !validNamespace.MatchString(user.Username) {
		return fmt.Errorf("Username must be 4 - 30, include a-z, 0-9 and '_'")
	}

	if len(user.Password) < 5 {
		return fmt.Errorf("Password length should be more than 5")
	}

	validEmail := regexp.MustCompile("[\\w!#$%&'*+/=?^_`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[a-zA-Z0-9](?:[\\w-]*[\\w])?")
	if !validEmail.MatchString(user.Email) {
		return fmt.Errorf("Email illegal")
	}

	if err := Save(user, []byte(user.Id)); err != nil {
		return err
	}

	if _, err := LedisDB.HSet([]byte(GLOBAL_USER_INDEX), []byte(user.Username), []byte(user.Id)); err != nil {
		return err
	}

	return nil
}

func (user *User) Remove() error {
	if _, err := LedisDB.HSet([]byte(fmt.Sprintf("%s_remove", GLOBAL_USER_INDEX)), []byte(user.Username), []byte(user.Id)); err != nil {
		return err
	}

	if _, err := LedisDB.HDel([]byte(GLOBAL_USER_INDEX), []byte(user.Username)); err != nil {
		return err
	}

	return nil
}

func (user *User) Orgs(username string) (map[string]string, error) {
	result := map[string]string{}

	if exist, _, err := user.Has(username); err != nil {
		return nil, err
	} else if exist == false && err == nil {
		return nil, fmt.Errorf("User is not exist: %s", username)
	} else if exist == true && err == nil {
		for _, id := range user.Organizations {
			var org Organization

			if err := org.GetById(id); err == nil {
				result[org.Name] = org.Id
			}
		}
	}

	return result, nil
}

func (user *User) All() []*User {
	vfValues, _ := LedisDB.HGetAll([]byte(GLOBAL_USER_INDEX))

	allUsers := make([]*User, 0, 1)

	for _, vfValue := range vfValues {
		nowUser := new(User)
		nowUser.Has(string(vfValue.Field))
		allUsers = append(allUsers, nowUser)
	}

	return allUsers
}

func (user *User) Log(action, level, t int64, actionID string, content []byte) error {
	log := Log{Action: action, ActionId: actionID, Level: level, Type: t, Content: string(content), Created: time.Now().UnixNano() / int64(time.Millisecond)}
	log.Id = string(utils.GeneralKey(actionID))

	if err := log.Save(); err != nil {
		return err
	}

	user.Memo = append(user.Memo, log.Id)

	if err := user.Save(); err != nil {
		return err
	}

	return nil
}
