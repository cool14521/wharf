package cmd

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/codegangsta/cli"
)

var CmdVersion = cli.Command{
	Name:        "version",
	Usage:       "当前运行的程序版本",
	Description: "当前运行的程序版本",
	Action:      runVersion,
}

func runVersion(c *cli.Context) {
	fmt.Println("当前版本: " + beego.AppConfig.String("version"))
	fmt.Println("Standalone 模式: " + beego.AppConfig.String("docker::Standalone"))
	fmt.Println("开放注册: " + beego.AppConfig.String("docker::OpenSignup"))
}
