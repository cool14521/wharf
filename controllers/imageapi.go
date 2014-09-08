package controllers

import "github.com/astaxie/beego"

type ImageAPIController struct {
	beego.Controller
}

func (i *ImageAPIController) URLMapping() {
	i.Mapping("GetImageJSON", i.GetImageJSON)
	i.Mapping("PutImageJson", i.PutImageJson)
	i.Mapping("PutImageLayer", i.PutImageLayer)
	i.Mapping("PutChecksum", i.PutChecksum)
	i.Mapping("GetImageAncestry", i.GetImageAncestry)
	i.Mapping("GetImageLayer", i.GetImageLayer)
}

func (this *ImageAPIController) Prepare() {

}

//在 Push 的流程中，docker 客户端会先调用 GET /v1/images/:image_id/json 向服务器检查是否已经存在 JSON 信息。
//如果存在了 JSON 信息，docker 客户端就认为是已经存在了 layer 数据，不再向服务器 PUT layer 的 JSON 信息和文件了。
//如果不存在 JSON 信息，docker 客户端会先后执行 PUT /v1/images/:image_id/json 和 PUT /v1/images/:image_id/layer 。
func (this *ImageAPIController) GetImageJSON() {

}

//向数据库写入 Layer 的 JSON 数据
//TODO: 检查 JSON 是否合法
func (this *ImageAPIController) PutImageJson() {

}

//向本地硬盘写入 Layer 的文件
func (this *ImageAPIController) PutImageLayer() {

}

func (this *ImageAPIController) PutChecksum() {

}

func (this *ImageAPIController) GetImageAncestry() {

}

func (this *ImageAPIController) GetImageLayer() {

}
