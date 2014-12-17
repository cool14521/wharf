package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/dockercn/docker-bucket/markdown"
)

var CmdDoc = cli.Command{
	Name:        "category",
	Usage:       "通过命令同步或者获取文档数据",
	Description: "通过命令将远程Git库中存放的markdown文件同步到本地，并且将markdown格式转换成html后存入ledis中;通过命令可以查询同步之后的目录",
	Action:      runCategory,
	Flags: []cli.Flag{
		cli.StringFlag{"action", "", "输入操作的类型[sync 远程同步数据到本地;render 将文件数据处理后加入到缓存;save 将数据存入数据库中;query 查询(输入doc的前缀值，可查询doc目录;如果查询具体文件，请使用permalink参数)]", ""},
		cli.StringFlag{"remote", "", "输入同步远程库的地址,例如：https://github.com/xxx/xxx.git", ""},
		cli.StringFlag{"local", "", "输入远程库同步到本地的路径,例如：/root/data", ""},
		cli.StringFlag{"prefix", "", "输入同步类型的前缀名", ""},
		cli.StringFlag{"permalink", "", "输入查询文件的permalink值,例如：文件名xxx.md 则permalink值为xxx", ""},
	},
}

func runCategory(c *cli.Context) {
	action := strings.TrimSpace(c.String("action"))
	if len(action) == 0 {
		fmt.Println(errors.New("文档操作请输入action的值"))
		return
	}
	category := &markdown.Category{
		Remote: strings.TrimSpace(c.String("remote")),
		Local:  strings.TrimSpace(c.String("local")),
		Prefix: strings.TrimSpace(c.String("prefix")),
	}
	switch action {
	case "sync":
		if err := validate(category, "sync"); err != nil {
			fmt.Println(err)
			return
		} else if err := category.Sync(); err != nil {
			fmt.Println(err)
			return
		}
	case "render":
		if err := validate(category, "render"); err != nil {
			fmt.Println(err)
			return
		} else if err := category.Render(); err != nil {
			fmt.Println(err)
			return
		}
	case "save":
		if err := validate(category, "save"); err != nil {
			fmt.Println(err)
			return
		} else if err := category.Save(); err != nil {
			fmt.Println(err)
			return
		}
	case "query":
		if err := validate(category, "query"); err != nil {
			fmt.Println(err)
			return
		}
		if len(strings.TrimSpace(c.String("permalink"))) == 0 {
			if len(strings.TrimSpace(c.String("prefix"))) == 0 {
				fmt.Println(errors.New("请输入prefix的值"))
				return
			} else if _, err := category.Query(true); err != nil {
				fmt.Println(err)
			}
		} else if _, err := category.Query(false, strings.TrimSpace(c.String("permalink"))); err != nil {
			fmt.Println(err)
		}
	default:
		fmt.Println(errors.New("输入的action参数不存在"))
	}
}

func validate(category *markdown.Category, action string) error {
	switch action {
	case "sync":
		if len(strings.TrimSpace(category.Remote)) == 0 || len(strings.TrimSpace(category.Local)) == 0 {
			return errors.New("....markdown git地址初始化异常,请赋值remote和local")
		}
	case "render":
		if len(strings.TrimSpace(category.Local)) == 0 {
			return errors.New("....请输入local的值")
		} else if files, _ := ioutil.ReadDir(category.Local); len(files) == 0 {
			return errors.New("....本地路径不存在文件,无法进行转换处理，请执行sync操作,确认文件已经同步")
		}
	case "save":
		if _, err := os.Stat(".render"); err != nil || len(strings.TrimSpace(category.Prefix)) == 0 {
			return errors.New("....请确认是否值之前执行了sync、render的操作;检查prefix的值")
		}
	}
	return nil
}
