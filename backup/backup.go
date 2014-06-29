package backup

import (
  "errors"

  "github.com/astaxie/beego"
  "github.com/dockboard/docker-registry/models"
)

var (
  //上传队列
  UploadChan chan string = make(chan string, 1024*32)
  //上传完成队列
  ResultChan chan string = make(chan string, 1024)
)

var Backup bool

func UpdateBackupURL(filename, url string) error {

  image := &models.Image{ImageId: filename}

  has, err := models.Engine.Get(image)
  if err != nil {
    beego.Trace("[Error] Find image data in database encounter error, image id : " + filename)

    return err
  }

  if has == true {
    image.Backup = url
    _, err = models.Engine.Id(image.Id).Cols("Backup").Update(image)
    if err != nil {
      beego.Trace("[Error] Update the backup url error, image id : " + filename)

      return err
    }
  } else {
    return errors.New("Could not find image data in database,  image id : " + filename)
  }

  return nil

}
