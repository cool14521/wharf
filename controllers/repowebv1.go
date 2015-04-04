package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/astaxie/beego"

	"github.com/containerops/wharf/models"
	"github.com/containerops/wharf/utils"
)

type RepoWebAPIV1Controller struct {
	beego.Controller
}

func (this *RepoWebAPIV1Controller) Prepare() {
	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func (this *RepoWebAPIV1Controller) JSONOut(code int, message string, data interface{}) {
	if data == nil {
		this.Data["json"] = map[string]string{"message": message}
	} else {
		this.Data["json"] = data
	}

	this.Ctx.Output.Context.Output.SetStatus(code)
	this.ServeJson()
}

func (this *RepoWebAPIV1Controller) URLMapping() {
	this.Mapping("GetRepositories", this.GetRepositories)
	this.Mapping("PostRepository", this.PostRepository)
	this.Mapping("PutRepository", this.PutRepository)
	this.Mapping("GetRepository", this.GetRepository)
	this.Mapping("GetCollaborators", this.GetCollaborators)
	this.Mapping("PostCollaborator", this.PostCollaborator)
	this.Mapping("PutCollaborator", this.PutCollaborator)
}

func (this *RepoWebAPIV1Controller) GetRepositories() {
	user := new(models.User)
	org := new(models.Organization)
	repos := make([]models.Repository, 0)

	if exist, _, err := user.Has(this.Ctx.Input.Param(":namespace")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		if exist, _, err := org.Has(this.Ctx.Input.Param(":namespace")); err != nil {
			this.JSONOut(http.StatusBadRequest, err.Error(), nil)
			return
		} else if exist == false {
			this.JSONOut(http.StatusBadRequest, "Invalide Namespace", nil)
			return
		}

		for _, id := range org.Repositories {
			repo := new(models.Repository)
			if err := repo.Get(id); err != nil {
				continue
			}
			repos = append(repos, *repo)
		}

		this.JSONOut(http.StatusOK, "", repos)
		return
	} else if exist == true {
		for _, id := range user.Repositories {
			repo := new(models.Repository)
			if err := repo.Get(id); err != nil {
				continue
			}
			repos = append(repos, *repo)
		}

		this.JSONOut(http.StatusOK, "", repos)
		return
	}
}

func (this *RepoWebAPIV1Controller) PostRepository() {
	user := new(models.User)
	org := new(models.Organization)
	repo := new(models.Repository)

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &repo); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	if exist, _, err := repo.Has(this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":repository")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == true {
		this.JSONOut(http.StatusBadRequest, "Repository already exist.", nil)
		return
	}

	repo.Id = string(utils.GeneralKey(fmt.Sprint(repo.Namespace, repo.Repository)))
	repo.Created = time.Now().UnixNano() / int64(time.Millisecond)
	repo.Updated = time.Now().UnixNano() / int64(time.Millisecond)
	repo.Collaborators, repo.Permissions = []string{}, []string{}

	if err := repo.Save(); err != nil {
		this.JSONOut(http.StatusBadRequest, "Repository save error.", nil)
		return
	}

	if exist, _, err := user.Has(this.Ctx.Input.Param(":namespace")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		if exist, _, err := org.Has(this.Ctx.Input.Param(":namespace")); err != nil {
			this.JSONOut(http.StatusBadRequest, err.Error(), nil)
			return
		} else if exist == false {
			this.JSONOut(http.StatusBadRequest, "Invalide Namespace", nil)
			return
		}

		org.Repositories = append(org.Repositories, repo.Id)
		org.Updated = time.Now().UnixNano() / int64(time.Millisecond)

		if err := org.Save(); err != nil {
			this.JSONOut(http.StatusBadRequest, "Organization save error.", nil)
			return
		}
	} else if exist == true {
		user.Repositories = append(user.Repositories, repo.Id)
		user.Updated = time.Now().UnixNano() / int64(time.Millisecond)

		if err := user.Save(); err != nil {
			this.JSONOut(http.StatusBadRequest, err.Error(), nil)
			return
		}
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	repo.Log(models.ACTION_ADD_REPO, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, repo.Id, memo)
	user.Log(models.ACTION_ADD_REPO, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, user.Id, memo)

	this.JSONOut(http.StatusOK, "Repository create successfully!", nil)
	return
}

func (this *RepoWebAPIV1Controller) PutRepository() {
	repo := new(models.Repository)

	if exist, _, err := repo.Has(this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":repository")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := json.Unmarshal(this.Ctx.Input.CopyBody(), &repo); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	repo.Updated = time.Now().UnixNano() / int64(time.Millisecond)

	if err := repo.Save(); err != nil {
		this.JSONOut(http.StatusBadRequest, "Repository save error.", nil)
		return
	}

	memo, _ := json.Marshal(this.Ctx.Input.Header)
	repo.Log(models.ACTION_ADD_REPO, models.LEVELINFORMATIONAL, models.TYPE_WEBV1, repo.Id, memo)

	this.JSONOut(http.StatusOK, "Repository update successfully!", nil)
	return
}

func (this *RepoWebAPIV1Controller) GetRepository() {
	repo := new(models.Repository)

	if exist, _, err := repo.Has(this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":repository")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	}

	this.JSONOut(http.StatusOK, "", repo)
	return
}

