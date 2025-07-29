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
	log_level := flag.String("l", "info", "日志等级")
	input := flag.String("f", "", "从文件读入主页内容")
	flag.Parse()
	switch *log_level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
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
		logrus.WithField("component", "main").Infof("load chapter regex %s", conf.Chapter_regex)
		down.content_regex = regexp.MustCompile(conf.Content_regex)
		logrus.WithField("component", "main").Infof("load content regex %s", conf.Content_regex)
		down.name_regex = regexp.MustCompile(conf.Name_regex)
		logrus.WithField("component", "main").Infof("load name regex %s", conf.Name_regex)
		down.title_regex = regexp.MustCompile(conf.Title_regex)
		logrus.WithField("component", "main").Infof("load title regex %s", conf.Title_regex)
	} else {
		down.name_regex = regexp.MustCompile(`<meta\s+property="og:novel:book_name"\s+content="([^"]+)"\s*/?>`)
		down.chapter_regex = regexp.MustCompile(`<dd><a\s+href\s*=\s*"([^"]+)">`)
		down.title_regex = regexp.MustCompile(`<h1\s+class="wap_none">\s*(.*?)\s*</h1>`)
		down.content_regex = regexp.MustCompile(`<br\s*/?>\s*([^<]+?)\s*<br\s*/?>`)
	}
	var err error
	var book Book
	if *input == "" {
		book, err = down.Download(*mainpage)
	} else {
		var data []byte
		data, err = os.ReadFile(*input)
		if err != nil {
			logrus.WithField("downloader", "Download").Errorf("Error when read input file %s", err)
			book, err = down.Download(*mainpage)
		}
		logrus.WithField("downloader", "Download").Info("Read mainpage from file")
		book, err = down.Download_from_file(string(data), *mainpage)
	}

	if err != nil {
		logrus.WithField("downloader", "Download").Errorf("Error when download book %s", err)
		os.Exit(-1)
	}
	logrus.WithField("downloader", "Download").Info("Download book success")
	if *outname == "" {
		*outname = book.name
	}
	file, err := os.Create(*outname + ".txt")
	if err != nil {
		logrus.WithField("downloader", "Download").Errorf("Error when create file %s", err)
		os.Exit(-1)
	}
	defer file.Close()
	for _, chapter := range book.chapters {
		file.WriteString(chapter.title + "\n" + chapter.content + "\n")
		logrus.WithField("downloader", "Download").Infof("Write chapter %s", chapter.title)
	}
	logrus.WithField("downloader", "Download").Info("Write book success")

}
