package markdown

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/shurcooL/go/github_flavored_markdown"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
)

/*
说明：
程序使用存储的数据库为ledis

程序功能
1 对执行参数进行验证（doc行为）
2 初始化数据库连接（doc属性）
3 同步远程数据
4 数据本地预处理
5 删除掉本地存在，远程库已经删除的数据
6 对存在数据进行添加和更新
*/

type Doc struct {
	Remote  string           //远程git库的地址
	Local   string           //doc同步到本地的路径
	Prefix  string           //doc的前缀名，和router关联，例如存入HSet(a xx.md time=20131112 12:00:00),可通过/a/xx来获取数据
	Storage string           //doc处理后存入数据文件的路径
	Db      int              //doc所在数据库
	itemMap map[string]*Item //doc中的文件集
	conn    *ledis.DB        //操作doc的数据库连接
}

type Item struct {
	key      string //文件关键字
	title    string //文件标题
	desc     string //文件描述
	keywords string //文件关键字集合
	updated  string //文件更新时间
	content  string //文件内容（经过render之后的html格式）
	tags     string //文件标签集合
	path     string //文件路径
}

//远程同步的到本地的操作
func (doc *Doc) Sync() {
	doc.verify("sync")
	fmt.Println("....开始同步git数据")
	//判断本地路径是否存在，不存在则创建
	if !isDirExist(doc.Local) {
		createDir(doc.Local)
		doc.clone()
		varlength := len(strings.Split(doc.Remote, "/"))
		//重新复制本地路径local的值，定位到git对应的目录下
		doc.Local = doc.Local + "/" + strings.Split(strings.Split(doc.Remote, "/")[varlength-1], ".")[0]
	} else {
		//判断本地文件夹存在，是否包含所需要的git库
		varlength := len(strings.Split(doc.Remote, "/"))
		githubRepo := strings.Split(strings.Split(doc.Remote, "/")[varlength-1], ".")[0]
		//库已经存在
		if repoExist(githubRepo, doc.Local) {
			doc.Local = doc.Local + "/" + strings.Split(strings.Split(doc.Remote, "/")[varlength-1], ".")[0]
			doc.pull()
		} else {
			//库不存在
			doc.clone()
			doc.Local = doc.Local + "/" + strings.Split(strings.Split(doc.Remote, "/")[varlength-1], ".")[0]
		}
	}
	fmt.Println("....仓库[", doc.Remote, "]同步本地完成")
}

//数据预处理
func (doc *Doc) Transform() {
	doc.verify("transform")
	//生成item集合赋值给doc对象
	//1 读取文件数据
	//2 处理文件数据（生成完成Item数据）
	files, _ := ioutil.ReadDir(doc.Local)
	fileMap := make(map[string]*Item, len(files))
	for _, file := range files {
		if file.Name() == "README.md" || file.Name() == "rss.md" || file.Name() == "sitemap.md" || strings.HasPrefix(file.Name(), ".git") {
			continue
		}
		cmd := exec.Command("git", "log", "-1", "--format=\"%ai\"", "--", file.Name())
		cmd.Dir = doc.Local
		output, _ := cmd.Output()
		regex, _ := regexp.Compile(`(?m)^"(.*) .*?0800`)
		outputString := string(output)
		result := regex.FindStringSubmatch(outputString)
		var timeString string
		for _, v := range result {
			timeString = v
		}
		item := new(Item)
		item.updated = timeString
		item.path = fmt.Sprint(doc.Local, "/", file.Name())
		item.key = strings.Split(file.Name(), ".")[0]
		fileMap[item.key] = item
	}
	handleEnd := make(chan bool)
	itemChans := make(chan string, len(fileMap))
	go func() {
		var i int
		for {
			select {
			case msg := <-itemChans:
				fmt.Println(msg)
				i++
			}
			if i == len(fileMap) {
				handleEnd <- true
			}
		}
	}()

	for _, item := range fileMap {
		go func(item *Item, itemChan chan string) {
			msg := item.generate()
			itemChan <- msg
		}(item, itemChans)
	}

	finish := <-handleEnd
	if finish {
		doc.itemMap = fileMap
	}
	fmt.Println("....文件预处理全部完成")
}

//数据查询
func (doc *Doc) Query(isDict bool, key string) {
	doc.initDB()
	switch isDict {
	case true:
		//展示doc的目录数据
		if len(strings.TrimSpace(doc.Prefix)) == 0 {
			fmt.Println("...请输入查询目录doc的前缀值")
			break
		}
		fileNames, _ := doc.conn.HKeys([]byte(doc.Prefix))
		i := 0
		for _, fileName := range fileNames {
			i = i + 1
			fmt.Println("manage", i)
			data, _ := doc.conn.HGet([]byte(doc.Prefix), []byte(string(fileName)))
			fmt.Printf("%s\t%s\t%s\n", doc.Prefix, string(fileName), string(data))
		}
	case false:
		if len(strings.TrimSpace(key)) == 0 {
			fmt.Println("...请输入查询markdown文件的key值")
			break
		}
		attrs, _ := doc.conn.HKeys([]byte(key))
		for _, attr := range attrs {
			data, _ := doc.conn.HGet([]byte(key), []byte(string(attr)))
			fmt.Printf("%s\t%s\t%s\n", key, string(attr), string(data))
		}
	}
}