func (this *RepoWebAPIV1Controller) GetCollaborators() {
	repo := new(models.Repository)
	user := new(models.User)
	org := new(models.Organization)

	if exist, _, err := repo.Has(this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":repository")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Repository Invalid", nil)
		return
	}

	if exist, _, err := user.Has(this.Ctx.Input.Param(":namespace")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == true {
		this.JSONOut(http.StatusOK, "", repo.Collaborators)
		return
	}

	if exist, _, err := org.Has(this.Ctx.Input.Param(":namespace")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusOK, "Namespace Invalid", nil)
		return
	} else {
		this.JSONOut(http.StatusOK, "", repo.Permissions)
		return
	}
}

func (this *RepoWebAPIV1Controller) PostCollaborator() {
	repo := new(models.Repository)
	user := new(models.User)
	collaborator := new(models.User)

	if exist, _, err := collaborator.Has(this.Ctx.Input.Param(":collaborator")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Collaborator Invalid", nil)
	}

	if exist, _, err := repo.Has(this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":repository")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Repository Invalid", nil)
		return
	}

	if exist, _, err := user.Has(this.Ctx.Input.Param(":namespace")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == true {
		for _, v := range repo.Collaborators {
			if v == this.Ctx.Input.Param(":collaborator") {
				this.JSONOut(http.StatusBadRequest, "User already in collaborators", nil)
				return
			}
		}

		repo.Collaborators = append(repo.Collaborators, this.Ctx.Input.Param(":collaborator"))
		repo.Updated = time.Now().UnixNano() / int64(time.Millisecond)

		if err := repo.Save(); err != nil {
			this.JSONOut(http.StatusBadRequest, "Repository save error.", nil)
			return
		}

		collaborator.Repositories = append(collaborator.Repositories, repo.Id)
		collaborator.Updated = time.Now().UnixNano() / int64(time.Millisecond)

		if err := collaborator.Save(); err != nil {
			this.JSONOut(http.StatusBadRequest, "User save error.", nil)
			return
		}

		this.JSONOut(http.StatusOK, "Add collaborator successfully.", nil)
		return
	}

	team := new(models.Team)
	org := new(models.Organization)

	if exist, _, err := org.Has(this.Ctx.Input.Param(":namespace")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Organization Invalid", nil)
		return
	}

	if exist, _, err := team.Has(this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":collaborator")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Team Invalid", nil)
		return
	}

	for _, v := range repo.Permissions {
		if v == this.Ctx.Input.Param(":collaborator") {
			this.JSONOut(http.StatusBadRequest, "User already in collaborators", nil)
			return
		}
	}

	repo.Permissions = append(repo.Permissions, this.Ctx.Input.Param(":collaborator"))
	repo.Updated = time.Now().UnixNano() / int64(time.Millisecond)

	if err := repo.Save(); err != nil {
		this.JSONOut(http.StatusBadRequest, "Repository save error.", nil)
		return
	}

	team.Repositories = append(team.Repositories, repo.Id)
	team.Updated = time.Now().UnixNano() / int64(time.Millisecond)

	if err := team.Save(); err != nil {
		this.JSONOut(http.StatusBadRequest, "Team save error.", nil)
		return
	}

	this.JSONOut(http.StatusOK, "Add collaborator successfully.", nil)
	return
}

