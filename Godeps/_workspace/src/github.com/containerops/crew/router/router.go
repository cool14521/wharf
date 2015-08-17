package router

import (
	"github.com/Unknwon/macaron"

	"github.com/macaron-contrib/binding"

	"github.com/containerops/crew/handler"
)

func SetRouters(m *macaron.Macaron) {
	m.Group("/v1", func() {
		//Session Router
		m.Group("/token", func() {
			m.Post("/", handler.W1PostToken)
		})

		//User Router
		m.Group("/user", func() {
			//Signin and Signup
			m.Post("/", binding.Bind(handler.UserSignup{}), handler.W1UserSignup)
			m.Post("/auth", handler.W1UserSignin)
			//List All Users
			m.Get("/list/:count/:page", handler.W1GetUserList)
			//Profile
			m.Put("/:user/profile", handler.W1PutUserProfile)
			m.Get("/:user/profile", handler.W1GetUserProfile)
			m.Post("/:user/gravatar", handler.W1PostUserGravatar)
			//Put Password
			m.Put("/:user/passwd", handler.W1PutUserPasswd)
			//List User Teams and Organizations
			m.Get("/:user/organizations", handler.W1GetUserOrganizations)
			m.Get("/:user/teams", handler.W1GetUserTeams)
		})

		//Organization Router
		m.Group("/org", func() {
			m.Post("/", handler.W1PostOrganization)
			m.Put("/:org", handler.W1PutOrganization)
			m.Get("/:org", handler.W1GetOrganization)
			m.Delete("/:org", handler.W1DeleteOrganization)

			//Team Router
			m.Group("/:org/team", func() {
				m.Post("/", handler.W1PostTeam)
				m.Get("/list", handler.W1GetTeams)
				m.Put("/:team", handler.W1PutTeam)
				m.Get("/:team", handler.W1GetTeam)
				m.Delete("/:team", handler.W1DeleteTeam)

				//User Management
				m.Group("/:team/user", func() {
					m.Get("/list", handler.W1GetTeamUsers)
					m.Put("/:user", handler.W1PutTeamUser)
					m.Delete("/:user", handler.W1DeleteTeamUser)
				})
			})
		})
	})
}
