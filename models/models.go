package models

import (
	"strings"
	"sync"

	"github.com/astaxie/beego"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
)

const (
	USER_SYMBLE         = "@"
	ORGANIZATION_SYMBLE = "#"
	REPOSITORY_SYMBLE   = "$"
	IMAGE_SYMBLE        = "&"
	TEMPLATE_SYMBLE     = "*"
	JOB_SYMBLE          = "!"
)

var (
	ledisOnce sync.Once
	nowLedis  *ledis.Ledis
	LedisDB   *ledis.DB
)

func setSessionEngine() {
	beego.SessionProvider = beego.AppConfig.String("session::Provider")
	beego.SessionSavePath = beego.AppConfig.String("session::SavePath")

	switch beego.AppConfig.String("docker::RunMode") {
	case "Bucket":
		beego.SessionName = "Bucket"
	case "Registry":
		beego.SessionName = beego.AppConfig.String("docker::endpoint")
	default:
		beego.SessionName = "Bucket"
	}

	beego.SessionHashKey = "dwzemsxoltmv"
}

func InitSession() {
	setSessionEngine()
}

// InitDb initializes the database.
func InitDb() {
	initLedisFunc := func() {
		cfg := new(config.Config)
		//cfg.DBName = "docker"
		cfg.DataDir = beego.AppConfig.String("ledisdb::DataDir")

		var err error
		nowLedis, err = ledis.Open(cfg)
		if err != nil {
			println(err.Error())
			panic(err)
		}
	}
	ledisOnce.Do(initLedisFunc)
	LedisDB, _ = nowLedis.Select(0)
}

func GetObjectKey(object string, id string) string {
	switch strings.TrimSpace(object) {
	case "user":
		return USER_SYMBLE + strings.TrimSpace(id)
	case "org":
		return ORGANIZATION_SYMBLE + strings.TrimSpace(id)
	case "repo":
		return REPOSITORY_SYMBLE + strings.TrimSpace(id)
	case "image":
		return IMAGE_SYMBLE + strings.TrimSpace(id)
	case "template":
		return TEMPLATE_SYMBLE + strings.TrimSpace(id)
	case "job":
		return JOB_SYMBLE + strings.TrimSpace(id)
	default:
		return ""
	}
}
