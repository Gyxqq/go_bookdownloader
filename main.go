package godown

import (
	"fmt"
	"runtime"
	"path"
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
	logrus.WithField("component", "main").Info("Starting server...")

	fmt.Print("输入小说主页url:")
	var url string
	fmt.Scanln(&url)
}
