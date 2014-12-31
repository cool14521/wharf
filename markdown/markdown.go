package markdown

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/shurcool/go/github_flavored_markdown"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

/*
说明：
程序使用存储的数据库为ledis

程序功能
1 对执行参数进行验证（category行为）
2 初始化数据库连接（category属性）
3 同步远程数据
4 数据本地预处理
5 删除掉本地存在，远程库已经删除的数据
6 对存在数据进行添加和更新
*/

var (
	ledisOnce sync.Once
	nowLedis  *ledis.Ledis
	conn      *ledis.DB
)

type Category struct {
	Remote string          //远程git库的地址
	Local  string          //category同步到本地的路径
	Prefix string          //category的前缀名，和router关联，例如存入HSet(a xx.md time=20131112 12:00:00),可通过/a/xx来获取数据
	DocMap map[string]*Doc //category中的文件集
	Conn   *ledis.DB       //操作category的数据库连接
}

type Doc struct {
	Permalink string //文章检索标志
	Title     string //文章标题
	Desc      string //文章描述
	Keywords  string //文章关键字集合
	Updated   string //文章更新时间
	Content   string //文章内容（经过render之后的html格式）
	Tags      string //文章标签集合
	Path      string //文章路径
	Author    string //文章作者
	Views     int64  //阅读次数
}

//远程同步的到本地的操作
func (category *Category) Sync() error {
	if err := category.validate("sync"); err != nil {
		return err
	}
	beego.Trace("[markdown]开始同步git数据")
	//判断本地路径是否存在，不存在则创建
	if !IsDirExist(category.Local) {
		CreateDir(category.Local)
		if err := category.clone(); err != nil {
			return err
		}
		varlength := len(strings.Split(category.Remote, "/"))
		//重新复制本地路径local的值，定位到git对应的目录下
		category.Local = category.Local + "/" + strings.Split(strings.Split(category.Remote, "/")[varlength-1], ".")[0]
	} else {
		//判断本地文件夹存在，是否包含所需要的git库
		varlength := len(strings.Split(category.Remote, "/"))
		githubRepo := strings.Split(strings.Split(category.Remote, "/")[varlength-1], ".")[0]
		//库已经存在
		if repoExist(githubRepo, category.Local) {
			category.Local = category.Local + "/" + strings.Split(strings.Split(category.Remote, "/")[varlength-1], ".")[0]
			if err := category.pull(); err != nil {
				return err
			}
		} else if err := category.clone(); err != nil {
			return err
		}
		category.Local = category.Local + "/" + strings.Split(strings.Split(category.Remote, "/")[varlength-1], ".")[0]
	}
	beego.Trace("[markdown]仓库[", category.Remote, "]同步本地完成")
	return nil
}

//数据预处理
func (category *Category) Render() error {
	if err := category.validate("render"); err != nil {
		return err
	}
	//生成doc集合赋值给category对象
	//1 读取文件数据
	//2 处理文件数据（生成完成Doc数据）
	files, _ := ioutil.ReadDir(category.Local)
	fileMap := make(map[string]*Doc, len(files))
	for _, file := range files {
		if file.Name() == "README.md" || file.Name() == "sitemap.md" || strings.HasPrefix(file.Name(), ".git") {
			continue
		}
		cmd := exec.Command("git", "log", "-1", "--format=\"%ai\"", "--", file.Name())
		cmd.Dir = category.Local
		output, _ := cmd.Output()
		regex, _ := regexp.Compile(`(?m)^"(.*) .*?0800`)
		outputString := string(output)
		result := regex.FindStringSubmatch(outputString)
		var timeString string
		for _, v := range result {
			timeString = v
		}
		doc := new(Doc)
		doc.Updated = timeString
		doc.Path = fmt.Sprint(category.Local, "/", file.Name())
		doc.Permalink = strings.Split(file.Name(), ".")[0]
		if err := doc.generate(); err != nil {
			beego.Error(err)
			continue
		}
		fileMap[doc.Permalink] = doc
	}
	//将fileMap存入到.render的文件中
	bytes, _ := json.Marshal(fileMap)
	if _, err := os.Stat(".render"); err == nil {
		os.Remove(".render")
	}
	err := ioutil.WriteFile(".render", bytes, 0660)
	if err != nil {
		return err
	}
	beego.Trace("[markdown]文件预处理全部完成")
	return nil
}

