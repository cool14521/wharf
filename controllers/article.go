/*

*/

package controllers

import (
	"github.com/astaxie/beego"
	"github.com/dockercn/docker-bucket/markdown"
)

type ArticleController struct {
	beego.Controller
}

func (this *ArticleController) GetArticle() {
	//加载markdown文件
	category := new(markdown.Category)
	docs, err := category.Query(false, "private-docker-registry-with-nginx")
	if err != nil {
		beego.Trace(err)
	}
	this.TplNames = "article.html"
	this.Data["content"] = docs[0].Content
	this.Render()
}
