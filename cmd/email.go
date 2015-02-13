package cmd

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/codegangsta/cli"

	"github.com/dockercn/wharf/models"
)

var CmdEmail = cli.Command{
	Name:        "email",
	Usage:       "email and email template configuration through CLI",
	Description: "Initlization email configuration and manage email template content through CLI",
	Action:      runEmail,
	Flags: []cli.Flag{
		cli.StringFlag{"object", "", "[server: Email Server Config; template: Email Template; message: Email Content", ""},
		cli.StringFlag{"action", "", "[add/del/update/query] Tip: message object only support add and query action.", ""},
		cli.StringFlag{"host", "", "Email server, like：smtp.exmail.qq.com", ""},
		cli.IntFlag{"port", 0, "Email server port, like 443", ""},
		cli.StringFlag{"user", "", "Email account", ""},
		cli.StringFlag{"password", "", "Email account passwd", ""},
		cli.StringFlag{"prefix", "", "Email tempalte prrefix", ""},
		cli.StringFlag{"path", "", "Email template prefix template", ""},
		cli.StringFlag{"to", "", "Email to", ""},
		cli.StringFlag{"cc", "", "Email cc", ""},
		cli.StringFlag{"bcc", "", "Email bcc", ""},
		cli.StringFlag{"from", "", "Email send from", ""},
		cli.StringFlag{"subject", "", "Email subject", ""},
		cli.StringFlag{"body", "", "Email body", ""},
		cli.StringFlag{"type", "", "Email type [html/text]", ""},
	},
}

func runEmail(c *cli.Context) {
	object := strings.TrimSpace(c.String("object"))
	action := strings.TrimSpace(c.String("action"))

	if len(object) == 0 {
		log.Fatalln("Please input --objcet value")
	} else if len(action) == 0 {
		log.Fatalln("Please input --action value")
	}

	models.InitDb()

	switch object {

	case "server":

		switch action {
		case "add":
			if err := validate4email(c.String("host"), c.String("user"), c.String("password")); err != nil {
				log.Fatalln(err)
			} else if c.Int("port") == 0 {
				log.Fatalln("Please input the --port value")
			}

			mailServer := new(models.MailServer)

			if err := mailServer.Add(c.String("host"), c.Int("port"), c.String("user"), c.String("password")); err != nil {
				log.Fatalln(err)
			}

			log.Println("Add email server config successfully")

			break

		case "del":
			if err := validate4email(c.String("host")); err != nil {
				log.Fatalln(err)
			}

			mailServer := new(models.MailServer)

			if err := mailServer.Delete(c.String("host")); err != nil {
				log.Fatalln(err)
			}

			log.Println("Del email server config successfully")

			break

		case "update":
			if err := validate4email(c.String("host"), c.String("user"), c.String("password")); err != nil {
				log.Fatalln(err)
			} else if c.Int("port") == 0 {
				log.Fatalln("Please input the --port value")
			}

			mailServer := new(models.MailServer)

			if err := mailServer.Add(c.String("host"), c.Int("port"), c.String("user"), c.String("password")); err != nil {
				log.Fatalln(err)
			}

			log.Println("Update email server config successfully")

			break

		case "query":
			if len(strings.TrimSpace(c.String("host"))) == 0 {
				log.Fatalln("Please input the --host value")
			}

			mail := new(models.MailServer)
			servers := mail.Query(c.String("host"))

			fmt.Printf("%v\n", servers[0])

			break

		default:
			log.Fatalln("Illegal --action value")
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

			log.Println("Add email template config successfully")

			break

		case "del":
			if err := validate4email(c.String("prefix")); err != nil {
				log.Fatalln(err)
			}

			tmpl := new(models.TemplateHtml)

			if err := tmpl.Delete(c.String("prefix")); err != nil {
				log.Fatalln(err)
			}

			log.Println("Del email template config successfully")

			break

		case "update":
			if err := validate4email(c.String("path"), c.String("prefix")); err != nil {
				log.Fatalln(err)
			}

			tmpl := new(models.TemplateHtml)

			if err := tmpl.Add(c.String("prefix"), c.String("path")); err != nil {
				log.Fatalln(err)
			}

			log.Println("Update email template config successfully")

			break

		case "query":
			if len(strings.TrimSpace(c.String("prefix"))) == 0 {
				log.Fatalln("Please input the --prefix value")
			}

			tmpl := new(models.TemplateHtml)
			tmpls := tmpl.Query(c.String("prefix"))

			fmt.Printf("%v\n", tmpls[0])

			break

		default:
			log.Fatalln("Illegal --action value")
		}

	case "message":

		switch action {
		case "add":
			cc := strings.Split(c.String("cc"), ",")
			bcc := strings.Split(c.String("bcc"), ",")
			if err := validate4email(c.String("to"), c.String("from"), c.String("type"), c.String("host"), c.String("prefix")); err != nil {
				log.Fatalln(err)
			}

			tmpl := new(models.TemplateHtml)
			if tmpls := tmpl.Query(c.String("prefix")); len(tmpls) == 0 {
				log.Fatalln(errors.New("Please add email template first"))
			}

			server := new(models.MailServer)
			if servers := server.Query(c.String("host")); len(servers) == 0 {
				log.Fatalln(errors.New("Please add email server first"))
			}

			msg := new(models.Message)
			if err := msg.Add(c.String("to"), c.String("from"), "Testing UTF-8 Subject And Email", c.String("body"), c.String("type"), c.String("prefix"), c.String("host"), cc, bcc); err != nil {
				log.Fatalln(err)
			}

			log.Println("Add email successfully")

			break

		case "query":
			if err := validate4email(c.String("prefix")); err != nil {
				log.Fatalln(err)
			}

			msg := new(models.Message)
			msgs := msg.Query(c.String("prefix"))

			for _, vmsg := range msgs {
				fmt.Printf("%v\n", vmsg)
			}

		default:
			log.Fatalln("Illegal --action value")
		}
	default:
		log.Fatalln("Illegal --object value")
	}
}

func validate4email(attrs ...string) error {
	for _, attr := range attrs {
		if len(strings.TrimSpace(attr)) == 0 {
			return errors.New(fmt.Sprint(attr, " value is not be null"))
		}
	}
	return nil
}
