package godown

import (
	"errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type Downloader interface {
	Download(url string) (Book, error)
	Get_Book_Name(desc string) (string, error)
	Get_Chapter_Urls(content string) ([]string, bool, error)
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
