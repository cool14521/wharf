package markdown

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
)

var conn *ledis.DB
var DbAddress string

var GitAddress string
var Local string
var Tag string

type Item struct {
	Key        string
	Title      string
	Desc       string
	Keywords   string
	UpdateTime string
	Content    string
	Tag        string
}

type GitSync struct {
	remote string
	local  string
	tag    string
}

func initDB(DbAddress string) {
	//初始化ledis数据库配置，默认为当前目录下的ledis文件夹中
	var path string
	var dbnumber int
	if len(strings.TrimSpace(DbAddress)) == 0 {
		path = "ledis"
		dbnumber = 0
		createDir(path)
	} else {
		//DbAddress example    /data/ledis?select=1
		path = strings.Split(strings.TrimSpace(DbAddress), "?")[0]
		//如果路径不存在，则创建路径
		if !isDirExist(path) {
			createDir(path)
		}
		var err error
		dbnumber, err = strconv.Atoi(strings.Split(strings.Split(strings.TrimSpace(DbAddress), "?")[1], "=")[1])
		if err != nil {
			dbnumber = 0
		}
	}
	cfg := new(config.Config)
	cfg.DataDir = path
	var err error
	nowLedis, err := ledis.Open(cfg)
	conn, err = nowLedis.Select(dbnumber)
	if err != nil {
		panic(err)
	}
	fmt.Println("....初始化数据库成功")
}

func initGit(gitAddress, local, tag string) *GitSync {
	if len(strings.TrimSpace(gitAddress)) == 0 || len(strings.TrimSpace(gitAddress)) == 0 || len(strings.TrimSpace(gitAddress)) == 0 {
		panic("....markdown git地址初始化异常")
	}
	//将git同步的对象初始化
	gitSync := new(GitSync)
	gitSync.remote = gitAddress
	gitSync.local = local
	gitSync.tag = tag

	fmt.Println("....初始化git同步对象成功")
	return gitSync
}

func gitClone(remote, local string) {
	fmt.Println("....开始进行克隆操作", "remote=", remote, ";local=", local)
	cmd := exec.Command("git", "clone", remote)
	cmd.Dir = local
	_, err := cmd.Output()
	if err != nil {
		panic(err)
	}
}

func gitPull(local string) {
	fmt.Println("....开始进行pull操作local=", local)
	cmd := exec.Command("git", "pull")
	cmd.Dir = local
	_, err := cmd.Output()
	if err != nil {
		panic(err)
	}
}

func isDirExist(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fi.IsDir()
}

func createDir(path string) {
	oldMask := syscall.Umask(0)
	os.Mkdir(path, os.ModePerm)
	syscall.Umask(oldMask)
}

func generateDict(tag, path, remote string) int {
	fmt.Println("....开始在ledis中生成目录")
	files, _ := ioutil.ReadDir(path)
	//设定协程数量
	fileMap := make(map[string]*Item, len(files))
	for _, file := range files {
		if file.Name() == "README.md" || file.Name() == "rss.md" || file.Name() == "sitemap.md" || strings.HasPrefix(file.Name(), ".git") {
			continue
		}
		cmd := exec.Command("git", "log", "-1", "--format=\"%ai\"", "--", file.Name())
		cmd.Dir = path
		output, _ := cmd.Output()
		regex, _ := regexp.Compile(`(?m)^"(.*) .*?0800`)
		outputString := string(output)
		result := regex.FindStringSubmatch(outputString)
		var timeString string
		for _, v := range result {
			timeString = v
		}
		item := new(Item)
		item.UpdateTime = timeString
		item.Key = strings.Split(file.Name(), ".")[0]
		fileMap[item.Key] = item
	}
	//删除数据库中多余的数据
	DeteleArticle(tag, fileMap)
	//定义线程处理文件插入article
	endChan := make(chan string)
	fileChan := make(chan bool, len(fileMap))
	for i, _ := range fileMap {
		go func(i string) {
			fileMap[i].InsertLedis(tag, path, fileMap[i].Key)
			fileChan <- true
		}(i)
	}
	//
	var i int
	go func() {
		for {
			select {
			case <-fileChan:
				i++
			}
			if i == len(fileMap) {
				endChan <- fmt.Sprint("....仓库[", remote, "]", "数据同步地址=", path, ";数据同步完成,共处理数据", len(fileMap), "条")
				break
			}
		}
	}()
	msg := <-endChan
	fmt.Println(msg)
	return len(fileMap)
}

