package cmd

import (
	"fmt"

	"github.com/codegangsta/cli"

	"github.com/dockercn/wharf/models"
)

var CmdAccount = cli.Command{
	Name:        "account",
	Usage:       "Manage account through CLI",
	Description: "Manage account in the wharf throught CLI.",
	Action:      runAccount,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "action",
			Value: "",
			Usage: "Action[add/active/unactive/log]",
		},
		cli.StringFlag{
			Name:  "email",
			Value: "",
			Usage: "Account's email",
		},
		cli.StringFlag{
			Name:  "username",
			Value: "",
			Usage: "Account's username",
		},
		cli.StringFlag{
			Name:  "passwd",
			Value: "",
			Usage: "Account's passwd",
		},
	},
}

func runAccount(c *cli.Context) {
	var action, email, username, passwd string

	if len(c.String("action")) > 0 {
		models.InitDb()
		action = c.String("action")
		switch action {
		case "add":
			if len(c.String("username")) > 0 && len(c.String("email")) > 0 && len(c.String("passwd")) > 0 {
				username = c.String("username")
				email = c.String("email")
				passwd = c.String("passwd")

				user := new(models.User)
				user.Username = username
				user.Password = passwd
				user.Email = email
				if err := user.Save(); err != nil {
					fmt.Println(fmt.Sprintf("Add user failure: %s", err.Error()))
				} else {
					fmt.Println(fmt.Sprintf("Add user successful: %s", username))
				}

			} else {
				fmt.Println("account add need username/email/passwd params")
			}

			break
		case "active":
			break
		case "unactive":
			break
		case "log":
			break
		default:
			fmt.Println("account only support add/active/unactive actions")
		}
	} else {
		fmt.Println("account only support add/active/unactive actions")
	}
}
