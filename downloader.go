package main

import (
	"errors"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Downloader interface {
	GetBookInfoAndChapterURLs(url string) (Book, []string, error)
	GetBookInfoAndChapterURLs_from_file(file_content string, url string) (Book, []string, error)
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
	url     string
}
type DownloaderImpl struct {
	Client        *http.Client
	name_regex    *regexp.Regexp
	chapter_regex *regexp.Regexp
	title_regex   *regexp.Regexp
	content_regex *regexp.Regexp
	thread_num    int
}

func uniqueStrings(input []string) []string {
	seen := make(map[string]struct{})
	result := []string{}

	for _, str := range input {
		if _, exists := seen[str]; !exists {
			seen[str] = struct{}{}
			result = append(result, str)
		}
	}

	return result
}

func (d *DownloaderImpl) GetBookInfoAndChapterURLs(url string) (Book, []string, error) {
	d.Client = &http.Client{}
	book := Book{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logrus.WithField("downloader", "GetBookInfoAndChapterURLs").Error("Error when create request")
		return book, nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36")
	resp, err := d.Client.Do(req)
	if err != nil {
		logrus.WithField("downloader", "GetBookInfoAndChapterURLs").Error("Error when get book main page")
		return book, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithField("downloader", "GetBookInfoAndChapterURLs").Error("Error when io read book main page")
		return book, nil, err
	}
	book.name, err = d.Get_Book_Name(string(body))
	if err != nil {
		logrus.WithField("downloader", "GetBookInfoAndChapterURLs").Error("Error when get book name")
		return book, nil, err
	}
	logrus.WithField("downloader", "GetBookInfoAndChapterURLs").Info("Get book name: " + book.name)
	chapter_urls, prefix, err := d.Get_Chapter_Urls(string(body))
	if err != nil {
		logrus.WithField("downloader", "GetBookInfoAndChapterURLs").Error("Error when get chapter urls")
		return book, nil, err
	}
	if !prefix {
		re := regexp.MustCompile(`(https?)://([a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`)
		match := re.FindStringSubmatch(url)
		if match == nil {
			logrus.WithField("downloader", "GetBookInfoAndChapterURLs").Error("Error when get domain")
			return book, nil, errors.New("error when get domain")
		}
		for i := range chapter_urls {
			chapter_urls[i] = match[0] + chapter_urls[i]
		}
	}
	return book, chapter_urls, nil
}

func (d *DownloaderImpl) GetBookInfoAndChapterURLs_from_file(file_content string, url string) (Book, []string, error) {
	d.Client = &http.Client{}
	book := Book{}
	var err error
	body := file_content
	book.name, err = d.Get_Book_Name(string(body))
	if err != nil {
		logrus.WithField("downloader", "GetBookInfoAndChapterURLs_from_file").Error("Error when get book name")
		return book, nil, err
	}
	logrus.WithField("downloader", "GetBookInfoAndChapterURLs_from_file").Info("Get book name: " + book.name)
	chapter_urls, prefix, err := d.Get_Chapter_Urls(string(body))
	if err != nil {
		logrus.WithField("downloader", "GetBookInfoAndChapterURLs_from_file").Error("Error when get chapter urls")
		return book, nil, err
	}
	if !prefix {
		re := regexp.MustCompile(`(https?)://([a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`)
		match := re.FindStringSubmatch(url)
		if match == nil {
			logrus.WithField("downloader", "GetBookInfoAndChapterURLs_from_file").Error("Error when get domain")
			return book, nil, errors.New("error when get domain")
		}
		for i := range chapter_urls {
			chapter_urls[i] = match[0] + chapter_urls[i]
		}
	}
	return book, chapter_urls, nil
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
	urls = uniqueStrings(urls)
	if strings.HasPrefix(urls[0], "http") {
		return urls, true, nil
	} else {
		return urls, false, nil
	}
}

func (d *DownloaderImpl) Get_Chapters(urls []string, chapter_url_map map[string]int) ([]Chapter, error) {
	logrus.WithField("component", "Get_Chapters").Info("start getting chapters...")
	var chapters []Chapter
	var wg sync.WaitGroup
	var mut sync.Mutex
	ch := make(chan struct{}, d.thread_num)
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			ch <- struct{}{}
			defer func() {
				<-ch
			}()
			title, content, err := d.Get_Chapter_Content(url)
			if err != nil {
				for i := 0; i < 5; i++ {
					title, content, err = d.Get_Chapter_Content(url)
					if err == nil {
						break
					}
					logrus.WithField("component", "Get_Chapters").Warnf("error getting chapter: %s content, retrying... (%d/5)", url, i+1)
					time.Sleep(time.Second * 2)
				}
				if err != nil {
					logrus.WithField("component", "Get_Chapters").Errorf("failed to get chapter: %s content after 5 retries", url)
					return
				}
				logrus.WithField("component", "Get_Chapters").Errorf("error getting chapter: %s content but retries succeded", url)
				return
			}
			mut.Lock()
			chapters = append(chapters, Chapter{index: chapter_url_map[url], title: title, content: content, url: url})
			mut.Unlock()
		}(url)
		time.Sleep(time.Millisecond * 100)
	}
	wg.Wait()
	logrus.WithField("component", "Get_Chapters").Infof("Got %d chapters", len(chapters))
	sort.Slice(chapters, func(i, j int) bool {
		return chapters[i].index < chapters[j].index
	})
	return chapters, nil
}

func (d *DownloaderImpl) Get_Chapter_Content(url string) (title string, content string, err error) {
	logrus.WithField("component", "Get_Chapter_Content").Infof("start getting chapter: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logrus.WithField("component", "Get_Chapter_Content").Errorf("failed to create request: %s", err)
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.90 Safari/537.36")
	resp, err := d.Client.Do(req)
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
	match := d.title_regex.FindStringSubmatch(string(body))
	if len(match) < 1 {
		logrus.WithField("component", "Get_Chapter_Content").Errorf("error getting title: %s", url)
		logrus.WithField("component", "Get_Chapter_Content").Debugf("body: %s", string(body))
		return "", "", errors.New("error getting title")
	}
	title = match[1]
	contents := d.content_regex.FindAllStringSubmatch(string(body), -1)
	for _, c := range contents {
		content = content + c[1] + "\n"
	}
	logrus.WithField("component", "Get_Chapter_Content").Infof("got chapter content: %s", title)
	return title, content, nil
}