package main

import (
	"encoding/json"
	"flag"
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"regexp"
	"runtime"
)

type down_config struct {
	Name_regex    string `json:"name_regex"`
	Chapter_regex string `json:"chapter_regex"`
	Title_regex   string `json:"title_regex"`
	Content_regex string `json:"content_regex"`
}

func main() {
	logrus.SetFormatter(&nested.Formatter{
		HideKeys:    true,
		FieldsOrder: []string{"component"},
		CallerFirst: true,
		CustomCallerFormatter: func(f *runtime.Frame) string {
			// 只显示文件名和行号
			_, filename := path.Split(f.File)
			return fmt.Sprintf(" %s:%d", filename, f.Line)
		},
	})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	mainpage := flag.String("u", "", "小说主页url")
	max_threads := flag.Int("t", 20, "下载线程数")
	config := flag.String("c", "", "配置文件")
	outname := flag.String("o", "", "输出文件名")
	flag.Parse()
	if *mainpage == "" {
		fmt.Print("输入小说主页url:")
		fmt.Scanln(&mainpage)
	}
	down := DownloaderImpl{thread_num: *max_threads}
	if *config != "" {
		data, err := os.ReadFile(*config)
		if err != nil {
			logrus.WithField("component", "main").Fatal(err)
		}
		var conf down_config
		err = json.Unmarshal(data, &conf)
		if err != nil {
			logrus.WithField("component", "main").Fatal(err)
		}
		if conf.Chapter_regex == "" || conf.Content_regex == "" || conf.Name_regex == "" || conf.Title_regex == "" {
			logrus.WithField("component", "main").Fatal("err config file")
		}
		down.chapter_regex = regexp.MustCompile(conf.Chapter_regex)
		down.content_regex = regexp.MustCompile(conf.Content_regex)
		down.name_regex = regexp.MustCompile(conf.Name_regex)
		down.title_regex = regexp.MustCompile(conf.Title_regex)
	} else {
		down.name_regex = regexp.MustCompile(`<meta\s+property="og:novel:book_name"\s+content="([^"]+)"\s*/?>`)
		down.chapter_regex = regexp.MustCompile(`<dd><a\s+href\s*=\s*"([^"]+)">`)
		down.title_regex = regexp.MustCompile(`<h1\s+class="wap_none">\s*(.*?)\s*</h1>`)
		down.content_regex = regexp.MustCompile(`<br\s*/?>\s*([^<]+?)\s*<br\s*/?>`)
	}

	book, err := down.Download(*mainpage)
	if err != nil {
		logrus.WithField("downloader", "Download").Errorf("Error when download book %s", err)
		return
	}
	logrus.WithField("downloader", "Download").Info("Download book success")
	if *outname == "" {
		*outname = book.name
	}
	file, err := os.Create(*outname + ".txt")
	if err != nil {
		logrus.WithField("downloader", "Download").Errorf("Error when create file %s", err)
		return
	}
	defer file.Close()
	for _, chapter := range book.chapters {
		file.WriteString(chapter.title + "\n" + chapter.content + "\n")
		logrus.WithField("downloader", "Download").Infof("Write chapter %s", chapter.title)
	}
	logrus.WithField("downloader", "Download").Info("Write book success")
}
