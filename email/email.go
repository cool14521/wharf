package email

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"time"

	"github.com/dockercn/docker-bucket/models"
)

//默认5分钟发一次
func MailService() {
	log.Println("..........邮件服务已经正常启动..........")
	go func() {
		for {
			//加载模板列表去到prefix的集合，再去prefix对应的message
			tmpl := new(models.TemplateHtml)
			tmpls := tmpl.Query()
			for i, _ := range tmpls {
				msg := new(models.Message)
				msgs := msg.Query(tmpls[i].Prefix)
				for j, _ := range msgs {
					mailServer := new(model.MailServer)
					server := mailServer.Query(msgs[j].Host)
					isSend, err := Send(server[0], msgs[j])
					if isSend {
						msgs[j].Update()
					}
				}
			}
			time.Sleep(5 * time.Minute)
		}
	}()
}

//调用发送邮件服务 返回值bool表示是否发送  error表示如果发送是成功还是失败
func Send(mailServer *models.MailServer, msg *models.Message) (bool, error) {
	//判断邮件状态  I待发送 A 已经发送 X 发送失败    发送失败超过10次，停止发送
	if msg.Status == "A" {
		return false, nil
	} else if msg.Status == "X" && msg.Count > 10 {
		return false, nil
	}
	header := make(map[string]string)
	header["From"] = msg.From
	header["To"] = msg.To
	header["Subject"] = msg.Subject
	header["Content-Type"] = msg.Type
	content := ""
	for k, v := range header {
		content += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	content += "\r\n" + msg.Body
	auth := smtp.PlainAuth("", mailServer.User, mailServer.Password, mailServer.Host)
	err := sendMailUsingTLS(fmt.Sprintf("%s:%d", mailServer.Host, mailServer.Port), auth, msg.From, []string{msg.To}, []byte(content))
	//判断发送是否成功
	msg.Count = msg.Count + 1
	if err != nil {
		msg.Status = "X"
		return true, err
	}
	msg.Status = "A"
	return true, nil
}

func sendMailUsingTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) (err error) {
	//create smtp client
	c, err := dial(addr)
	if err != nil {
		log.Println("Create smpt client error:", err)
		return err
	}
	defer c.Close()
	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				log.Println("Error during AUTH", err)
				return err
			}
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

//return a smtp client
func dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		log.Println("Dialing Error:", err)
		return nil, err
	}
	//分解主机端口字符串
	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}
