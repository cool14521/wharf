package email

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"github.com/dockercn/docker-bucket/models"
)

const (
	maxLineLength = 76
)

var stop bool

//默认5分钟发一次
func StartService() {
	checkService()
	go func() {
		for {
			var msg_count, tmpl_count, msg_success_count, msg_failed_count int //分别统计信息，模板,发送成功条数,发送失败条数

			//加载模板列表去到prefix的集合，再去prefix对应的message
			tmpl := new(models.TemplateHtml)
			tmpls := tmpl.Query()
			tmpl_count = len(tmpls)

			for i, _ := range tmpls {
				msg := new(models.Message)
				msgs := msg.Query(tmpls[i].Prefix)
				for j, _ := range msgs {
					msg_count++
					fmt.Printf("%#v\n", msgs[j])
					mailServer := new(models.MailServer)
					server := mailServer.Query(msgs[j].Host)
					isSend, _ := Send(server[0], msgs[j])
					if isSend {
						if msg.Status == "A" {
							msg_success_count++
						} else if msg.Status == "X" {
							msg_failed_count++
						}
						msgs[j].Update()
					}
				}
			}
			if stop {
				return
			}
			log.Printf("[email]数据库中共有%d个模板，数据库中一共有邮件%d封，本次发送成功%d条，发送失败%d条\n", tmpl_count, msg_count, msg_success_count, msg_failed_count)
			time.Sleep(5 * time.Minute)
		}
	}()
}

func StopService() {
	stop = true
}

func checkService() {
	mailServer := new(models.MailServer)
	servers := mailServer.Query()
	if len(servers) == 0 {
		log.Println("[email]数据库中未设置任何smtp服务器,无法发送任何邮件")
		return
	}
	log.Println("[email]邮件服务器设置检查完成，邮件服务正常启动")
}

//调用发送邮件服务 返回值bool表示是否发送  error表示如果发送是成功还是失败
func Send(mailServer *models.MailServer, msg *models.Message) (bool, error) {
	//判断邮件状态  I待发送 A 已经发送 X 发送失败    发送失败超过10次，停止发送
	if msg.Status == "A" {
		return false, nil
	} else if msg.Status == "X" && msg.Count > 10 {
		return false, nil
	}
	auth := smtp.PlainAuth("", mailServer.User, mailServer.Password, mailServer.Host)
	to := make([]string, 0, len(msg.Cc)+len(msg.Bcc)+1)
	to = append(append(append(to, msg.To), msg.Cc...), msg.Bcc...)
	// Check to make sure there is at least one recipient and one "From" address
	if msg.From == "" || len(to) == 0 {
		return false, errors.New("Must specify at least one From address and one To address")
	}
	from, err := mail.ParseAddress(msg.From)
	if err != nil {
		return false, err
	}
	raw, err := msg2bytes(msg)
	if err != nil {
		return false, err
	}
	//err = SendMail(mailServer.Host+":"+strconv.Itoa(mailServer.Port), auth, from.Address, to, raw)
	address := fmt.Sprintf("%s:%d", mailServer.Host, 465)
	err = sendMailUsingTLS(address, auth, from.Address, to, raw)
	//判断发送是否成功
	msg.Count = msg.Count + 1
	if err != nil {
		msg.Status = "X"
		return true, err
	}
	msg.Status = "A"
	return true, nil
}

