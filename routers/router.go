package routers

import (
	"github.com/astaxie/beego"

	"github.com/containerops/wharf/controllers"
	"github.com/containerops/wharf/filters"
)

func init() {
	//Web Interface
	beego.Router("/", &controllers.WebController{}, "get:GetIndex")
	beego.Router("/auth", &controllers.WebController{}, "get:GetAuth")
	beego.Router("/setting", &controllers.WebController{}, "get:GetSetting")
	beego.Router("/dashboard", &controllers.WebController{}, "get:GetDashboard")
	beego.Router("/signout", &controllers.WebController{}, "get:GetSignout")
	beego.Router("/admin/auth", &controllers.WebController{}, "get:GetAdminAuth")
	beego.Router("/admin", &controllers.WebController{}, "get:GetAdmin")

	beego.Router("/c/:namespace/:compose", &controllers.WebController{}, "get:GetCompose")
	beego.Router("/r/:namespace/:repository", &controllers.WebController{}, "get:GetRepository")
	beego.Router("/o/:org", &controllers.WebController{}, "get:GetOrganization")
	beego.Router("/u/:username", &controllers.WebController{}, "get:GetUser")

	//Static File Route
	beego.Router("/pubkeys.gpg", &controllers.FileController{}, "get:GetGPG")

	//Web API
	web := beego.NewNamespace("/w1",
		//user routers
		beego.NSNamespace("/user",
			beego.NSRouter("/signin", &controllers.UserWebAPIV1Controller{}, "post:Signin"),
			beego.NSRouter("/signup", &controllers.UserWebAPIV1Controller{}, "post:Signup"),
			beego.NSRouter("/users", &controllers.UserWebAPIV1Controller{}, "get:GetUsers"),
			beego.NSRouter("/:username", &controllers.UserWebAPIV1Controller{}, "get:GetUser"),
			beego.NSRouter("/:username/namespaces", &controllers.UserWebAPIV1Controller{}, "get:GetNamespaces"),
			beego.NSRouter("/:username/gravatar", &controllers.UserWebAPIV1Controller{}, "post:PostGravatar"),
			beego.NSRouter("/:username/password", &controllers.UserWebAPIV1Controller{}, "put:PutPassword"),
			beego.NSRouter("/:username/profile", &controllers.UserWebAPIV1Controller{}, "put:PutProfile"),
		),

		//repository routers
		beego.NSNamespace("/repository",
			beego.NSRouter("/:namespace/repositories", &controllers.RepoWebAPIV1Controller{}, "get:GetRepositories"),
			beego.NSRouter("/:namespace/:repository", &controllers.RepoWebAPIV1Controller{}, "post:PostRepository"),
			beego.NSRouter("/:namespace/:repository", &controllers.RepoWebAPIV1Controller{}, "put:PutRepository"),
			beego.NSRouter("/:namespace/:repository", &controllers.RepoWebAPIV1Controller{}, "get:GetRepository"),
			beego.NSRouter("/:namespace/:repository/collaborators", &controllers.RepoWebAPIV1Controller{}, "get:GetCollaborators"),
			beego.NSRouter("/:namespace/:repository/collaborators/:collaborator", &controllers.RepoWebAPIV1Controller{}, "post:PostCollaborator"),
			beego.NSRouter("/:namespace/:repository/collaborators/:collaborator", &controllers.RepoWebAPIV1Controller{}, "put:PutCollaborator"),
		),

		//organization routers
		beego.NSNamespace("/organization",
			beego.NSRouter("/:username/orgs", &controllers.OrganizationWebV1Controller{}, "get:GetOrgs"),
			beego.NSRouter("/:username/orgs/joins", &controllers.OrganizationWebV1Controller{}, "get:GetJoinOrgs"),
			beego.NSRouter("/:username/teams", &controllers.OrganizationWebV1Controller{}, "get:GetTeams"),
			beego.NSRouter("/:username/teams/joins", &controllers.OrganizationWebV1Controller{}, "get:GetJoinTeams"),
			beego.NSRouter("/:username/:org", &controllers.OrganizationWebV1Controller{}, "post:PostOrg"),
			beego.NSRouter("/:username/:org", &controllers.OrganizationWebV1Controller{}, "put:PutOrg"),
			beego.NSRouter("/:org", &controllers.OrganizationWebV1Controller{}, "get:GetOrg"),
			beego.NSRouter("/:org/repos/public", &controllers.OrganizationWebV1Controller{}, "get:GetPublicRepos"),
			beego.NSRouter("/:org/repos/private", &controllers.OrganizationWebV1Controller{}, "get:GetPrivateRepos"),
			beego.NSRouter("/:org/teams", &controllers.TeamWebV1Controller{}, "get:GetOrgTeams"),
			beego.NSRouter("/:username/:org/team", &controllers.TeamWebV1Controller{}, "post:PostTeam"),
			beego.NSRouter("/:username/:org/team/:team", &controllers.TeamWebV1Controller{}, "put:PutTeam"),
			beego.NSRouter("/:username/:org/team/:team", &controllers.TeamWebV1Controller{}, "get:GetTeam"),
			beego.NSRouter("/:username/:org/team/:team/:member", &controllers.TeamWebV1Controller{}, "post:PostMember"),
			beego.NSRouter("/:username/:org/team/:team/:member", &controllers.TeamWebV1Controller{}, "put:PutMember"),
		),
	)

	//Docker Registry API V1 remain
	beego.Router("/_ping", &controllers.PingAPIV1Controller{}, "get:GetPing")

	//Docker Registry API V1
	apiv1 := beego.NewNamespace("/v1",
		beego.NSRouter("/_ping", &controllers.PingAPIV1Controller{}, "get:GetPing"),
		beego.NSRouter("/users", &controllers.UserAPIV1Controller{}, "get:GetUsers"),
		beego.NSRouter("/users", &controllers.UserAPIV1Controller{}, "post:PostUsers"),

		beego.NSNamespace("/repositories",
			beego.NSRouter("/:namespace/:repo_name/tags/:tag", &controllers.RepoAPIV1Controller{}, "put:PutTag"),
			beego.NSRouter("/:namespace/:repo_name/images", &controllers.RepoAPIV1Controller{}, "put:PutRepositoryImages"),
			beego.NSRouter("/:namespace/:repo_name/images", &controllers.RepoAPIV1Controller{}, "get:GetRepositoryImages"),
			beego.NSRouter("/:namespace/:repo_name/tags", &controllers.RepoAPIV1Controller{}, "get:GetRepositoryTags"),
			beego.NSRouter("/:namespace/:repo_name", &controllers.RepoAPIV1Controller{}, "put:PutRepository"),
		),

		beego.NSNamespace("/images",
			beego.NSRouter("/:image_id/ancestry", &controllers.ImageAPIV1Controller{}, "get:GetImageAncestry"),
			beego.NSRouter("/:image_id/json", &controllers.ImageAPIV1Controller{}, "get:GetImageJSON"),
			beego.NSRouter("/:image_id/layer", &controllers.ImageAPIV1Controller{}, "get:GetImageLayer"),
			beego.NSRouter("/:image_id/json", &controllers.ImageAPIV1Controller{}, "put:PutImageJSON"),
			beego.NSRouter("/:image_id/layer", &controllers.ImageAPIV1Controller{}, "put:PutImageLayer"),
			beego.NSRouter("/:image_id/checksum", &controllers.ImageAPIV1Controller{}, "put:PutChecksum"),
		),
	)

	//Docker Registry API V2
	apiv2 := beego.NewNamespace("/v2",
		beego.NSRouter("/", &controllers.PingAPIV2Controller{}, "get:GetPing"),
		//Push
		beego.NSRouter("/:namespace/:repo_name/blobs/:digest", &controllers.BlobAPIV2Controller{}, "head:HeadDigest"),
		beego.NSRouter("/:namespace/:repo_name/blobs/uploads", &controllers.BlobAPIV2Controller{}, "post:PostBlobs"),
		beego.NSRouter("/:namespace/:repo_name/blobs/uploads/:uuid", &controllers.BlobAPIV2Controller{}, "put:PutBlobs"),
		beego.NSRouter("/:namespace/:repo_name/manifests/:tag", &controllers.ManifestsAPIV2Controller{}, "put:PutManifests"),
		//Pull
		beego.NSRouter("/:namespace/:repo_name/tags/list", &controllers.ManifestsAPIV2Controller{}, "get:GetTags"),
		beego.NSRouter("/:namespace/:repo_name/manifests/:tag", &controllers.ManifestsAPIV2Controller{}, "get:GetManifests"),
		beego.NSRouter("/:namespace/:repo_name/blobs/:digest", &controllers.BlobAPIV2Controller{}, "get:GetBlobs"),
	)

	//Dockerfile Build API V1
	buildv1 := beego.NewNamespace("/b1",
		beego.NSRouter("/build", &controllers.BuilderAPIV1Controller{}, "post:PostBuild"),
		beego.NSRouter("/status", &controllers.BuilderAPIV1Controller{}, "get:GetStatus"),
	)

	//Auth Fiters
	beego.InsertFilter("/w1/*", beego.BeforeRouter, filters.FilterDebug)
	beego.InsertFilter("/v1/*", beego.BeforeRouter, filters.FilterDebug)
	beego.InsertFilter("/v2/*", beego.BeforeRouter, filters.FilterDebug)
	beego.InsertFilter("/b1/*", beego.BeforeRouter, filters.FilterDebug)

	beego.InsertFilter("/v1/repositories/*", beego.BeforeRouter, filters.FilterAuth)
	beego.InsertFilter("/v1/images/*", beego.BeforeRouter, filters.FilterAuth)
	beego.InsertFilter("/v2/*", beego.BeforeRouter, filters.FilterAuth)

	beego.AddNamespace(web)
	beego.AddNamespace(apiv1)
	if beego.AppConfig.String("docker::API") == "v2" {
		beego.AddNamespace(apiv2)
	}

	beego.AddNamespace(buildv1)
}
