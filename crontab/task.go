package crontab

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/toolbox"
	"github.com/dockercn/docker-bucket/markdown"
)

func InitTask() {
	task := toolbox.NewTask("tk1", "0 0 */1 * * *", UpdateDocs)
	//toolbox.AddTask("tk1", task)
	//toolbox.StartTask()
	//defer toolbox.StopTask()
	task.Run()
}

func UpdateDocs() error {
	//每两个小时更新docs、documents目录
	category := new(markdown.Category)
	category.Local = beego.AppConfig.String("category::Local")
	category.Remote = beego.AppConfig.String("category::Remote")
	category.Prefix = beego.AppConfig.String("category::Prefix")
	if err := category.Sync(); err != nil {
		beego.Trace("Sync错误，err=", err)
		return err
	} else if err = category.Render(); err != nil {
		beego.Trace("Render错误，err=", err)
		return err
	} else if err = category.Save(); err != nil {
		beego.Trace("Save错误，err=", err)
		return err
	}
	beego.Error("同步更新文档完成......")
	return nil
}
