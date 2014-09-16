package utils

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

func NowToBytes() []byte {
	return Int64ToBytes(time.Now().Unix())
}

func BoolToBytes(boolean bool) []byte {
	rst := make([]byte, 0)
	return strconv.AppendBool(rst, boolean)
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

func SendActiveEmail(code string, email string, host string, port int, user string, password string) error {
	return nil
}

func SendAddEmail(username string, passwd string, email string, host string, port int, user string, password string) error {
	return nil
}

func GeneralToken(key string) string {
	md5String := fmt.Sprintf("%v%v", key, string(time.Now().Unix()))
	h := md5.New()
	h.Write([]byte(md5String))
	return hex.EncodeToString(h.Sum(nil))
}

func EncodePassword(username string, password string) string {
	md5String := fmt.Sprintf("%v%v%v", username, password, "docker.cn")
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

func RemoveDuplicateString(s *[]string) {
	found := make(map[string]bool)
	j := 0

	for i, val := range *s {
		if _, ok := found[val]; !ok {
			found[val] = true
			(*s)[j] = (*s)[i]
			j++
		}
	}

	*s = (*s)[:j]
}
