package markdown

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/toolbox"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
)

var conn *ledis.DB

type Item struct {
	Key        string
	Title      string
	Desc       string
	Keywords   string
	UpdateTime string
	Content    string
	Tag        string
}

func initDB(path string) {
	cfg := new(config.Config)
	cfg.DataDir = path
	var err error
	nowLedis, err := ledis.Open(cfg)
	conn, err = nowLedis.Select(1)
	if err != nil {
		println(err)
	}
}

func InitTask() {
	//初始化数据库连接
	initDB(beego.AppConfig.String("ledisdb::DataDir"))

	//初始化任务并执行
	task := toolbox.NewTask("tk1", "0 0 */2 * * *", SyncData)
	//toolbox.AddTask("tk1", task)
	//toolbox.StartTask()
	task.Run()
	//defer toolbox.StopTask()
}

func SyncData() error {
	//每两个小时更新docs、documents目录
	cmd := exec.Command("git", "pull")
	cmd.Dir = beego.AppConfig.String("docker::DocsPath")
	_, err := cmd.Output()
	if err != nil {
		beego.Error("cmd Error=", err)
		return err
	}
	generateDict("A", beego.AppConfig.String("docker::DocsPath"))
	return nil
}

//typeArticle 文章类型
//A 普通文章  H 帮助文档
func generateDict(typeArticle, path string) {
	files, err := ioutil.ReadDir(path)
	//设定协程数量
	fileMap := make(map[string]string, len(files))

	if err != nil {
		beego.Error("errReadDir=", err)
	}

	for _, file := range files {
		if file.Name() == "Readme.md" || file.Name() == "rss.md" || strings.HasPrefix(file.Name(), ".git") {
			continue
		}
		cmd := exec.Command("git", "log", "-1", "--format=\"%ai\"", "--", file.Name())
		cmd.Dir = path
		output, err := cmd.Output()
		if err != nil {
			beego.Error("CommandErr=", err)
		}
		regex, _ := regexp.Compile(`(?m)^"(.*) .*?0800`)
		outputString := string(output)
		result := regex.FindStringSubmatch(outputString)
		var timeString string
		for _, v := range result {
			timeString = v
		}
		item := new(Item)
		item.UpdateTime = timeString
		fileMap[strings.Split(file.Name(), ".")[0]] = timeString
		go item.InsertLedis(typeArticle, path, file.Name())
	}
	DeteleArticle(typeArticle, fileMap)
}

func FindDetail(path, fileName string) (string, string, string, string, string) {
	f, err := os.Open(path + "/" + fileName)
	if err != nil {
		return "", "", "", "", ""
	}
	defer f.Close()
	buff := bufio.NewReader(f)
	content := ""
	title := ""
	desc := ""
	tag := ""
	keywords := ""
	for {
		line, err := buff.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		if strings.HasPrefix(line, "@title:") {
			content = content + line
			title = strings.TrimRight(line, "\n")
			title = strings.Replace(title, "#", "", 1)
			continue
		} else if strings.HasPrefix(line, "@keywords:") {
			keywords = strings.TrimRight(line, "\n")
			keywords = strings.Replace(title, "@keywords:", "", 1)
			continue
		} else if strings.HasPrefix(line, "@desc:") {
			desc = strings.TrimRight(line, "\n")
			desc = strings.Replace(title, "@desc:", "", 1)
			continue
		} else if strings.HasPrefix(line, "@tag:") {
			tag = strings.TrimRight(line, "\n")
			tag = strings.Replace(title, "@tag:", "", 1)
			continue
		}
		content = content + line
	}
	return title, desc, keywords, content, tag
}

func (item *Item) InsertLedis(typeArticle, path, fileName string) {
	title, desc, keywords, content, tag := FindDetail(path, fileName)
	item.Title = title
	item.Desc = desc
	item.Keywords = keywords
	item.Content = content
	item.Tag = tag
	item.Key = strings.Split(fileName, ".")[0]

	//插入ledis 目录
	conn.HSet([]byte(typeArticle), []byte(item.Key), []byte(item.Title+"|"+item.UpdateTime))
	//插入文章内容
	conn.HSet([]byte(item.Key), []byte("title"), []byte(item.Title))
	conn.HSet([]byte(item.Key), []byte("desc"), []byte(item.Desc))
	conn.HSet([]byte(item.Key), []byte("keywords"), []byte(item.Keywords))
	conn.HSet([]byte(item.Key), []byte("content"), []byte(item.Content))
}

func DeteleArticle(typeArticle string, fileMap map[string]string) {
	fileNames, _ := conn.HKeys([]byte(typeArticle))
	for _, fileName := range fileNames {
		if _, found := fileMap[string(fileName)]; !found {
			//在目录中没有查到该文件，则进行删除
			conn.HDel([]byte(typeArticle), []byte(string(fileName)))
			conn.HDel([]byte(string(fileName)), []byte("title"), []byte("desc"), []byte("keywords"), []byte("content"))
		}
	}
}

func ShowArticleList() {
	time.Sleep(60 * time.Second)
	fileNames, _ := conn.HKeys([]byte("A"))
	i := 0
	for _, fileName := range fileNames {
		i = i + 1
		fmt.Println("manage", i)
		data, _ := conn.HGet([]byte("A"), []byte(string(fileName)))
		fmt.Printf("%s\t%s\t%s\n", "A", string(fileName), string(data))
	}
}
