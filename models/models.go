package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/astaxie/beego"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"

	"github.com/dockercn/wharf/utils"
)

var (
	ledisOnce sync.Once
	nowLedis  *ledis.Ledis
	LedisDB   *ledis.DB
)

// InitDb initializes the database.
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

func Save(obj interface{}, key []byte) (err error) {
	s := reflect.TypeOf(obj).Elem()

	//循环处理 Struct 的每一个 Field
	for i := 0; i < s.NumField(); i++ {
		//获取 Field 的 Value
		value := reflect.ValueOf(obj).Elem().Field(s.Field(i).Index[0])

		//判断 Field 不为空
		//	if utils.IsEmptyValue(value) == false {
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
					//		fmt.Println("Slice 保存的UUID:::", nowUUID)
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
				sliceValue := reflect.ValueOf(stringSlice) // 这里将slice转成reflect.Value类型
				nowField.Set(sliceValue)
			}
		default:
		}
	}
	return nil
}
