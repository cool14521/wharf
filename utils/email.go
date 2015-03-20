package utils

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
)

type Email struct {
	Host        string
	Port        int
	User        string
	Password    string
	From        string
	To          string
	Subject     string
	ContentType string
	Body        string
}

func SendEmail(email *Email) error {
	header := make(map[string]string)
	header["From"] = email.From
	header["To"] = email.To
	header["Subject"] = email.Subject
	header["Content-Type"] = email.ContentType

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + email.Body

	auth := smtp.PlainAuth("", email.User, email.Password, email.Host)

	err := sendMailUsingTLS(fmt.Sprintf("%s:%d", email.Host, email.Port), auth, email.From, []string{email.To}, []byte(message))

	if err != nil {
		return err
	}

	return nil
}

func dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		log.Println("Dialing Error:", err)
		return nil, err
	}

	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

func sendMailUsingTLS(addr string, auth smtp.Auth, from string,
	to []string, msg []byte) (err error) {

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
