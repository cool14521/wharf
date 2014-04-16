package controllers

import (
  "github.com/astaxie/beego"
)

type ImageController struct {
  beego.Controller
}

func (this *ImageController) Prepare() {
  this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Version", beego.AppConfig.String("Version"))
  this.Ctx.Output.Context.ResponseWriter.Header().Set("X-Docker-Registry-Standalone", beego.AppConfig.String("Standalone"))
}

// http://docs.docker.io/en/latest/reference/api/registry_api/#layer
// GET /v1/images/(image_id)/layer
// Get image layer for a given image_id
// 读取 image 的 layer ：local 存储时读取 /registry/images/(image_id)/layer 文件；第三方存储时重定向到第三方存储文件的 URL。
// Example Request:
//    GET /v1/images/088b4505aa3adc3d35e79c031fa126b403200f02f51920fbd9b7c503e87c7a2c/layer HTTP/1.1
//    Host: registry-1.docker.io
//    Accept: application/json
//    Content-Type: application/json
// Example Response:
//    HTTP/1.1 200
//    Vary: Accept
//    X-Docker-Registry-Version: 0.6.0
//    Cookie: (Cookie provided by the Registry)
//    {layer binary data stream}
// Status Codes: 
//    200 – OK
//    401 – Requires authorization
//    404 – Image not found
func (this *ImageController) GETLayer() {

}

// http://docs.docker.io/en/latest/reference/api/registry_api/#layer
// Put image layer for a given image_id
// 写入 image 的 layer ：local 存储时写入 /registry/images/(image_id)/layer 文件；第三方存储时保存到第三方的存储空间。
// Example Request:
//    PUT /v1/images/088b4505aa3adc3d35e79c031fa126b403200f02f51920fbd9b7c503e87c7a2c/layer HTTP/1.1
//    Host: registry-1.docker.io
//    Transfer-Encoding: chunked
//    Authorization: Token signature=123abc,repository="foo/bar",access=write
//    {layer binary data stream}
// Example Response:
//    HTTP/1.1 200
//    Vary: Accept
//    Content-Type: application/json
//    X-Docker-Registry-Version: 0.6.0
//    ""
// Status Codes: 
//    200 – OK
//    401 – Requires authorization
//    404 – Image not found
func (this *ImageController) PUTLayer() {

}

// http://docs.docker.io/en/latest/reference/api/registry_api/#images
// Gut image for a given image_id
// 读取 image 的信息，读取 /registry/images/(image_id)/json 文件的内容
// Example Request:
//    GET /v1/images/088b4505aa3adc3d35e79c031fa126b403200f02f51920fbd9b7c503e87c7a2c/json HTTP/1.1
//    Host: registry-1.docker.io
//    Accept: application/json
//    Content-Type: application/json
//    Cookie: (Cookie provided by the Registry)
// Example Response:
//    HTTP/1.1 200
//    Vary: Accept
//    Content-Type: application/json
//    X-Docker-Registry-Version: 0.6.0
//    X-Docker-Size: 456789
//    X-Docker-Checksum: b486531f9a779a0c17e3ed29dae8f12c4f9e89cc6f0bc3c38722009fe6857087
//    {
//      id: "088b4505aa3adc3d35e79c031fa126b403200f02f51920fbd9b7c503e87c7a2c",
//     parent: "aeee6396d62273d180a49c96c62e45438d87c7da4a5cf5d2be6bee4e21bc226f",
//     created: "2013-04-30T17:46:10.843673+03:00",
//     container: "8305672a76cc5e3d168f97221106ced35a76ec7ddbb03209b0f0d96bf74f6ef7",
//     container_config: {
//         Hostname: "host-test",
//         User: "",
//         Memory: 0,
//         MemorySwap: 0,
//         AttachStdin: false,
//         AttachStdout: false,
//         AttachStderr: false,
//         PortSpecs: null,
//         Tty: false,
//         OpenStdin: false,
//         StdinOnce: false,
//         Env: null,
//         Cmd: [
//           "/bin/bash",
//           "-c",
//           "apt-get -q -yy -f install libevent-dev"
//         ],
//         Dns: null,
//         Image: "imagename/blah",
//         Volumes: { },
//         VolumesFrom: ""
//     },
//     docker_version: "0.1.7"
// }
func (this *ImageController) GETJSON() {

}

