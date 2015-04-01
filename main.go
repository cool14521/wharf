package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/astaxie/beego"
	"github.com/codegangsta/cli"

	"github.com/containerops/wharf/cmd"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	beego.SetLogger("file", fmt.Sprintf("{\"filename\":\"%s/%s.log\"}", beego.AppConfig.String("log::FilePath"), beego.AppConfig.String("log::FileName")))

	app := cli.NewApp()
	app.Name = beego.AppConfig.String("appname")
	app.Usage = beego.AppConfig.String("usage")
	app.Version = beego.AppConfig.String("version")
	app.Author = "Meaglit Ma"
	app.Email = "genedna@gmail.com"

	app.Commands = []cli.Command{
		cmd.CmdWeb,
	}

	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Run(os.Args)
}
