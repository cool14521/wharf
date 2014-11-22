package models

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
)

//邮件信息定义
type Message struct {
	To          string
	From        string
	Subject     string
	Body        string
	ContentType string
	Prefix      string //邮件的前缀名称，用来匹配prefix的模板
	Host        string //mailServer的标志字段
	Status      string //记录发送状态 'I待发送 A发送成功 X发送失败'
	Count       int    //记录发送次数
}

type TemplateHtml struct {
	Prefix  string
	Content []byte
}

//邮件服务器定义
type MailServer struct {
	Host     string
	Port     int
	User     string
	Password string
}

func (tpl *TemplateHtml) Add(prefix, filePath string) error {
	//进行模板验证
	if len(strings.TrimSpace(prefix)) == 0 {
		log.Println("邮件模板前缀名为空值")
		return errors.New("邮件模板前缀名不能为空")
	} else if err, isFile := IsFileExist(filePath); err != nil || !isFile {
		log.Println("导入模板的路径非法,请检验文件是否存在")
		return err
	}
	//验证完成，进行存库操作
	bytes, _ := ioutil.ReadFile(filePath)

	tpl.Prefix = prefix
	tpl.Content = bytes

	LedisDB.HSet([]byte("template"), []byte(tpl.Prefix), tpl.Content)
	return nil
}

func (tpl *TemplateHtml) Query(prefixArr ...string) []*TemplateHtml {
	tmpls := make([]*TemplateHtml, 0)
	if len(prefixArr) == 0 {
		prefixes, _ := LedisDB.HKeys([]byte("template"))
		for _, prefix := range prefixes {
			tmpl := new(TemplateHtml)
			content, _ := LedisDB.HGet([]byte("template"), []byte(string(prefix)))
			tmpl.Prefix = string(prefix)
			tmpl.Content = content
			tmpls = append(tmpls, tmpl)
		}
		return tmpls
	} else if len(prefixArr) == 1 {
		tmpl := new(TemplateHtml)
		content, _ := LedisDB.HGet([]byte("template"), []byte(prefixArr[0]))
		tmpl.Prefix = prefixArr[0]
		tmpl.Content = content
		tmpls = append(tmpls, tmpl)
		return tmpls
	}
	for _, prefix := range prefixArr {
		tmpl := new(TemplateHtml)
		content, _ := LedisDB.HGet([]byte("template"), []byte(prefix))
		tmpl.Prefix = prefix
		tmpl.Content = content
		tmpls = append(tmpls, tmpl)
	}
	return tmpls
}

func (tpl *TemplateHtml) Delete(prefixArr ...string) error {
	if len(prefixArr) == 0 {
		log.Println("查询模板的前缀名为空值")
		return errors.New("查询模板的前缀值不能为空")
	}
	for _, prefix := range prefixArr {
		LedisDB.HDel([]byte("template"), []byte(prefix))
	}
	return nil
}

func (mailServer *MailServer) Add(host string, port int, user string, password string) error {
	//完成字符验证 长度验证，正则验证
	if len(strings.TrimSpace(host)) == 0 {
		log.Println("邮件服务器的地址为空值")
		return errors.New("邮件服务器的地址不能为空")
	} else if len(strings.TrimSpace(user)) == 0 {
		log.Println("邮件服务器用户名为空值")
		return errors.New("邮件服务器的用户名不能为空")
	} else if len(strings.TrimSpace(password)) == 0 {
		log.Println("邮件服务器的密码为空值")
		return errors.New("邮件服务器的秘法不能为空")
	} else if port == 0 {
		log.Println("邮件服务器端口为空值")
		return errors.New("邮件服务器端口不能为空")
	}
	mailServer.Host = host
	mailServer.Port = port
	mailServer.User = user
	mailServer.Password = password
	//测试邮件服务器是否联通,不联通返回错误，添加邮件服务器失败
	err := mailServer.DialTest()
	if err != nil {
		return err
	}
	//服务器存储
	mailServer4json, _ := json.Marshal(mailServer)
	LedisDB.HSet([]byte("mailServer"), []byte(mailServer.Host), mailServer4json)
	return nil
}

func (mailServer *MailServer) Query(hostArr ...string) []*MailServer {
	mailServers := make([]*MailServer, 0)
	if len(hostArr) == 0 {
		hosts, _ := LedisDB.HKeys([]byte("mailServer"))
		for _, host := range hosts {
			mailServer4json, _ := LedisDB.HGet([]byte("mailServer"), host)
			var mailServer_new MailServer
			json.Unmarshal(mailServer4json, &mailServer_new)
			mailServers = append(mailServers, &mailServer_new)
		}
		return mailServers
	} else if len(hostArr) == 1 {
		mailServer4json, _ := LedisDB.HGet([]byte("mailServer"), []byte(hostArr[0]))
		var mailServer_new MailServer
		json.Unmarshal(mailServer4json, &mailServer_new)
		mailServers = append(mailServers, &mailServer_new)
		return mailServers
	}
	for _, host := range hostArr {
		mailServer4json, _ := LedisDB.HGet([]byte("mailServer"), []byte(host))
		var mailServer_new MailServer
		json.Unmarshal(mailServer4json, &mailServer_new)
		mailServers = append(mailServers, &mailServer_new)
	}
	return mailServers
}

