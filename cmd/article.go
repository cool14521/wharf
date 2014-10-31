package cmd

import (
	_ "github.com/astaxie/beego/session/redis"
	"github.com/codegangsta/cli"

	"github.com/dockercn/docker-bucket/markdown"
	_ "github.com/dockercn/docker-bucket/routers"
)

var CmdArticle = cli.Command{
	Name:        "article",
	Usage:       "获取文章列表",
	Description: "通过命令获取所有文章列表",
	Action:      runArticle,
	Flags:       []cli.Flag{},
}

func runArticle(c *cli.Context) {
	//设定 HTTP 的静态文件处理地址
	markdown.InitTask()
	markdown.ShowArticleList()

}
