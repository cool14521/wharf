package routers

import (
	"github.com/astaxie/beego"
	"github.com/dockercn/docker-bucket/controllers"
)

func init() {
	//Web Interface
	beego.Router("/", &controllers.MainController{})
	beego.Router("/auth", &controllers.AuthController{}, "get:Get")

	//Static File
	beego.Router("/favicon.ico", &controllers.StaticController{}, "get:GetFavicon")
	//TODO sitemap/rss/robots.txt

	//CI Service API
	drone := beego.NewNamespace("/d1",
		beego.NSRouter("/yaml", &controllers.DroneAPIController{}, "post:PostYAML"),
	)

	//Docker Registry API V1 remain
	beego.Router("/_ping", &controllers.PingAPIController{}, "get:GetPing")
	beego.Router("/_status", &controllers.StatusAPIController{})

	//Docker Registry API V1
	api := beego.NewNamespace("/v1",
		beego.NSRouter("/_ping", &controllers.PingAPIController{}, "get:GetPing"),
		beego.NSRouter("/_status", &controllers.StatusAPIController{}),
		beego.NSRouter("/users", &controllers.UsersAPIController{}, "get:GetUsers"),
		beego.NSRouter("/users", &controllers.UsersAPIController{}, "post:PostUsers"),

		beego.NSNamespace("/repositories",
			beego.NSRouter("/:namespace/:repo_name/tags/:tag", &controllers.RepositoryAPIController{}, "put:PutTag"),
			beego.NSRouter("/:namespace/:repo_name/images", &controllers.RepositoryAPIController{}, "put:PutRepositoryImages"),
			beego.NSRouter("/:namespace/:repo_name/images", &controllers.RepositoryAPIController{}, "get:GetRepositoryImages"),
			beego.NSRouter("/:namespace/:repo_name/tags", &controllers.RepositoryAPIController{}, "get:GetRepositoryTags"),
			beego.NSRouter("/:namespace/:repo_name", &controllers.RepositoryAPIController{}, "put:PutRepository"),
		),

		beego.NSNamespace("/images",
			beego.NSRouter("/:image_id/ancestry", &controllers.ImageAPIController{}, "get:GetImageAncestry"),
			beego.NSRouter("/:image_id/json", &controllers.ImageAPIController{}, "get:GetImageJSON"),
			beego.NSRouter("/:image_id/layer", &controllers.ImageAPIController{}, "get:GetImageLayer"),
			beego.NSRouter("/:image_id/json", &controllers.ImageAPIController{}, "put:PutImageJSON"),
			beego.NSRouter("/:image_id/layer", &controllers.ImageAPIController{}, "put:PutImageLayer"),
			beego.NSRouter("/:image_id/checksum", &controllers.ImageAPIController{}, "put:PutChecksum"),
		),
	)

	beego.AddNamespace(drone)
	beego.AddNamespace(api)
}
