package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/astaxie/beego"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"

	"github.com/dockercn/wharf/utils"
)

const (
	GLOBAL_USER_INDEX         = "GLOBAL_USER_INDEX"
	GLOBAL_REPOSITORY_INDEX   = "GLOBAL_REPOSITORY_INDEX"
	GLOBAL_ORGANIZATION_INDEX = "GLOBAL_ORGANIZATION_INDEX"
	GLOBAL_TEAM_INDEX         = "GLOBAL_TEAM_INDEX"
	GLOBAL_IMAGE_INDEX        = "GLOBAL_IMAGE_INDEX"
	GLOBAL_TARSUM_INDEX       = "GLOBAL_TARSUM_INDEX"
	GLOBAL_TAG_INDEX          = "GLOBAL_TAG_INDEX"
	GLOBAL_COMPOSE_INDEX      = "GLOBAL_COMPOSE_INDEX"
	GLOBAL_ADMIN_INDEX        = "GLOBAL_ADMIN_INDEX"
	GLOBAL_PRIVILEGE_INDEX    = "GLOBAL_PRIVILEGE_INDEX"
	GLOBAL_LOG_INDEX          = "GLOBAL_LOG_INDEX"
)

var (
	ledisOnce sync.Once
	nowLedis  *ledis.Ledis
	LedisDB   *ledis.DB
)

func InitDb() {
	initLedisFunc := func() {
		cfg := new(config.Config)
		cfg.DataDir = beego.AppConfig.String("ledisdb::DataDir")
		var err error
		nowLedis, err = ledis.Open(cfg)
		if err != nil {
			println(err.Error())
			panic(err)
		}
	}

	ledisOnce.Do(initLedisFunc)

	db, _ := beego.AppConfig.Int("ledisdb::DB")

	LedisDB, _ = nowLedis.Select(db)
}

func GetUUID(ObjectType, Object string) (UUID []byte, err error) {

	index := ""

	switch strings.TrimSpace(ObjectType) {

	case "user":
		index = GLOBAL_USER_INDEX
	case "repository":
		index = GLOBAL_REPOSITORY_INDEX
	case "organization":
		index = GLOBAL_ORGANIZATION_INDEX
	case "team":
		index = GLOBAL_TEAM_INDEX
	case "image":
		index = GLOBAL_IMAGE_INDEX
	case "tarsum":
		index = GLOBAL_TARSUM_INDEX
	case "tag":
		index = GLOBAL_TAG_INDEX
	case "compose":
		index = GLOBAL_COMPOSE_INDEX
	case "admin":
		index = GLOBAL_ADMIN_INDEX
	case "log":
		index = GLOBAL_LOG_INDEX
	default:

	}

	if UUID, err = LedisDB.HGet([]byte(index), []byte(Object)); err != nil {
		return nil, err
	} else {
		return UUID, nil
	}

}

func Save(obj interface{}, key []byte) (err error) {
	s := reflect.TypeOf(obj).Elem()

	for i := 0; i < s.NumField(); i++ {

		value := reflect.ValueOf(obj).Elem().Field(s.Field(i).Index[0])

		switch value.Kind() {

		case reflect.String:
			if _, err := LedisDB.HSet(key, []byte(s.Field(i).Name), []byte(value.String())); err != nil {
				return err
			}

		case reflect.Bool:
			if _, err := LedisDB.HSet(key, []byte(s.Field(i).Name), utils.BoolToBytes(value.Bool())); err != nil {
				return err
			}

		case reflect.Int64:
			if _, err := LedisDB.HSet(key, []byte(s.Field(i).Name), utils.Int64ToBytes(value.Int())); err != nil {
				return err
			}
		case reflect.Slice:
			if "[]string" == value.Type().String() && !value.IsNil() {
				strJson := "["

				for i := 0; i < value.Len(); i++ {
					nowUUID := value.Index(i).String()
					if i != 0 {
						strJson += fmt.Sprintf(`,"%s"`, nowUUID)
					} else {
						strJson += fmt.Sprintf(`"%s"`, nowUUID)
					}
				}

				strJson += "]"

				if _, err := LedisDB.HSet(key, []byte(s.Field(i).Name), []byte(strJson)); err != nil {
					return err
				}
			}

		default:
		}

	}
	return nil
}

func Get(obj interface{}, UUID []byte) (err error) {

	nowTypeElem := reflect.ValueOf(obj).Elem()
	types := nowTypeElem.Type()

	for i := 0; i < nowTypeElem.NumField(); i++ {

		nowField := nowTypeElem.Field(i)
		nowFieldName := types.Field(i).Name

		switch nowField.Kind() {

		case reflect.String:
			nowValue, err := LedisDB.HGet(UUID, []byte(nowFieldName))
			nowField.SetString(string(nowValue))
			if err != nil {
				return err
			}

		case reflect.Bool:
			nowValue, err := LedisDB.HGet(UUID, []byte(nowFieldName))
			nowField.SetBool(utils.BytesToBool(nowValue))
			if err != nil {
				return err
			}

		case reflect.Int64:
			nowValue, err := LedisDB.HGet(UUID, []byte(nowFieldName))
			nowField.SetInt(utils.BytesToInt64(nowValue))
			if err != nil {
				return err
			}

		case reflect.Slice:
			if "[]string" == nowField.Type().String() {
				nowValue, err := LedisDB.HGet(UUID, []byte(nowFieldName))

				var stringSlice []string
				err = json.Unmarshal(nowValue, &stringSlice)

				if err != nil && (len(nowValue) > 0) {
					return err
				}

				sliceValue := reflect.ValueOf(stringSlice)
				nowField.Set(sliceValue)
			}

		default:
		}
	}

	return nil

}
