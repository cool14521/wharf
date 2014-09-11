package models

import (
	"github.com/astaxie/beego"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
	"sync"
)

var (
	ledisOnce sync.Once
	nowLedis  *ledis.Ledis
	LedisDB   *ledis.DB
)

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