func (mailServer *MailServer) Delete(hostArr ...string) error {
	if len(hostArr) == 0 {
		log.Println("邮件服务器host为空值")
		return errors.New("邮件服务器不能为空")
	}
	for _, host := range hostArr {
		LedisDB.HDel([]byte("mailServer"), []byte(host))
	}
	return nil
}

func IsFileExist(filePath string) (error, bool) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return err, false
	} else if fi.IsDir() {
		return errors.New("传入参数应为文件而不是文件夹"), false
	}
	return nil, true
}

func (msg *Message) Add(to, from, subject, body, contentType, prefix, host string, model ...interface{}) error {
	//完成字符验证 长度验证，正则验证
	if len(strings.TrimSpace(to)) == 0 {
		log.Println("收件箱为空值")
		return errors.New("收件箱的地址不能为空")
	} else if len(strings.TrimSpace(from)) == 0 {
		log.Println("发件人的邮箱为空值")
		return errors.New("发件人的邮箱不能为空值")
	} else if len(strings.TrimSpace(body)) == 0 && strings.TrimSpace(contentType) != "html" {
		log.Println("发送内容为空值")
		return errors.New("发送内容不能为空")
	} else if len(strings.TrimSpace(contentType)) == 0 {
		log.Println("邮件类型空值")
		return errors.New("邮件类型不能为空")
	} else if len(strings.TrimSpace(prefix)) == 0 && strings.TrimSpace(contentType) == "text/html;charset=UTF-8" {
		log.Println("邮件依赖模板为空值")
		return errors.New("邮件依赖模板不能为空")
	} else if len(strings.TrimSpace(host)) == 0 {
		log.Println("邮件服务器为空值")
		return errors.New("邮件服务器不能为空值")
	} else if len(model) == 0 && strings.TrimSpace(contentType) == "text/html;charset=UTF-8" {
		log.Println("模板对象对象为空值")
		return errors.New("未传入模板依赖对象")
	} else if len(model) > 1 {
		log.Println("模板依赖对象大于1个")
		return errors.New("模板依赖对象只能有1个")
	}
	msg.To = to
	msg.From = from
	msg.Subject = subject
	msg.Prefix = prefix
	//判断类型，如果是html类型的邮件则对其进行渲染 对msg.Body渲染赋值
	if strings.TrimSpace(contentType) == "html" {
		msg.ContentType = "text/html; charset=UTF-8"
		err := msg.Render(model[0])
		if err != nil {
			return err
		}
	} else {
		msg.ContentType = contentType
		msg.Body = body
	}
	msg.Host = host
	msg.Status = "I"
	msg.Count = 0
	//存储的结构为prefix md5(to) msg4json
	to_md5 := EncodeMd5(msg.To)
	msg4json, _ := json.Marshal(msg)
	LedisDB.HSet([]byte(msg.Prefix), []byte(to_md5), msg4json)
	return nil
}

func (msg *Message) Query(prefixArr ...string) []*Message {
	msgs := make([]*Message, 0)
	if len(prefixArr) > 0 {
		for _, prefix := range prefixArr {
			toes, _ := LedisDB.HKeys([]byte(prefix))
			for _, to := range toes {
				var new_msg Message
				msg4json, _ := LedisDB.HGet([]byte(prefix), to)
				json.Unmarshal(msg4json, &new_msg)
				msgs = append(msgs, &new_msg)
			}
		}
		return msgs
	}
	return nil
}

func (msg *Message) Update() {
	msg4json, _ := json.Marshal(msg)
	to_md5 := EncodeMd5(msg.To)
	LedisDB.HSet([]byte(msg.Prefix), []byte(to_md5), msg4json)
}

func (msg *Message) Render(model interface{}) error {
	content, _ := LedisDB.HGet([]byte("template"), []byte(msg.Prefix))
	tmpl, err := template.New("tmpl").Parse(string(content))
	if err != nil {
		log.Println("渲染模板失败")
		return err
	}
	var body bytes.Buffer
	err = tmpl.Execute(&body, model)
	if err != nil {
		log.Println("传入结构体与", msg.Prefix, "模板不匹配，渲染失败")
		return err
	}
	msg.Body = body.String()
	return nil
}

func EncodeMd5(email string) string {
	h := md5.New()
	h.Write([]byte(email))
	return hex.EncodeToString(h.Sum(nil))
}

func (mailServer *MailServer) DialTest() error {
	_, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", mailServer.Host, mailServer.Port), nil)
	if err != nil {
		log.Println("Dialing Error:", err)
		return err
	}
	return nil
}
