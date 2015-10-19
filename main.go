package main

import (
	"os"

	"github.com/codegangsta/cli"

	"github.com/containerops/wharf/cmd"
	"github.com/containerops/wrench/setting"
)

func main() {
	setting.SetConfig("conf/containerops.conf")

	app := cli.NewApp()

	app.Name = setting.AppName
	app.Usage = setting.Usage
	app.Version = setting.Version
	app.Author = setting.Author
	app.Email = setting.Email

	app.Commands = []cli.Command{
		cmd.CmdWeb,
	}

	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Run(os.Args)
}
