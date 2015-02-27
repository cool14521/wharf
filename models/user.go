package models

import (
	"fmt"
	"regexp"
)

type User struct {
	UUID              string   `json:"UUID"`          //
	Username          string   `json:"username"`      //
	Password          string   `json:"password"`      //
	Email             string   `json:"email"`         //
	Fullname          string   `json:"fullname"`      //
	Company           string   `json:"company"`       //
	Location          string   `json:"location"`      //
	Mobile            string   `json:"mobile"`        //
	URL               string   `json:"url"`           //
	Gravatar          string   `json:"gravatar"`      //
	Created           int64    `json:"created"`       //
	Updated           int64    `json:"updated"`       //
	Repositories      []string `json:"repositories"`  //
	Organizations     []string `json:"organizations"` //
	Teams             []string `json:"teams"`         //
	Starts            []string `json:"starts"`        //
	Comments          []string `json:"comments"`      //
	Memo              []string `json:"memo"`          //
	JoinOrganizations []string `json:"joinorganizations"`
	JoinTeams         []string `json:"jointeams"`
}

func (user *User) Has(username string) (bool, []byte, error) {

	UUID, err := GetUUID("user", username)

	if err != nil {
		return false, nil, err
	}

	if len(UUID) <= 0 {
		return false, nil, nil
	}

	err = Get(user, UUID)

	return true, UUID, err
}

func (user *User) Save() error {
	//https://github.com/docker/docker/blob/28f09f06326848f4117baf633ec9fc542108f051/registry/registry.go#L27
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

	if err := Save(user, []byte(user.UUID)); err != nil {
		return err
	}

	if _, err := LedisDB.HSet([]byte(GLOBAL_USER_INDEX), []byte(user.Username), []byte(user.UUID)); err != nil {
		return err
	}

	return nil
}

func (user *User) Remove() error {
	if _, err := LedisDB.HSet([]byte(fmt.Sprintf("%s_remove", GLOBAL_USER_INDEX)), []byte(user.Username), []byte(user.UUID)); err != nil {
		return err
	}

	if _, err := LedisDB.HDel([]byte(GLOBAL_USER_INDEX), []byte(user.Username)); err != nil {
		return err
	}

	return nil
}

func (user *User) Get(username, password string) error {
	if exist, UUID, err := user.Has(username); err != nil {
		return err
	} else if exist == false && err == nil {
		return fmt.Errorf("User is not exist: %s", username)
	} else if exist == true && err == nil {
		if err := Get(user, UUID); err != nil {
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

func (user *User) Orgs(username string) (map[string]string, error) {
	result := map[string]string{}

	if exist, _, err := user.Has(username); err != nil {
		return nil, err
	} else if exist == false && err == nil {
		return nil, fmt.Errorf("User is not exist: %s", username)
	} else if exist == true && err == nil {
		for _, uuid := range user.Organizations {
			var org Organization

			if err := org.Get(uuid); err == nil {
				result[org.Organization] = org.UUID
			}
		}
	}

	return result, nil
}
