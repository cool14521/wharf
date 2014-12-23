package models

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/dockercn/docker-bucket/utils"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
	"reflect"
	"strings"
	"sync"
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

func Save(obj interface{}, key []byte) error {
	s := reflect.TypeOf(obj).Elem()

	//循环处理 Struct 的每一个 Field
	for i := 0; i < s.NumField(); i++ {
		//获取 Field 的 Value
		value := reflect.ValueOf(obj).Elem().Field(s.Field(i).Index[0])

		//判断 Field 不为空
		if utils.IsEmptyValue(value) == false {
			switch value.Kind() {
			case reflect.String:
				if _, err := LedisDB.HSet(key, []byte(s.Field(i).Name), []byte(value.String())); err != nil {
					return err
				}
			case reflect.Bool:
				if _, err := LedisDB.HSet(key, []byte(s.Field(i).Name), utils.BoolToBytes(value.Bool())); err != nil {
					return err
				}
			case reflect.Int64:
				if _, err := LedisDB.HSet(key, []byte(s.Field(i).Name), utils.Int64ToBytes(value.Int())); err != nil {
					return err
				}
			default:
				return fmt.Errorf("不支持的数据类型 %s:%s", s.Field(i).Name, value.Kind().String())
			}
		}

	}
	return nil
}