func (this *RepoWebAPIV1Controller) PutCollaborator() {
	repo := new(models.Repository)
	user := new(models.User)
	collaborator := new(models.User)

	if exist, _, err := collaborator.Has(this.Ctx.Input.Param(":collaborator")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Collaborator Invalid", nil)
	}

	if exist, _, err := repo.Has(this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":repository")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Repository Invalid", nil)
		return
	}

	if exist, _, err := user.Has(this.Ctx.Input.Param(":namespace")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == true {
		for i, v := range repo.Collaborators {
			if v == collaborator.Username {
				repo.Collaborators = append(repo.Collaborators[:i], repo.Collaborators[i+1:]...)
				repo.Updated = time.Now().UnixNano() / int64(time.Millisecond)

				if err := repo.Save(); err != nil {
					this.JSONOut(http.StatusBadRequest, "Repository save error.", nil)
					return
				}

				for k, t := range collaborator.Repositories {
					if t == repo.Id {
						collaborator.Repositories = append(collaborator.Repositories[:k], collaborator.Repositories[k+1:]...)
						collaborator.Updated = time.Now().UnixNano() / int64(time.Millisecond)

						if err := collaborator.Save(); err != nil {
							this.JSONOut(http.StatusBadRequest, "User save error.", nil)
							return
						}
					}
				}

				this.JSONOut(http.StatusOK, "Remove collaborator successfully.", nil)
				return
			}
		}

		this.JSONOut(http.StatusBadRequest, "Remove collaborator failure.", nil)
	}

	team := new(models.Team)
	org := new(models.Organization)

	if exist, _, err := org.Has(this.Ctx.Input.Param(":namespace")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Organization Invalid", nil)
		return
	}

	if exist, _, err := team.Has(this.Ctx.Input.Param(":namespace"), this.Ctx.Input.Param(":collaborator")); err != nil {
		this.JSONOut(http.StatusBadRequest, err.Error(), nil)
		return
	} else if exist == false {
		this.JSONOut(http.StatusBadRequest, "Team Invalid", nil)
		return
	}

	for i, v := range repo.Permissions {
		if v == this.Ctx.Input.Param(":collaborator") {
			repo.Permissions = append(repo.Collaborators[:i], repo.Collaborators[i+1:]...)
			repo.Updated = time.Now().UnixNano() / int64(time.Millisecond)

			if err := repo.Save(); err != nil {
				this.JSONOut(http.StatusBadRequest, "Repository save error.", nil)
				return
			}

			for k, t := range org.Repositories {
				if t == repo.Id {
					org.Repositories = append(org.Repositories[:k], org.Repositories[:k+1]...)
					repo.Updated = time.Now().UnixNano() / int64(time.Millisecond)

					if err := org.Save(); err != nil {
						this.JSONOut(http.StatusBadRequest, "Repository save error.", nil)
					}
				}
			}
		}
	}

	for i, v := range team.Repositories {
		if v == repo.Id {
			team.Repositories = append(team.Repositories[:i], team.Repositories[i+1:]...)
			team.Updated = time.Now().UnixNano() / int64(time.Millisecond)

			if err := team.Save(); err != nil {
				this.JSONOut(http.StatusBadRequest, "Team save error.", nil)
				return
			}

			this.JSONOut(http.StatusOK, "Add collaborator successfully.", nil)
			return
		}
	}

	this.JSONOut(http.StatusBadRequest, "Remove collaborator failure.", nil)
}