func (doc *Doc) Save() {
	doc.verify("save")
	doc.initDB()
	//清除掉数据库中多余的部分
	var i int //记录删除记录数
	fileNames, _ := doc.conn.HKeys([]byte(doc.Prefix))
	for _, fileName := range fileNames {
		if _, found := doc.itemMap[string(fileName)]; !found {
			//在目录中没有查到该文件，则进行删除
			doc.conn.HDel([]byte(doc.Prefix), []byte(string(fileName)))
			doc.conn.HDel([]byte(string(fileName)), []byte("title"), []byte("desc"), []byte("keywords"), []byte("content"), []byte("tags"))
			i++
		}
	}
	fmt.Println("....已经删除数据库中多余数据")
	//插入或更新数据库
	for _, item := range doc.itemMap {
		doc.conn.HSet([]byte(doc.Prefix), []byte(item.key), []byte(item.title+"|"+item.updated))
		//插入文章内容
		doc.conn.HSet([]byte(item.key), []byte("title"), []byte(item.title))
		doc.conn.HSet([]byte(item.key), []byte("desc"), []byte(item.desc))
		doc.conn.HSet([]byte(item.key), []byte("keywords"), []byte(item.keywords))
		doc.conn.HSet([]byte(item.key), []byte("content"), []byte(item.content))
		doc.conn.HSet([]byte(item.key), []byte("tags"), []byte(item.tags))
	}
	fmt.Println("....插入或更新数据成功")
}

func (doc *Doc) verify(action string) {
	switch action {
	case "sync":
		if len(strings.TrimSpace(doc.Remote)) == 0 || len(strings.TrimSpace(doc.Local)) == 0 {
			panic("....markdown git地址初始化异常,请赋值remote和local")
		}
	case "transform":
		if !isDirExist(doc.Local) {
			panic("....本地路径不存在,请执行sync操作")
		} else {
			if files, _ := ioutil.ReadDir(doc.Local); len(files) == 0 {
				panic("....本地路径不存在文件,无法进行转换处理，请执行sync操作,确认文件已经同步")
			}
		}
	case "save":
		if len(doc.itemMap) == 0 || len(strings.TrimSpace(doc.Prefix)) == 0 {
			panic("....请确认是否值之前执行了sync、transform的操作")
		} else {
			if len(strings.TrimSpace(doc.Storage)) == 0 {
				panic("....请输入数据文件的存储路径")
			}
		}
	case "query":
		if len(strings.TrimSpace(doc.Storage)) == 0 {
			panic("....请输入数据文件的存储路径")
		}
	}
}

func (doc *Doc) initDB() {
	//如果存储路径不存在，则创建路径
	if !isDirExist(doc.Storage) {
		createDir(doc.Storage)
	}
	cfg := new(config.Config)
	cfg.DataDir = doc.Storage
	var err error
	nowLedis, err := ledis.Open(cfg)
	doc.conn, err = nowLedis.Select(doc.Db)
	if err != nil {
		panic(err)
	}
	fmt.Println("....初始化数据库成功")
}

func (doc *Doc) clone() {
	fmt.Println("....开始进行克隆操作", "remote=", doc.Remote, ";local=", doc.Local)
	cmd := exec.Command("git", "clone", doc.Remote)
	cmd.Dir = doc.Local
	_, err := cmd.Output()
	if err != nil {
		panic(err)
	}
}

func (doc *Doc) pull() {
	fmt.Println("....开始进行pull操作local=", doc.Local)
	cmd := exec.Command("git", "pull")
	cmd.Dir = doc.Local
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

func (item *Item) generate() string {
	f, err := os.Open(item.path)
	if err != nil {
		return fmt.Sprint(item.path, ";文件预处理失败;err=", err)
	}
	defer f.Close()
	buff := bufio.NewReader(f)
	content := ""
	title := ""
	desc := ""
	tags := ""
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
		} else if strings.HasPrefix(line, "@tags:") {
			tags = strings.TrimRight(line, "\n")
			tags = strings.Replace(title, "@tags:", "", 1)
			continue
		}
		content = content + line
	}
	item.title = title
	item.keywords = keywords
	item.tags = tags
	item.desc = desc
	item.content = markdown2html(content)
	return fmt.Sprint("....", item.path, ";文件预处理完成;success")
}

func markdown2html(content string) string {
	output := github_flavored_markdown.Markdown([]byte(content))
	body := template.HTML(output)
	return (fmt.Sprint(body))
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
