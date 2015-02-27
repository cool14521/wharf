package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func Int64ToBytes(i int64) []byte {
	b_buf := new(bytes.Buffer)
	binary.Write(b_buf, binary.BigEndian, i)
	return b_buf.Bytes()
}

func BytesToInt64(b []byte) int64 {
	b_buf := bytes.NewBuffer(b)
	x, _ := binary.ReadVarint(b_buf)
	return x
}

func NowToBytes() []byte {
	return Int64ToBytes(time.Now().Unix())
}

func TimeToBytes(now time.Time) []byte {
	return Int64ToBytes(now.Unix())
}

func BoolToBytes(boolean bool) []byte {
	if boolean == true {
		return []byte("true")
	} else {
		return []byte("false")
	}
}

func BytesToBool(value []byte) bool {
	if boolean, _ := strconv.ParseBool(string(value)); boolean == true {
		return true
	} else if boolean == false {
		return false
	} else {
		return false
	}
}

func IsEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func ToString(args ...interface{}) string {
	result := ""

	for _, arg := range args {
		switch val := arg.(type) {
		case int:
			result += strconv.Itoa(val)
		case int64:
			result += strconv.FormatInt(val, 10)
		case string:
			result += val
		}
	}
	return result
}

func GeneralKey(key string) []byte {
	md5String := fmt.Sprintf("%v%v", key, time.Now())
	h := md5.New()
	h.Write([]byte(md5String))
	return []byte(hex.EncodeToString(h.Sum(nil)))
}

func EncodePassword(username string, password string) string {
	md5String := fmt.Sprintf("%s%s%s", username, password, "wharf")
	h := md5.New()
	h.Write([]byte(md5String))

	return hex.EncodeToString(h.Sum(nil))
}

//Encode the authorization string
func EncodeBasicAuth(username string, password string) string {
	auth := username + ":" + password
	msg := []byte(auth)
	authorization := make([]byte, base64.StdEncoding.EncodedLen(len(msg)))
	base64.StdEncoding.Encode(authorization, msg)
	return string(authorization)
}

// decode the authorization string
func DecodeBasicAuth(authorization string) (username string, password string, err error) {
	basic := strings.Split(strings.TrimSpace(authorization), " ")
	if len(basic) <= 1 {
		return "", "", err
	}

	decLen := base64.StdEncoding.DecodedLen(len(basic[1]))
	decoded := make([]byte, decLen)
	authByte := []byte(basic[1])
	n, err := base64.StdEncoding.Decode(decoded, authByte)

	if err != nil {
		return "", "", err
	}
	if n > decLen {
		return "", "", fmt.Errorf("Something went wrong decoding auth config")
	}

	arr := strings.SplitN(string(decoded), ":", 2)
	if len(arr) != 2 {
		return "", "", fmt.Errorf("Invalid auth configuration file")
	}

	username = arr[0]
	password = strings.Trim(arr[1], "\x00")

	return username, password, nil
}

func IsDirExists(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	} else {
		return fi.IsDir()
	}

	panic("not reached")
}

func IsFileExists(filePath string) (error, bool) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return err, false
	} else if fi.IsDir() {
		return errors.New("传入参数应为文件而不是文件夹"), false
	}
	return nil, true
}

func EncodeEmail(email string) string {
	h := md5.New()
	h.Write([]byte(email))
	return hex.EncodeToString(h.Sum(nil))
}