func syncAndSave(gitSync *GitSync) {
	fmt.Println("....开始同步git数据")
	//判断本地路径是否存在，不存在则创建
	if !isDirExist(gitSync.local) {
		createDir(gitSync.local)
		gitClone(gitSync.remote, gitSync.local)
		varlength := len(strings.Split(gitSync.remote, "/"))
		//重新复制本地路径local的值，定位到git对应的目录下
		gitSync.local = gitSync.local + "/" + strings.Split(strings.Split(gitSync.remote, "/")[varlength-1], ".")[0]
	} else {
		//判断本地文件夹存在，是否包含所需要的git库
		varlength := len(strings.Split(gitSync.remote, "/"))
		githubRepo := strings.Split(strings.Split(gitSync.remote, "/")[varlength-1], ".")[0]
		//库已经存在
		if repoExist(githubRepo, gitSync.local) {
			gitSync.local = gitSync.local + "/" + strings.Split(strings.Split(gitSync.remote, "/")[varlength-1], ".")[0]
			gitPull(gitSync.local)
		} else {
			//库不存在
			gitClone(gitSync.remote, gitSync.local)
			gitSync.local = gitSync.local + "/" + strings.Split(strings.Split(gitSync.remote, "/")[varlength-1], ".")[0]
		}
	}
	//数据同步完成，开始进行存储
	fmt.Println("....仓库[", gitSync.remote, "]同步本地完成,准备插入数据到ledis中")
	generateDict(gitSync.tag, gitSync.local, gitSync.remote)
}

func Run() {
	//初始化ledis数据库连接
	initDB(DbAddress)
	//初始化git仓库
	gitSync := initGit(GitAddress, Local, Tag)

	//进行git库的同步操作
	syncAndSave(gitSync)
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
			title = strings.TrimRight(line, "\n")
			title = strings.Replace(title, "@title", "", 1)
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

func (item *Item) InsertLedis(tag, path, fileName string) {
	title, desc, keywords, content, tag := FindDetail(path, fileName)
	item.Title = title
	item.Keywords = keywords
	item.Content = content
	item.Tag = tag
	item.Desc = desc
	item.Key = strings.Split(fileName, ".")[0]

	//插入ledis 目录
	fmt.Println(tag, item.Key, item.Title+"|"+item.UpdateTime)
	fmt.Println(item.Key, "title", item.Title)
	fmt.Println(item.Key, "desc", item.Desc)
	fmt.Println(item.Key, "keywords", item.Keywords)
	fmt.Println(item.Key, "content", item.Content)

	conn.HSet([]byte(tag), []byte(item.Key), []byte(item.Title+"|"+item.UpdateTime))
	//插入文章内容
	conn.HSet([]byte(item.Key), []byte("title"), []byte(item.Title))
	conn.HSet([]byte(item.Key), []byte("desc"), []byte(item.Desc))
	conn.HSet([]byte(item.Key), []byte("keywords"), []byte(item.Keywords))
	conn.HSet([]byte(item.Key), []byte("content"), []byte(item.Content))
}

//删除掉目录中没有，但是数据库中存在的数据
func DeteleArticle(tag string, fileMap map[string]*Item) {
	fileNames, _ := conn.HKeys([]byte(tag))
	for _, fileName := range fileNames {
		if _, found := fileMap[string(fileName)]; !found {
			//在目录中没有查到该文件，则进行删除
			conn.HDel([]byte(tag), []byte(string(fileName)))
			conn.HDel([]byte(string(fileName)), []byte("title"), []byte("desc"), []byte("keywords"), []byte("content"))
		}
	}
}

func Show(tag string) {
	initDB(DbAddress)
	fmt.Println("进入展示数据")
	fileNames, _ := conn.HKeys([]byte(tag))
	i := 0
	for _, fileName := range fileNames {
		i = i + 1
		fmt.Println("manage", i)
		data, _ := conn.HGet([]byte(tag), []byte(string(fileName)))
		fmt.Printf("%s\t%s\t%s\n", tag, string(fileName), string(data))
	}
}

func repoExist(namespace, path string) bool {
	files, _ := ioutil.ReadDir(path)
	for _, file := range files {
		if file.Name() == namespace {
			return true
		}
	}
	return false
}
