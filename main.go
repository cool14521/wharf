package main

import (
	"fmt"
	"os"

	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/containerops/wharf/cmd"
)

func init() {
	//
}

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
