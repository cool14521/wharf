package backup

import (
  "fmt"

  "github.com/astaxie/beego"
  "github.com/qiniu/api/io"
  "github.com/qiniu/api/rs"
)

func QiniuBackup(uploadChan, resultChan chan string) {

  for {
    //Image Layer 的文件名, 即 Image 的 ID
    filename := <-uploadChan
    filepath := fmt.Sprintf("%v/images/%v/layer", beego.AppConfig.String("docker::BasePath"), filename)

    var policy rs.PutPolicy
    policy.Scope = beego.AppConfig.String("qiniu::ImageBucket") + ":" + filename

    var fileExtra = &io.PutExtra{
      MimeType: "application/x-tar",
    }

    beego.Trace("[Backup Image ID] " + filename)

    //TODO: 判断文件是否存在

    ret := new(io.PutRet)
    token := policy.Token(nil)
    err := io.PutFile(nil, &ret, token, filename, filepath, fileExtra)

    if err != nil {
      //反复上传失败的文件
      uploadChan <- filename

      beego.Trace("[Backup Erro Message] " + err.Error())

    } else {

      beego.Trace("[Backup Successfully Message] " + filename)

      resultChan <- filename
    }
  }
}

func QiniuResult(resultChan chan string) {
  for {
    filename := <-resultChan

    beego.Trace("[Backup Message From Channel] " + filename)

    UpdateBackupURL(filename, fmt.Sprintf("http://%v.%v/%v", beego.AppConfig.String("qiniu::ImageBucket"), beego.AppConfig.String("qiniu::Domain"), filename))
  }
}
