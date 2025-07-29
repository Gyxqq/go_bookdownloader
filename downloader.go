package godown

import (
	"errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

type Downloader interface {
	Download(url string) (Book, error)
	Get_Book_Name(desc string) (string, error)
	Get_Chapter_Urls(content string) (urls []string, prefix bool, err error)
	Get_Chapters(urls []string) ([]Chapter, error)
	Get_Chapter_Content(url string) (title string, content string, err error)
}
type Book struct {
	name     string
	chapters []Chapter
}
type Chapter struct {
	index   int
	title   string
	content string
}
type DownloaderImpl struct {
	Client        *http.Client
	name_regex    *regexp.Regexp
	chapter_regex *regexp.Regexp
	title_regex   *regexp.Regexp
	content_regex *regexp.Regexp
}

func (d *DownloaderImpl) Download(url string) (Book, error) {
	d.Client = &http.Client{}
	book := Book{}
	resp, err := d.Client.Get(url)
	if err != nil {
		logrus.WithField("downloader", "Download").Error("Error when get book main page")
		return book, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithField("downloader", "Download").Error("Error when io read book main page")
		return book, err
	}
	book.name, err = d.Get_Book_Name(string(body))
	if err != nil {
		logrus.WithField("downloader", "Download").Error("Error when get book name")
		return book, err
	}
	chapter_urls, prefix, err := d.Get_Chapter_Urls(string(body))
	if err != nil {
		logrus.WithField("downloader", "Download").Error("Error when get chapter urls")
		return book, err
	}
	if !prefix {
		re := regexp.MustCompile(`(https?)://([a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`)
		match := re.FindStringSubmatch(url)
		if match == nil {
			logrus.WithField("downloader", "Download").Error("Error when get domain")
			return book, errors.New("error when get domain")
		}
		for i := range chapter_urls {
			chapter_urls[i] = match[0] + chapter_urls[i]
		}
	}
	return book, nil
}

func (d *DownloaderImpl) Get_Book_Name(desc string) (string, error) {
	match := d.name_regex.FindStringSubmatch(desc)
	if len(match) < 1 || match[1] == "" {
		logrus.WithField("downloader", "Get_Book_Name").Error("No book name found")
		return "", errors.New("no book name found")
	}
	return match[1], nil
}

func (d *DownloaderImpl) Get_Chapter_Urls(content string) ([]string, bool, error) {
	allmatch := d.chapter_regex.FindAllStringSubmatch(content, -1)
	if len(allmatch) < 1 {
		logrus.WithField("downloader", "Get_Chapter_Urls").Error("No chapter urls found")
		return nil, false, errors.New("no chapter urls found")
	}
	var urls []string
	for _, match := range allmatch {
		urls = append(urls, match[1])
	}
	if strings.HasPrefix(urls[0], "http") {
		return urls, true, nil
	} else {
		return urls, false, nil
	}
}

func (d *DownloaderImpl) Get_Chapters(urls []string) ([]Chapter, error) {
	logrus.WithField("component", "Get_Chapters").Info("start getting chapters...")
	var chapters []Chapter
	var wg sync.WaitGroup
	var mut sync.Mutex
	ch := make(chan struct{}, 100)
	for url_index, url := range urls {
		wg.Add(1)
		go func(index int, url string) {
			defer wg.Done()
			ch <- struct{}{}
			defer func() {
				<-ch
			}()
			title, content, err := d.Get_Chapter_Content(url)
			if err != nil {
				logrus.WithField("component", "Get_Chapters").Errorf("error getting chapter: %d content", index)
				return
			}
			mut.Lock()
			chapters = append(chapters, Chapter{index: index, title: title, content: content})
			mut.Unlock()
		}(url_index, url)

	}
	return chapters, nil
}

func (d *DownloaderImpl) Get_Chapter_Content(url string) (title string, content string, err error) {
	logrus.WithField("component", "Get_Chapter_Content").Infof("start getting chapter: %s", url)
	resp, err := d.Client.Get(url)
	if err != nil {
		logrus.WithField("component", "Get_Chapter_Content").Errorf("error getting chapter: %s", url)
		return "", "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithField("component", "Get_Chapter_Content").Errorf("error reading body: %s", url)
		return "", "", err
	}
	title = d.title_regex.FindStringSubmatch(string(body))[1]
	contents := d.content_regex.FindAllStringSubmatch(string(body), -1)
	for _, c := range contents {
		content = content + c[1]
	}
	return title, content, nil
}
