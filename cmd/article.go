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
	markdown.GitAddress = "http://github.com/chliang2030598/docs1.git::data::A;http://github.com/chliang2030598/docs2.git::data::B"
	markdown.Run()
	markdown.ShowArticleList()
}