//数据查询
func (category *Category) Query(isDict bool, permalink ...string) ([]*Doc, error) {
	var err error
	if err = category.validate("query"); err != nil {
		return nil, err
	}
	category.initDB()
	docs := make([]*Doc, 0)
	switch isDict {
	case true:
		//展示category的目录数据
		if len(strings.TrimSpace(category.Prefix)) == 0 {
			err = errors.New("请输入查询目录category的前缀值")
			return nil, err
		}
		permalinks, _ := category.Conn.HKeys([]byte(category.Prefix))
		for _, permalink := range permalinks {
			doc := new(Doc)
			data, _ := category.Conn.HGet([]byte(category.Prefix), []byte(string(permalink)))
			doc.Permalink = string(permalink)
			doc.Title = strings.Split(string(data), "|")[0]
			doc.Updated = strings.Split(string(data), "|")[1]
			docs = append(docs, doc)
		}
	case false:
		if len(strings.TrimSpace(permalink[0])) == 0 {
			err = errors.New("请输入查询markdown文件的permalink值")
			break
		}
		attrs, _ := category.Conn.HKeys([]byte(permalink[0]))
		//未查到文件 返回错误
		if len(attrs) == 0 {
			return nil, errors.New("查询文件不存在")
		}
		doc := new(Doc)
		for _, attr := range attrs {
			data, _ := category.Conn.HGet([]byte(permalink[0]), attr)
			attr2string := string(attr)
			switch attr2string {
			case "permalink":
				doc.Permalink = string(data)
			case "title":
				doc.Title = string(data)
			case "desc":
				doc.Desc = string(data)
			case "keywords":
				doc.Keywords = string(data)
			case "updated":
				doc.Updated = string(data)
			case "content":
				doc.Content = string(data)
			case "tags":
				doc.Tags = string(data)
			case "path":
				doc.Path = string(data)
			case "author":
				doc.Author = string(data)
			case "views":
				doc.Views, _ = strconv.ParseInt(string(data), 0, 64)
			}
		}
		docs = append(docs, doc)
		//查阅数加1
		doc.Views = doc.Views + 1
		category.Conn.HSet([]byte(permalink[0]), []byte("views"), []byte(fmt.Sprint(doc.Views)))
	}
	return docs, nil
}

func (category *Category) Save() error {
	var err error
	if err = category.validate("save"); err != nil {
		return err
	} else if err = category.load(); err != nil {
		return err
	}
	category.initDB()
	//清除掉数据库中多余的部分
	var deleted, updated, insert int //记录删除，更新，插入记录数
	permalinks, _ := category.Conn.HKeys([]byte(category.Prefix))
	for _, permalink := range permalinks {
		beego.Trace(string(permalink))
		if doc, found := category.DocMap[string(permalink)]; !found {
			//在目录中没有查到该文件，则进行删除
			category.Conn.HDel([]byte(category.Prefix), []byte(string(permalink)))
			category.Conn.HDel([]byte(string(permalink)), []byte("title"), []byte("desc"), []byte("keywords"), []byte("content"), []byte("tags"), []byte("author"), []byte("views"), []byte("updated"))
			deleted++
		} else if found && doc.Permalink != "" {
			//如果存在，则更新数据库中数据，从docMap中移除
			category.Conn.HSet([]byte(category.Prefix), []byte(doc.Permalink), []byte(doc.Title+"|"+doc.Updated))
			category.Conn.HSet([]byte(doc.Permalink), []byte("desc"), []byte(doc.Desc))
			category.Conn.HSet([]byte(doc.Permalink), []byte("keywords"), []byte(doc.Keywords))
			category.Conn.HSet([]byte(doc.Permalink), []byte("content"), []byte(doc.Content))
			category.Conn.HSet([]byte(doc.Permalink), []byte("tags"), []byte(doc.Tags))
			category.Conn.HSet([]byte(doc.Permalink), []byte("author"), []byte(doc.Author))
			category.Conn.HSet([]byte(doc.Permalink), []byte("updated"), []byte(doc.Updated))
			delete(category.DocMap, doc.Permalink)
			updated++
		}
	}
	//插入或更新数据库
	for _, doc := range category.DocMap {
		category.Conn.HSet([]byte(category.Prefix), []byte(doc.Permalink), []byte(doc.Title+"|"+doc.Updated))
		//插入文章内容
		category.Conn.HSet([]byte(doc.Permalink), []byte("title"), []byte(doc.Title))
		category.Conn.HSet([]byte(doc.Permalink), []byte("desc"), []byte(doc.Desc))
		category.Conn.HSet([]byte(doc.Permalink), []byte("keywords"), []byte(doc.Keywords))
		category.Conn.HSet([]byte(doc.Permalink), []byte("content"), []byte(doc.Content))
		category.Conn.HSet([]byte(doc.Permalink), []byte("tags"), []byte(doc.Tags))
		category.Conn.HSet([]byte(doc.Permalink), []byte("author"), []byte(doc.Author))
		category.Conn.HSet([]byte(doc.Permalink), []byte("updated"), []byte(doc.Updated))
		category.Conn.HSet([]byte(doc.Permalink), []byte("views"), []byte("0"))
		insert++
	}
	beego.Trace("[markdown]本次操作输入数据", insert, "条，删除数据", deleted, "条，更新数据", updated, "条")
	os.Remove(".render")
	return nil
}

