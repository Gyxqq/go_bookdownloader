package godown

import (
	"regexp"
	"net/http"
)

type Downloader interface {
	Download(url string) (Book, error)
	Get_Book_Name(desc string) (string, error)
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
	Client *http.Client
	name_regex    *regexp.Regexp
	title_regex   *regexp.Regexp
	content_regex *regexp.Regexp
}

func (d *DownloaderImpl) Download(url string) (Book, error) {

}
