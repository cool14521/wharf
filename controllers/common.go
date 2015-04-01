package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/astaxie/beego"

	"github.com/containerops/wharf/models"
)

func manifestsConvertV1(data []byte) error {
	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}

	tag := manifest["tag"]
	namespace, repository := strings.Split(manifest["name"].(string), "/")[0], strings.Split(manifest["name"].(string), "/")[1]

	for k := len(manifest["history"].([]interface{})) - 1; k >= 0; k-- {
		v := manifest["history"].([]interface{})[k]
		compatibility := v.(map[string]interface{})["v1Compatibility"].(string)

		var image map[string]interface{}
		if err := json.Unmarshal([]byte(compatibility), &image); err != nil {
			return err
		}

		i := map[string]string{}
		r := new(models.Repository)

		if k == 0 {
			i["Tag"] = tag.(string)
		}
		i["id"] = image["id"].(string)

		//Put V1 JSON
		if err := r.PutJSONFromManifests(i, namespace, repository); err != nil {
			return err
		}

		if k == 0 {
			//Put V1 Tag
			if err := r.PutTagFromManifests(image["id"].(string), namespace, repository, tag.(string), string(data)); err != nil {
				return err
			}
		}

		img := new(models.Image)

		tarsum := manifest["fsLayers"].([]interface{})[k].(map[string]interface{})["blobSum"].(string)
		sha256 := strings.Split(tarsum, ":")[1]

		//Put Image Json
		if err := img.PutJSON(image["id"].(string), v.(map[string]interface{})["v1Compatibility"].(string), models.APIVERSION_V2); err != nil {
			return err
		}

		//Put Image Layer
		basePath := beego.AppConfig.String("docker::BasePath")
		layerfile := fmt.Sprintf("%v/uuid/%v/layer", basePath, sha256)

		if err := img.PutLayer(image["id"].(string), layerfile, true, int64(image["Size"].(float64))); err != nil {
			return err
		}

		//Put Checksum
		if err := img.PutChecksum(image["id"].(string), sha256, true, ""); err != nil {
			return err
		}

		//Put Ancestry
		if err := img.PutAncestry(image["id"].(string)); err != nil {
			return err
		}
	}

	return nil
}
