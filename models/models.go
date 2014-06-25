package models

import (
	"fmt"
	"github.com/astaxie/beego"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"log"
)

var Engine *xorm.Engine

func setEngine() {
	host := beego.AppConfig.String("mysql::Host")
	name := beego.AppConfig.String("mysql::Name")
	user := beego.AppConfig.String("mysql::User")
	passwd := beego.AppConfig.String("mysql::Passwd")

	var err error
	conn := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8", user, passwd, host, name)
	beego.Trace("Initialized database connStr ->", conn)

	Engine, err = xorm.NewEngine("mysql", conn)
	if err != nil {
		log.Fatalf("models.init -> fail to conntect database: %v", err)
	}

	Engine.ShowDebug = true
	Engine.ShowErr = true
	Engine.ShowSQL = true

	beego.Trace("Initialized database ->", name)

}

// InitDb initializes the database.
func InitDb() {
	setEngine()
	err := Engine.Sync(new(User), new(Profile), new(Organization), new(Member), new(Image), new(Repository), new(Tag), new(Comment), new(Star), new(History))
	if err != nil {
		log.Fatalf("models.init -> fail to sync database: %v", err)
	}
}
