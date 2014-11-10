package cmd

import (
	"errors"
	"fmt"
	"strings"

	_ "github.com/astaxie/beego/session/redis"
	"github.com/codegangsta/cli"
	"github.com/dockercn/docker-bucket/markdown"
	_ "github.com/dockercn/docker-bucket/routers"
)

var CmdLedis = cli.Command{
	Name:        "ledis",
	Usage:       "获取文章列表",
	Description: "通过命令存取ledis中的数据",
	Action:      runLedis,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "action",
			Value: "",
			Usage: "输入操作的类型[save 存储;show 展示]",
		},
		cli.StringFlag{
			Name:  "gitAddress",
			Value: "",
			Usage: "输入同步远程库的地址,例如：http://github.com/chliang2030598/docs1.git",
		},
		cli.StringFlag{
			Name:  "dbaddress",
			Value: "/data/ledis?select=1",
			Usage: "输入ledis数据库所在路径,默认：/data/ledis?select=1",
		},
		cli.StringFlag{
			Name:  "local",
			Value: "",
			Usage: "输入远程库同步到本地的路径,例如：/root/data",
		},
		cli.StringFlag{
			Name:  "tag",
			Value: "",
			Usage: "输入同步类型的分类",
		},
	},
}

func runLedis(c *cli.Context) {
	action := strings.TrimSpace(c.String("action"))
	if len(action) == 0 {
		fmt.Println(errors.New("启动ledis请输入action的值"))
		return
	}
	switch action {
	case "save":
		if len(strings.TrimSpace(c.String("gitAddress"))) == 0 || len(strings.TrimSpace(c.String("local"))) == 0 || len(strings.TrimSpace(c.String("tag"))) == 0 {
			fmt.Println(errors.New("save必须输入gitAddress、local、tag的值"))
			break
		}
		markdown.GitAddress = strings.TrimSpace(c.String("gitAddress"))
		markdown.Local = strings.TrimSpace(c.String("local"))
		markdown.Tag = strings.TrimSpace(c.String("tag"))
		markdown.DbAddress = strings.TrimSpace(c.String("dbaddress"))
		markdown.Run()
	case "show":
		if len(strings.TrimSpace(c.String("tag"))) == 0 {
			errors.New("请输入tag的值")
			break
		}
		markdown.DbAddress = strings.TrimSpace(c.String("dbaddress"))
		markdown.Show(strings.TrimSpace(c.String("tag")))
	default:
		fmt.Println(errors.New("输入的action参数不存在"))
	}
}