func (category *Category) load() error {
	//加载.render文件，对category.docmap赋值
	bytes, _ := ioutil.ReadFile(".render")
	err := json.Unmarshal(bytes, &category.DocMap)
	if err != nil {
		return errors.New("加载缓存文件失败，请重新执行render操作")
	}
	beego.Trace("[markdown]加载缓存文件成功，开始进行数据库操作")
	return nil
}

func (category *Category) validate(action string) error {
	switch action {
	case "sync":
		if len(strings.TrimSpace(category.Remote)) == 0 || len(strings.TrimSpace(category.Local)) == 0 {
			return errors.New("markdown git地址初始化异常,请赋值remote和local")
		}
	case "render":
		if !IsDirExist(category.Local) {
			return errors.New("本地路径不存在,请执行sync操作")
		} else if files, _ := ioutil.ReadDir(category.Local); len(files) == 0 {
			return errors.New("本地路径不存在文件,无法进行转换处理，请执行sync操作,确认文件已经同步")
		}
	case "save":
		if _, err := os.Stat(".render"); err != nil || len(strings.TrimSpace(category.Prefix)) == 0 {
			return errors.New("请确认是否值之前执行了sync、render的操作")
		}
	}
	return nil
}

func (category *Category) initDB() {
	category.Conn = conn
}

func init() {
	//如果存储路径不存在，则创建路径
	if !IsDirExist(beego.AppConfig.String("markdown::DataDir")) {
		CreateDir(beego.AppConfig.String("markdown::DataDir"))
	}
	initLedisFunc := func() {
		cfg := new(config.Config)
		cfg.DataDir = beego.AppConfig.String("markdown::DataDir")
		var err error
		nowLedis, err = ledis.Open(cfg)
		if err != nil {
			beego.Error(err)
		}
	}
	ledisOnce.Do(initLedisFunc)
	var err error
	db, _ := beego.AppConfig.Int("markdown::Db")
	conn, err = nowLedis.Select(db)
	if err != nil {
		beego.Error(err)
	}
}

func (category *Category) clone() error {
	beego.Trace("[markdown]开始进行克隆操作", "remote=", category.Remote, ";local=", category.Local)
	cmd := exec.Command("git", "clone", category.Remote)
	cmd.Dir = category.Local
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}

func (category *Category) pull() error {
	beego.Trace("[markdown]开始进行pull操作local=", category.Local)
	cmd := exec.Command("git", "pull")
	cmd.Dir = category.Local
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}

func IsDirExist(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fi.IsDir()
}

func CreateDir(path string) {
	oldMask := syscall.Umask(0)
	os.Mkdir(path, os.ModePerm)
	syscall.Umask(oldMask)
}

func (doc *Doc) generate() error {
	f, err := os.Open(doc.Path)
	if err != nil {
		return errors.New(fmt.Sprint(doc.Path, ";文件预处理失败;err=", err))
	}
	defer f.Close()
	buff := bufio.NewReader(f)

	for {
		line, err := buff.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		if strings.HasPrefix(line, "@title:") {
			doc.Title = strings.TrimRight(line, "\n")
			doc.Title = strings.Replace(doc.Title, "@title:", "", 1)
			continue
		} else if strings.HasPrefix(line, "@keywords:") {
			doc.Keywords = strings.TrimRight(line, "\n")
			doc.Keywords = strings.Replace(doc.Keywords, "@keywords:", "", 1)
			continue
		} else if strings.HasPrefix(line, "@desc:") {
			doc.Desc = strings.TrimRight(line, "\n")
			doc.Desc = strings.Replace(doc.Desc, "@desc:", "", 1)
			continue
		} else if strings.HasPrefix(line, "@tags:") {
			doc.Tags = strings.TrimRight(line, "\n")
			doc.Tags = strings.Replace(doc.Tags, "@tags:", "", 1)
			continue
		} else if strings.HasPrefix(line, "@author:") {
			doc.Author = strings.TrimRight(line, "\n")
			doc.Author = strings.Replace(doc.Author, "@author:", "", 1)
			continue
		}
		doc.Content = doc.Content + line
	}
	doc.Content = markdown2html(doc.Content)
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
