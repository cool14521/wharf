package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/dockercn/docker-bucket/markdown"
)

var CmdDoc = cli.Command{
	Name:        "doc",
	Usage:       "通过命令同步或者获取文档数据",
	Description: "通过命令将远程Git库中存放的markdown文件同步到本地，并且将markdown格式转换成html后存入ledis中;通过命令可以查询同步之后的目录",
	Action:      runDoc,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "action",
			Value: "",
			Usage: "输入操作的类型[sync 远程同步数据到本地;transform 将文件数据处理后加入到缓存;save 将数据存入数据库中;query 查询(输入doc的前缀值，可查询doc目录;如果查询具体文件，请使用key参数)]",
		},
		cli.StringFlag{
			Name:  "remote",
			Value: "",
			Usage: "输入同步远程库的地址,例如：https://github.com/xxx/xxx.git",
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
		cli.StringFlag{
			Name:  "key",
			Value: "",
			Usage: "输入查询文件的key值,例如：文件名xxx.md 则key值为xxx",
		},
	},
}

func runDoc(c *cli.Context) {
	action := strings.TrimSpace(c.String("action"))
	if len(action) == 0 {
		fmt.Println(errors.New("文档操作请输入action的值"))
		return
	}
	doc := &markdown.Doc{
		Remote:  strings.TrimSpace(c.String("remote")),
		Local:   strings.TrimSpace(c.String("local")),
		Prefix:  strings.TrimSpace(c.String("prefix")),
		Storage: strings.TrimSpace(c.String("storage")),
		Db:      c.Int("db"),
	}
	switch action {
	case "sync":
		doc.Sync()
	case "transform":
		doc.Transform()
	case "save":
		doc.Save()
	case "query":
		if len(strings.TrimSpace(c.String("key"))) == 0 {
			if len(strings.TrimSpace(c.String("prefix"))) == 0 {
				errors.New("请输入prefix的值")
				break
			}
			doc.Query(true, "")
		} else {
			doc.Query(false, strings.TrimSpace(c.String("key")))
		}
	default:
		fmt.Println(errors.New("输入的action参数不存在"))
	}
}
