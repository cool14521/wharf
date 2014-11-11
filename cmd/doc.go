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

var CmdDoc = cli.Command{
	Name:        "doc",
	Usage:       "通过命令同步或者获取文档数据",
	Description: "通过命令同步或者获取文档数据",
	Action:      runDoc,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "action",
			Value: "",
			Usage: "输入操作的类型[save 存储;show 展示]",
		},
		cli.StringFlag{
			Name:  "remote",
			Value: "",
			Usage: "输入同步远程库的地址,例如：http://github.com/chliang2030598/docs1.git",
		},
		cli.StringFlag{
			Name:  "storage",
			Value: "/data/ledis",
			Usage: "输入ledis数据库所在路径,默认：/data/ledis",
		},
		cli.IntFlag{
			Name:  "db",
			Value: 1,
			Usage: "文档存储的数据库，默认：1",
		},
		cli.StringFlag{
			Name:  "local",
			Value: "",
			Usage: "输入远程库同步到本地的路径,例如：/root/data",
		},
		cli.StringFlag{
			Name:  "prefix",
			Value: "",
			Usage: "输入同步类型的前缀名",
		},
	},
}

func runDoc(c *cli.Context) {
	action := strings.TrimSpace(c.String("action"))
	if len(action) == 0 {
		fmt.Println(errors.New("文档操作请输入action的值"))
		return
	}
	switch action {
	case "save":
		if len(strings.TrimSpace(c.String("remote"))) == 0 || len(strings.TrimSpace(c.String("local"))) == 0 || len(strings.TrimSpace(c.String("prefix"))) == 0 {
			fmt.Println(errors.New("save必须输入remote、local、prefix的值"))
			break
		}
		markdown.Remote = strings.TrimSpace(c.String("remote"))
		markdown.Local = strings.TrimSpace(c.String("local"))
		markdown.Prefix = strings.TrimSpace(c.String("prefix"))
		markdown.Storage = strings.TrimSpace(c.String("storage"))
		markdown.Db = c.Int("db")
		markdown.Run()
	case "show":
		if len(strings.TrimSpace(c.String("prefix"))) == 0 {
			errors.New("请输入prefix的值")
			break
		}
		markdown.Storage = strings.TrimSpace(c.String("storage"))
		markdown.Db = c.Int("db")
		markdown.Show(strings.TrimSpace(c.String("prefix")))
	default:
		fmt.Println(errors.New("输入的action参数不存在"))
	}
}
