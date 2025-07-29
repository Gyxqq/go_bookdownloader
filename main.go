package main

import (
	"fmt"
	"path"
	"regexp"
	"runtime"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

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
	fmt.Print("输入小说主页url:")
	var url string
	fmt.Scanln(&url)
	down := DownloaderImpl{}
	down.name_regex = regexp.MustCompile(`<meta\s+property="og:novel:book_name"\s+content="([^"]+)"\s*/?>`)
	down.chapter_regex = regexp.MustCompile(`<dd><a\s+href\s*=\s*"([^"]+)">`)
	down.title_regex = regexp.MustCompile(`<h1\s+class="wap_none">\s*(.*?)\s*</h1>`)
	down.content_regex = regexp.MustCompile(`<br\s*/?>\s*([^<]+?)\s*<br\s*/?>`)
	book, err := down.Download(url)
	if err != nil {
		logrus.WithField("downloader", "Download").Errorf("Error when download book %s", err)
		return
	}
	for _, chapter := range book.chapters {
		fmt.Printf("%d. %s\n", chapter.index, chapter.title)
		fmt.Println(chapter.content)
	}
}