package models

import (
	"fmt"
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
}

func InitSession() {
	setSessionEngine()
}

// InitDb initializes the database.
func InitDb() {
	initLedisFunc := func() {
		cfg := new(config.Config)
		cfg.DataDir = beego.AppConfig.String("ledisdb::DataDir")
		var err error
		nowLedis, err = ledis.Open(cfg)
		if err != nil {
			println(err.Error())
			panic(err)
		}
	}
	ledisOnce.Do(initLedisFunc)
	db, _ := beego.AppConfig.Int("ledisdb::DB")
	LedisDB, _ = nowLedis.Select(db)
}

//获取服务器全局存储的 Key 值
func GetServerKeys(object string) string {
	switch strings.TrimSpace(object) {
	case "user":
		return fmt.Sprintf("%susers", USER_SYMBLE)
	case "org":
		return fmt.Sprintf("%sorgs", ORGANIZATION_SYMBLE)
	case "repo":
		return fmt.Sprintf("%srepos", REPOSITORY_SYMBLE)
	case "image":
		return fmt.Sprintf("%simages", IMAGE_SYMBLE)
	case "template":
		return fmt.Sprintf("%stemplates", TEMPLATE_SYMBLE)
	case "job":
		return fmt.Sprintf("%sjob", JOB_SYMBLE)
	default:
		return ""
	}
}

//获取对象存储的 Key
func GetObjectKey(object string, id string) string {
	switch strings.TrimSpace(object) {
	case "user":
		return fmt.Sprintf("%s%s", USER_SYMBLE, strings.TrimSpace(id))
	case "org":
		return fmt.Sprintf("%s%s", ORGANIZATION_SYMBLE, strings.TrimSpace(id))
	case "repo":
		return fmt.Sprintf("%s%s", REPOSITORY_SYMBLE, strings.TrimSpace(id))
	case "image":
		return fmt.Sprintf("%s%s", IMAGE_SYMBLE, strings.TrimSpace(id))
	case "template":
		return fmt.Sprintf("%s%s", TEMPLATE_SYMBLE, strings.TrimSpace(id))
	case "job":
		return fmt.Sprintf("%s%s", JOB_SYMBLE, strings.TrimSpace(id))
	default:
		return ""
	}
}
