package cmd

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/dockercn/docker-bucket/models"
)

var CmdEmail = cli.Command{
	Name:        "email",
	Usage:       "通过命令完成邮件模块的初始化",
	Description: "通过命令可以管理邮件服务器设置，邮件模板设置，邮件内容设置",
	Action:      runEmail,
	Flags: []cli.Flag{
		cli.StringFlag{"object", "", "选择设置的对象[server 邮件服务器;template 邮件模板; message 邮件信息", ""},
		cli.StringFlag{"action", "", "选择操作类型[add 添加;del 删除;update 更新;query 查询] 注：message只提供add和query操作", ""},
		cli.StringFlag{"host", "", "输入邮件服务器的地址", ""},
		cli.IntFlag{"port", 0, "输入邮件服务器的端口", ""},
		cli.StringFlag{"user", "", "输入邮件服务器的用户名", ""},
		cli.StringFlag{"password", "", "输入邮件服务器的密码", ""},
		cli.StringFlag{"prefix", "", "邮件模板的前缀名", ""},
		cli.StringFlag{"path", "", "前缀模板的路径", ""},
		cli.StringFlag{"to", "", "收件人", ""},
		cli.StringFlag{"cc", "", "抄送", ""},
		cli.StringFlag{"bcc", "", "密送", ""},
		cli.StringFlag{"from", "", "发件人", ""},
		cli.StringFlag{"subject", "", "主题", ""},
		cli.StringFlag{"body", "", "邮件内容", ""},
		cli.StringFlag{"type", "", "邮件内容类型[注：html代表html类型 text代表文本类型]", ""},
	},
}

func runEmail(c *cli.Context) {
	object := strings.TrimSpace(c.String("object"))
	action := strings.TrimSpace(c.String("action"))
	if len(object) == 0 {
		log.Fatalln("未选择处理的对象，请输入object的值")
	} else if len(action) == 0 {
		log.Fatalln("未选择对象操作类型，请输入action的值")
	}
	models.InitDb()
	switch object {
	case "server":
		switch action {
		case "add":
			//对输入参数进行验证，看是否符合要求
			if err := validate4email(c.String("host"), c.String("user"), c.String("password")); err != nil {
				log.Fatalln(err)
			} else if c.Int("port") == 0 {
				log.Fatalln("邮件服务器端口未设置，请对port参数赋值")
			}
			mailServer := new(models.MailServer)
			if err := mailServer.Add(c.String("host"), c.Int("port"), c.String("user"), c.String("password")); err != nil {
				log.Fatalln(err)
			}
			log.Println("邮件服务器添加成功")
		case "del":
			if err := validate4email(c.String("host")); err != nil {
				log.Fatalln(err)
			}
			mailServer := new(models.MailServer)
			if err := mailServer.Delete(c.String("host")); err != nil {
				log.Fatalln(err)
			}
			log.Println("邮件服务器删除成功")
		case "update":
			if err := validate4email(c.String("host"), c.String("user"), c.String("password")); err != nil {
				log.Fatalln(err)
			} else if c.Int("port") == 0 {
				log.Fatalln("邮件服务器端口未设置，请对port参数赋值")
			}
			mailServer := new(models.MailServer)
			if err := mailServer.Add(c.String("host"), c.Int("port"), c.String("user"), c.String("password")); err != nil {
				log.Fatalln(err)
			}
			log.Println("邮件服务器更新成功")
		case "query":
			//如果host为空，则显示全部列表
			mailServer := new(models.MailServer)
			if len(strings.TrimSpace(c.String("host"))) == 0 {
				mailServers := mailServer.Query()
				fmt.Printf("%#v\n", mailServers)
				log.Println("邮件服务器查询完成")
				return
			}
			mailServers := mailServer.Query(c.String("host"))
			fmt.Printf("%#v\n", mailServers[0])
			log.Println("邮件服务器查询完成")
		default:
			log.Fatalln("输入的action值非法，无法执行")
		}
	case "template":
		switch action {
		case "add":
			if err := validate4email(c.String("path"), c.String("prefix")); err != nil {
				log.Fatalln(err)
			}
			tmpl := new(models.TemplateHtml)
			if err := tmpl.Add(c.String("prefix"), c.String("path")); err != nil {
				log.Fatalln(err)
			}
			log.Println("邮件模板添加成功")
		case "del":
			if err := validate4email(c.String("prefix")); err != nil {
				log.Fatalln(err)
			}
			tmpl := new(models.TemplateHtml)
			if err := tmpl.Delete(c.String("prefix")); err != nil {
				log.Fatalln(err)
			}
			log.Println("删除模板成功")
		case "update":
			if err := validate4email(c.String("path"), c.String("prefix")); err != nil {
				log.Fatalln(err)
			}
			tmpl := new(models.TemplateHtml)
			if err := tmpl.Add(c.String("prefix"), c.String("path")); err != nil {
				log.Fatalln(err)
			}
			log.Println("邮件模板更新成功")
		case "query":
			tmpl := new(models.TemplateHtml)
			if len(strings.TrimSpace(c.String("prefix"))) == 0 {
				tmpls := tmpl.Query()
				fmt.Printf("%#v\n", tmpls)
				log.Println("模板查询成功")
				return
			}
			tmpls := tmpl.Query(c.String("prefix"))
			fmt.Printf("%#v\n", tmpls[0])
			log.Println("邮件模板查询成功")
		default:
			log.Fatalln("输入的action值非法，无法执行")
		}
	case "message":
		switch action {
		case "add":
			cc := strings.Split(c.String("cc"), ",")
			bcc := strings.Split(c.String("bcc"), ",")
			if err := validate4email(c.String("to"), c.String("from"), c.String("type"), c.String("host"), c.String("prefix")); err != nil {
				log.Fatalln(err)
			}
			msg := new(models.Message)
			if err := msg.Add(c.String("to"), c.String("from"), "测试中文主题", c.String("body"), c.String("type"), c.String("prefix"), c.String("host"), cc, bcc); err != nil {
				log.Fatalln(err)
			}
			log.Println("信息添加成功")
		case "query":
			msg := new(models.Message)
			if err := validate4email(c.String("prefix")); err != nil {
				log.Fatalln(err)
			}
			msgs := msg.Query(c.String("prefix"))
			fmt.Printf("%#v\n", msgs)
			log.Println("信息查询完成")
		default:
			log.Fatalln("输入的action值非法，无法执行")
		}
	default:
		log.Fatalln("输入的object值非法，无法执行")
	}
}

func validate4email(attrs ...string) error {
	for _, attr := range attrs {
		if len(strings.TrimSpace(attr)) == 0 {
			return errors.New(fmt.Sprint(attr, "的值不能为空"))
		}
	}
	return nil
}
