package markdown

import (
	"bufio"
	"encoding/json"
	"errors"
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
func (doc *Doc) Sync() error {
	if err := doc.verify("sync"); err != nil {
		return err
	}
	fmt.Println("....开始同步git数据")
	//判断本地路径是否存在，不存在则创建
	if !isDirExist(doc.Local) {
		createDir(doc.Local)
		if err := doc.clone(); err != nil {
			return err
		}
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
			if err := doc.pull(); err != nil {
				return err
			}
		} else if err := doc.clone(); err != nil {
			return err
		}
		doc.Local = doc.Local + "/" + strings.Split(strings.Split(doc.Remote, "/")[varlength-1], ".")[0]
	}
	fmt.Println("....仓库[", doc.Remote, "]同步本地完成")
	return nil
}

//数据预处理
func (doc *Doc) Render() error {
	if err := doc.verify("render"); err != nil {
		return err
	}
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
		if err := item.generate(); err != nil {
			fmt.Println(err)
			continue
		}
		fileMap[item.key] = item
	}

	//将fileMap存入到.render的文件中
	bytes, _ := json.Marshal(fileMap)
	if _, err := os.Stat(".render"); err == nil {
		os.Remove(".render")
	} else if err := ioutil.WriteFile(".render", bytes, 0660); err != nil {
		return err
	}
	fmt.Println("....文件预处理全部完成")
	return nil
}

//数据查询
func (doc *Doc) Query(isDict bool, key string) error {
	var err error
	if err = doc.verify("query"); err != nil {
		return err
	}
	doc.initDB()
	switch isDict {
	case true:
		//展示doc的目录数据
		if len(strings.TrimSpace(doc.Prefix)) == 0 {
			err = errors.New("...请输入查询目录doc的前缀值")
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
			err = errors.New("...请输入查询markdown文件的key值")
			break
		}
		attrs, _ := doc.conn.HKeys([]byte(key))
		for _, attr := range attrs {
			data, _ := doc.conn.HGet([]byte(key), []byte(string(attr)))
			fmt.Printf("%s\t%s\t%s\n", key, string(attr), string(data))
		}
	}
	return err
}

func (doc *Doc) Save() error {
	var err error
	if err = doc.verify("save"); err != nil {
		return err
	} else if err = doc.load(); err != nil {
		return err
	} else if err = doc.initDB(); err != nil {
		return err
	}
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
	return nil
}

func (doc *Doc) load() error {
	//加载.render文件，对doc.itemmap赋值
	bytes, _ := ioutil.ReadFile(".render")
	err := json.Unmarshal(bytes, &doc.itemMap)
	if err != nil {
		return errors.New("....加载缓存文件失败，请重新执行render操作")
	}
	fmt.Println("....加载缓存文件成功，开始进行数据库操作")
	return nil
}

func (doc *Doc) verify(action string) error {
	var err error
	switch action {
	case "sync":
		if len(strings.TrimSpace(doc.Remote)) == 0 || len(strings.TrimSpace(doc.Local)) == 0 {
			err = errors.New("....markdown git地址初始化异常,请赋值remote和local")
		}
	case "render":
		if !isDirExist(doc.Local) {
			err = errors.New("....本地路径不存在,请执行sync操作")
			break
		} else if files, _ := ioutil.ReadDir(doc.Local); len(files) == 0 {
			err = errors.New("....本地路径不存在文件,无法进行转换处理，请执行sync操作,确认文件已经同步")
		}
	case "save":
		if _, err := os.Stat(".render"); err != nil || len(strings.TrimSpace(doc.Prefix)) == 0 {
			err = errors.New("....请确认是否值之前执行了sync、render的操作")
		} else if len(strings.TrimSpace(doc.Storage)) == 0 {
			err = errors.New("....请输入数据文件的存储路径")
		}
	case "query":
		if len(strings.TrimSpace(doc.Storage)) == 0 {
			err = errors.New("....请输入数据文件的存储路径")
		}
	}
	return err
}

func (doc *Doc) initDB() error {
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
		return err
	}
	fmt.Println("....初始化数据库成功")
	return nil
}

func (doc *Doc) clone() error {
	fmt.Println("....开始进行克隆操作", "remote=", doc.Remote, ";local=", doc.Local)
	cmd := exec.Command("git", "clone", doc.Remote)
	cmd.Dir = doc.Local
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}

func (doc *Doc) pull() error {
	fmt.Println("....开始进行pull操作local=", doc.Local)
	cmd := exec.Command("git", "pull")
	cmd.Dir = doc.Local
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
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

func (item *Item) generate() error {
	f, err := os.Open(item.path)
	if err != nil {
		return errors.New(fmt.Sprint(item.path, ";文件预处理失败;err=", err))
	}
	defer f.Close()
	buff := bufio.NewReader(f)

	for {
		line, err := buff.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		if strings.HasPrefix(line, "@title:") {
			item.title = strings.TrimRight(line, "\n")
			item.title = strings.Replace(item.title, "@title", "", 1)
			continue
		} else if strings.HasPrefix(line, "@keywords:") {
			item.keywords = strings.TrimRight(line, "\n")
			item.keywords = strings.Replace(item.keywords, "@keywords:", "", 1)
			continue
		} else if strings.HasPrefix(line, "@desc:") {
			item.desc = strings.TrimRight(line, "\n")
			item.desc = strings.Replace(item.desc, "@desc:", "", 1)
			continue
		} else if strings.HasPrefix(line, "@tags:") {
			item.tags = strings.TrimRight(line, "\n")
			item.tags = strings.Replace(item.tags, "@tags:", "", 1)
			continue
		}
		item.content = item.content + line
	}
	item.content = markdown2html(item.content)
	fmt.Println("....", item.path, ";文件预处理完成;success")
	return nil
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
