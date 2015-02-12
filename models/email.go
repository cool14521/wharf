package models

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/textproto"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/dockercn/wharf/utils"
)

//邮件信息定义
type Message struct {
	To          string
	Cc          []string
	Bcc         []string
	From        string
	Subject     string
	Body        string
	Type        string
	Headers     textproto.MIMEHeader
	Attachments []*Attachment //附件
	Prefix      string        //邮件的前缀名称，用来匹配prefix的模板
	Host        string        //mailServer的标志字段
	Status      string        //记录发送状态 'I待发送 A发送成功 X发送失败'
	Count       int           //记录发送次数
}

//邮件模板定义
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

//附件结构体
type Attachment struct {
	Filename string
	Header   textproto.MIMEHeader
	Content  []byte
}

//给信息添加附件
func (msg *Message) AttachFile(args ...string) (a *Attachment, err error) {
	if len(args) < 1 && len(args) > 2 {
		err = errors.New("Must specify a file name and number of parameters can not exceed at least two")
		return
	}
	filename := args[0]
	id := ""
	if len(args) > 1 {
		id = args[1]
	}
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	ct := mime.TypeByExtension(filepath.Ext(filename))
	basename := path.Base(filename)
	return msg.Attach(f, basename, ct, id)
}

func (msg *Message) Attach(r io.Reader, filename string, args ...string) (a *Attachment, err error) {
	if len(args) < 1 && len(args) > 2 {
		err = errors.New("Must specify the file type and number of parameters can not exceed at least two")
		return
	}
	c := args[0] //Content-Type
	id := ""
	if len(args) > 1 {
		id = args[1] //Content-ID
	}
	var buffer bytes.Buffer
	if _, err = io.Copy(&buffer, r); err != nil {
		return
	}
	at := &Attachment{
		Filename: filename,
		Header:   textproto.MIMEHeader{},
		Content:  buffer.Bytes(),
	}
	// Get the Content-Type to be used in the MIMEHeader
	if c != "" {
		at.Header.Set("Content-Type", c)
	} else {
		// If the Content-Type is blank, set the Content-Type to "application/octet-stream"
		at.Header.Set("Content-Type", "application/octet-stream")
	}
	if id != "" {
		at.Header.Set("Content-Disposition", fmt.Sprintf("inline;\r\n filename=\"%s\"", filename))
		at.Header.Set("Content-ID", fmt.Sprintf("<%s>", id))
	} else {
		at.Header.Set("Content-Disposition", fmt.Sprintf("attachment;\r\n filename=\"%s\"", filename))
	}
	at.Header.Set("Content-Transfer-Encoding", "base64")
	msg.Attachments = append(msg.Attachments, at)
	return at, nil
}

func (tpl *TemplateHtml) Add(prefix, filePath string) error {
	//进行模板验证
	if len(strings.TrimSpace(prefix)) == 0 {
		log.Println("邮件模板前缀名为空值")
		return errors.New("邮件模板前缀名不能为空")
	} else if err, isFile := utils.IsFileExists(filePath); err != nil || !isFile {
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

func (msg *Message) Add(to, from, subject, body, contentType, prefix, host string, cc []string, bcc []string, model ...interface{}) error {
	//完成字符验证 长度验证，正则验证
	if m, _ := regexp.MatchString("[a-z0-9A-Z]@([a-z0-9A-Z]+(-[a-z0-9A-Z]+)?\\.)+[a-zA-Z]{2,}$", to); len(strings.TrimSpace(to)) == 0 || !m {
		log.Println("收件箱不符合规范")
		return errors.New("收件箱的地址不符合规范")
	} else if m, _ := regexp.MatchString("[a-z0-9A-Z]@([a-z0-9A-Z]+(-[a-z0-9A-Z]+)?\\.)+[a-zA-Z]{2,}$", from); len(strings.TrimSpace(from)) == 0 || !m {
		log.Println("发件人的邮箱不符合规范")
		return errors.New("发件人的邮箱不符合规范")
	} else if len(strings.TrimSpace(body)) == 0 && strings.TrimSpace(contentType) != "html" {
		log.Println("发送内容为空值")
		return errors.New("发送内容不能为空")
	} else if len(strings.TrimSpace(contentType)) == 0 {
		log.Println("邮件类型空值")
		return errors.New("邮件类型不能为空")
	} else if len(strings.TrimSpace(prefix)) == 0 && strings.TrimSpace(contentType) == "html" {
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
	msg.Cc = cc
	msg.Bcc = bcc
	//判断类型，如果是html类型的邮件则对其进行渲染 对msg.Body渲染赋值
	if contentType := strings.TrimSpace(contentType); contentType == "html" {
		msg.Type = contentType
		err := msg.Render(model[0])
		if err != nil {
			return err
		}
	} else if contentType == "text" {
		msg.Type = contentType
		msg.Body = body
	} else {
		log.Println("邮件类型为text或者html")
		return errors.New("邮件类型指定错误")
	}
	msg.Host = host
	msg.Status = "I"
	msg.Count = 0
	//存储的结构为prefix md5(to) msg4json
	to_md5 := utils.EncodeEmail(msg.To)
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
	to_md5 := utils.EncodeEmail(msg.To)
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

func (mailServer *MailServer) DialTest() error {
	_, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", mailServer.Host, mailServer.Port), nil)
	if err != nil {
		log.Println("Dialing Error:", err)
		return err
	}
	return nil
}
