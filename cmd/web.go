package cmd

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/codegangsta/cli"

	"github.com/Unknwon/macaron"

	crew "github.com/containerops/crew/web"
	dockyard "github.com/containerops/dockyard/web"
	generator "github.com/containerops/generator/web"
	"github.com/containerops/wharf/setting"
	"github.com/containerops/wharf/web"
	"github.com/containerops/wrench/db"
	"github.com/containerops/wrench/utils"
)

var CmdWeb = cli.Command{
	Name:        "web",
	Usage:       "start wharf web service",
	Description: "wharf is the module of handler docker repository and rkt image.",
	Action:      runWeb,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "address",
			Value: "0.0.0.0",
			Usage: "web service listen ip, default is 0.0.0.0; if listen with Unix Socket, the value is sock file path.",
		},
		cli.IntFlag{
			Name:  "port",
			Value: 80,
			Usage: "web service listen at port 80; if run with https will be 443.",
		},
	},
}

func runWeb(c *cli.Context) {
	m := macaron.New()

	if err := db.InitDB(setting.DBURI, setting.DBPasswd, setting.DBDB); err != nil {
		fmt.Printf("Connect Database Error %s", err.Error())
	}

	web.SetWharfMacaron(m)
	dockyard.SetDockyardMacaron(m)
	crew.SetCrewMacaron(m)
	generator.SetGeneratorMacaron(m)

	switch setting.ListenMode {
	case "http":
		listenaddr := fmt.Sprintf("%s:%d", c.String("address"), c.Int("port"))
		if err := http.ListenAndServe(listenaddr, m); err != nil {
			fmt.Printf("start wharf http service error: %v", err.Error())
		}
		break
	case "https":
		listenaddr := fmt.Sprintf("%s:443", c.String("address"))
		server := &http.Server{Addr: listenaddr, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS10}, Handler: m}
		if err := server.ListenAndServeTLS(setting.HttpsCertFile, setting.HttpsKeyFile); err != nil {
			fmt.Printf("start wharf https service error: %v", err.Error())
		}
		break
	case "unix":
		listenaddr := fmt.Sprintf("%s", c.String("address"))
		if utils.Exist(listenaddr) {
			os.Remove(listenaddr)
		}

		if listener, err := net.Listen("unix", listenaddr); err != nil {
			fmt.Printf("start wharf unix socket error: %v", err.Error())
		} else {
			server := &http.Server{Handler: m}
			if err := server.Serve(listener); err != nil {
				fmt.Printf("start wharf unix socket error: %v", err.Error())
			}
		}
		break
	default:
		break
	}
}
