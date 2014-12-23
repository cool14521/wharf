/*

*/

package controllers

import (
	"github.com/astaxie/beego"
	"github.com/dockercn/docker-bucket/markdown"
	"strings"
)

type ArticleController struct {
	beego.Controller
}

func (this *ArticleController) GetArticle() {
	//读取参数
	pernalink := this.Ctx.Input.Param(":article")
	//加载markdown文件
	category := new(markdown.Category)
	docs, err := category.Query(false, pernalink)
	if err != nil {
		this.Abort("401")
		return
	}
	this.TplNames = "article.html"
	this.Data["content"] = docs[0].Content
	this.Data["title"] = docs[0].Title
	this.Data["desc"] = docs[0].Desc
	this.Data["author"] = docs[0].Author
	this.Data["tags"] = strings.Split(docs[0].Tags, ",")
	this.Data["views"] = docs[0].Views
	this.Data["updated"] = docs[0].Updated
	this.Render()
	return
}