func msg2bytes(e *models.Message) ([]byte, error) {
	e.Headers = textproto.MIMEHeader{}
	buff := &bytes.Buffer{}
	w := multipart.NewWriter(buff)
	// Set the appropriate headers (overwriting any conflicts)
	// Leave out Bcc (only included in envelope headers)
	e.Headers.Set("To", e.To)
	if e.Cc != nil {
		e.Headers.Set("Cc", strings.Join(e.Cc, ","))
	}
	e.Headers.Set("From", e.From)
	e.Headers.Set("Subject", e.Subject)
	e.Headers.Set("MIME-Version", "1.0")
	e.Headers.Set("Content-Type", fmt.Sprintf("multipart/mixed;\r\n boundary=%s\r\n", w.Boundary()))
	// Write the envelope headers (including any custom headers)
	if err := headerToBytes(buff, e.Headers); err != nil {
		return nil, fmt.Errorf("Failed to render message headers: %s", err)
	}
	// Start the multipart/mixed part
	fmt.Fprintf(buff, "--%s\r\n", w.Boundary())
	header := textproto.MIMEHeader{}
	// Check to see if there is a Text or HTML field
	if e.Type != "" {
		subWriter := multipart.NewWriter(buff)
		// Create the multipart alternative part
		header.Set("Content-Type", fmt.Sprintf("multipart/alternative;\r\n boundary=%s\r\n", subWriter.Boundary()))
		// Write the header
		if err := headerToBytes(buff, header); err != nil {
			return nil, fmt.Errorf("Failed to render multipart message headers: %s", err)
		}
		// Create the body sections
		if e.Type == "text" {
			header.Set("Content-Type", fmt.Sprintf("text/plain; charset=UTF-8"))
			header.Set("Content-Transfer-Encoding", "quoted-printable")
			if _, err := subWriter.CreatePart(header); err != nil {
				return nil, err
			}
			// Write the text
			if err := quotePrintEncode(buff, e.Body); err != nil {
				return nil, err
			}
		}
		if e.Type == "html" {
			header.Set("Content-Type", fmt.Sprintf("text/html; charset=UTF-8"))
			header.Set("Content-Transfer-Encoding", "quoted-printable")
			if _, err := subWriter.CreatePart(header); err != nil {
				return nil, err
			}
			// Write the text
			if err := quotePrintEncode(buff, e.Body); err != nil {
				return nil, err
			}
		}
		if err := subWriter.Close(); err != nil {
			return nil, err
		}
	}
	// Create attachment part, if necessary
	for _, a := range e.Attachments {
		ap, err := w.CreatePart(a.Header)
		if err != nil {
			return nil, err
		}
		// Write the base64Wrapped content to the part
		base64Wrap(ap, a.Content)
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

// headerToBytes enumerates the key and values in the header, and writes the results to the IO Writer
func headerToBytes(w io.Writer, t textproto.MIMEHeader) error {
	for k, v := range t {
		// Write the header key
		_, err := fmt.Fprintf(w, "%s:", k)
		if err != nil {
			return err
		}
		// Write each value in the header
		for _, c := range v {
			_, err := fmt.Fprintf(w, " %s\r\n", c)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func quotePrintEncode(w io.Writer, s string) error {
	var buf [3]byte
	mc := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		// We're assuming Unix style text formats as input (LF line break), and
		// quoted-printble uses CRLF line breaks. (Literal CRs will become
		// "=0D", but probably shouldn't be there to begin with!)
		if c == '\n' {
			io.WriteString(w, "\r\n")
			mc = 0
			continue
		}
		var nextOut []byte
		if isPrintable(c) {
			nextOut = append(buf[:0], c)
		} else {
			nextOut = buf[:]
			qpEscape(nextOut, c)
		}
		// Add a soft line break if the next (encoded) byte would push this line
		// to or past the limit.
		if mc+len(nextOut) >= maxLineLength {
			if _, err := io.WriteString(w, "=\r\n"); err != nil {
				return err
			}
			mc = 0
		}
		if _, err := w.Write(nextOut); err != nil {
			return err
		}
		mc += len(nextOut)
	}
	// No trailing end-of-linejQuery21108770898473449051_1419300850373 Soft line break, then. TODO: is this sane?
	if mc > 0 {
		io.WriteString(w, "=\r\n")
	}
	return nil
}

// base64Wrap encodes the attachment content, and wraps it according to RFC 2045 standards (every 76 chars)
// The output is then written to the specified io.Writer
func base64Wrap(w io.Writer, b []byte) {
	// 57 raw bytes per 76-byte base64 line.
	const maxRaw = 57
	// Buffer for each line, including trailing CRLF.
	var buffer [maxLineLength + len("\r\n")]byte
	copy(buffer[maxLineLength:], "\r\n")
	// Process raw chunks until there's no longer enough to fill a line.
	for len(b) >= maxRaw {
		base64.StdEncoding.Encode(buffer[:], b[:maxRaw])
		w.Write(buffer[:])
		b = b[maxRaw:]
	}
	// Handle the last chunk of bytes.
	if len(b) > 0 {
		out := buffer[:base64.StdEncoding.EncodedLen(len(b))]
		base64.StdEncoding.Encode(out, b)
		out = append(out, "\r\n"...)
		w.Write(out)
	}
}

// qpEscape is a helper function for quotePrintEncode which escapes a
// non-printable byte. Expects len(dest) == 3.
func qpEscape(dest []byte, c byte) {
	const nums = "0123456789ABCDEF"
	dest[0] = '='
	dest[1] = nums[(c&0xf0)>>4]
	dest[2] = nums[(c & 0xf)]
}

// isPrintable returns true if the rune given is "printable" according to RFC 2045, false otherwise
func isPrintable(c byte) bool {
	return (c >= '!' && c <= '<') || (c >= '>' && c <= '~') || (c == ' ' || c == '\n' || c == '\t')
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
