package main

import (
	"bufio"
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
	if *mainpage == "" && *input == "" { // if both are empty, ask for url
		fmt.Print("输入小说主页url:")
		var s string
		fmt.Scanln(&s)
		mainpage = &s
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
		down.name_regex = regexp.MustCompile(`<meta\s+property="og:novel:book_name"\s+content="([^"]+)"\s*\/?>`)
		down.chapter_regex = regexp.MustCompile(`<dd><a\s+href\s*=\s*"([^"]+)">`)
		down.title_regex = regexp.MustCompile(`<h1\s+class="wap_none">\s*(.*?)\s*<\/h1>`)
		down.content_regex = regexp.MustCompile(`<br\s*\/?>\s*([^<]+?)\s*<br\s*\/?>`)
	}
	var err error
	var book Book
	var chapter_urls []string

	if *input == "" {
		book, chapter_urls, err = down.GetBookInfoAndChapterURLs(*mainpage)
	} else {
		var data []byte
		data, err = os.ReadFile(*input)
		if err != nil {
			logrus.WithField("downloader", "main").Errorf("Error when read input file %s", err)
			book, chapter_urls, err = down.GetBookInfoAndChapterURLs(*mainpage)
		} else {
			logrus.WithField("downloader", "main").Info("Read mainpage from file")
			book, chapter_urls, err = down.GetBookInfoAndChapterURLs_from_file(string(data), *mainpage)
		}
	}

	if err != nil {
		logrus.WithField("downloader", "main").Errorf("Error when getting book info %s", err)
		os.Exit(-1)
	}

	chapter_url_map := make(map[string]int)
	for i, url := range chapter_urls {
		chapter_url_map[url] = i
	}

	if *outname == "" {
		*outname = book.name
	}
	progress_filename := *outname + ".progress"
	var downloaded_urls = make(map[string]bool)

	if _, err := os.Stat(progress_filename); err == nil {
		logrus.WithField("component", "main").Infof("Progress file found: %s", progress_filename)
		file, err := os.Open(progress_filename)
		if err != nil {
			logrus.WithField("component", "main").Warnf("Could not read progress file: %s", err)
		} else {
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				downloaded_urls[scanner.Text()] = true
			}
			file.Close()
		}
		logrus.WithField("component", "main").Infof("Found %d downloaded chapters in progress file.", len(downloaded_urls))
	}

	var urls_to_download []string
	for _, url := range chapter_urls {
		if !downloaded_urls[url] {
			urls_to_download = append(urls_to_download, url)
		}
	}

	logrus.WithField("component", "main").Infof("%d chapters already downloaded. %d new chapters to download.", len(downloaded_urls), len(urls_to_download))

	if len(urls_to_download) == 0 {
		logrus.WithField("component", "main").Info("All chapters already downloaded. Exiting.")
		os.Exit(0)
	}

	new_chapters, err := down.Get_Chapters(urls_to_download, chapter_url_map)
	if err != nil {
		logrus.WithField("downloader", "main").Errorf("Error when downloading chapters %s", err)
		os.Exit(-1)
	}

	book.chapters = new_chapters

	output_file, err := os.OpenFile(*outname+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logrus.WithField("downloader", "main").Errorf("Error when opening output file %s", err)
		os.Exit(-1)
	}
	defer output_file.Close()

	progress_file, err := os.OpenFile(progress_filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logrus.WithField("downloader", "main").Errorf("Error when opening progress file %s", err)
		os.Exit(-1)
	}
	defer progress_file.Close()

	for _, chapter := range book.chapters {
		_, err := output_file.WriteString(chapter.title + "\n" + chapter.content + "\n")
		if err != nil {
			logrus.WithField("downloader", "main").Errorf("Error writing chapter %s to file: %s", chapter.title, err)
			continue // Or handle error more gracefully
		}
		_, err = progress_file.WriteString(chapter.url + "\n")
		if err != nil {
			logrus.WithField("downloader", "main").Errorf("Error writing to progress file for chapter %s: %s", chapter.title, err)
		}
		logrus.WithField("downloader", "main").Infof("Wrote chapter %s", chapter.title)
	}

	logrus.WithField("downloader", "main").Info("Write book success")
}