// http://docs.docker.io/en/latest/reference/api/registry_api/#images
// PUT /v1/images/(image_id)/json
// Put image for a given image_id
// 写入 image 的信息，写入到 /registry/images/(image_id)/json 文件
// Example Request:
//    PUT /v1/images/088b4505aa3adc3d35e79c031fa126b403200f02f51920fbd9b7c503e87c7a2c/json HTTP/1.1
//    Host: registry-1.docker.io
//    Accept: application/json
//    Content-Type: application/json
//    Cookie: (Cookie provided by the Registry)
//    {
//      id: "088b4505aa3adc3d35e79c031fa126b403200f02f51920fbd9b7c503e87c7a2c",
//      parent: "aeee6396d62273d180a49c96c62e45438d87c7da4a5cf5d2be6bee4e21bc226f",
//      created: "2013-04-30T17:46:10.843673+03:00",
//      container: "8305672a76cc5e3d168f97221106ced35a76ec7ddbb03209b0f0d96bf74f6ef7",
//      container_config: {
//          Hostname: "host-test",
//          User: "",
//          Memory: 0,
//          MemorySwap: 0,
//          AttachStdin: false,
//          AttachStdout: false,
//          AttachStderr: false,
//          PortSpecs: null,
//          Tty: false,
//          OpenStdin: false,
//          StdinOnce: false,
//          Env: null,
//          Cmd: [
//            "/bin/bash",
//            "-c",
//            "apt-get -q -yy -f install libevent-dev"
//          ],
//         Dns: null,
//         Image: "imagename/blah",
//         Volumes: { },
//         VolumesFrom: ""
//     },
//     docker_version: "0.1.7"
// }
// Example Response:
//    HTTP/1.1 200
//    Vary: Accept
//    Content-Type: application/json
//    X-Docker-Registry-Version: 0.6.0
//    ""
// Status Codes: 
//    200 – OK
//    401 – Requires authorization
func (this *ImageController) PUTJSON() {

}

// http://docs.docker.io/en/latest/reference/api/registry_api/#ancestry
// GET /v1/images/(image_id)/ancestry
// 读取 image 的所有上级 image_id ，读取 /registry/images/(image_id)/ancestry 文件并转换为 JSON 格式。
// Get ancestry for an image given an image_id
// Example Request:
//    GET /v1/images/088b4505aa3adc3d35e79c031fa126b403200f02f51920fbd9b7c503e87c7a2c/ancestry HTTP/1.1
//    Host: registry-1.docker.io
//    Accept: application/json
//    Content-Type: application/json
//    Cookie: (Cookie provided by the Registry)
// Example Response:
//    HTTP/1.1 200
//    Vary: Accept
//    Content-Type: application/json
//    X-Docker-Registry-Version: 0.6.0
//    ["088b4502f51920fbd9b7c503e87c7a2c05aa3adc3d35e79c031fa126b403200f",
//     "aeee63968d87c7da4a5cf5d2be6bee4e21bc226fd62273d180a49c96c62e4543",
//     "bfa4c5326bc764280b0863b46a4b20d940bc1897ef9c1dfec060604bdc383280",
//     "6ab5893c6927c15a15665191f2c6cf751f5056d8b95ceee32e43c5e8a3648544"]
// Status Codes: 
//    200 – OK
//    401 – Requires authorization
//    404 – Image not found
func (this *ImageController) GETAncestry() {

}

// Undocumented API
// 根据官方 docker-registry 的 put_image_checksum 方法分析：
// 检查 HTTP HEADER 的 Docker Client 版本：
//     如果是 0.10 版本从 X-Docker-Checksum-Payload 读取 checksum 值
//     如果是 0.10 以前的版本从 X-Docker-Checksum 读取 checksum 值
// 在上传 image 的 layer 文件时，checksum 被保存到 session 中，在存储之前检查是否和 session 中的值相同。
// checksum 的值存储在 /v1/images/(image_id)/checksum
// Status Codes: 
//    200 – OK
//    400 – 没有找到 checksum 或其它检查 checksum 合法性的时候出现错误
//    404 – Image not found
func (this *ImageController) PUTChecksum() {

}

// Undocumented API
// 根据官方 docker-registry 的 get_image_files 方法分析：
func (this *ImageController) GETFiles() {

}

// Undocumented API 
func (this *ImageController) GETPrivateLayer() {

}

func (this *ImageController) GETPrivateJSON() {

}


func (this *ImageController) GETPrivateFiles() {

}



func (this *ImageController) GETDiff() {
	
}
