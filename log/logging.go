package log

import (
	"fmt"
	"golang.org/x/exp/slog"
	"log"
	"os"
)

var messageOld = ""

func Debugf(format string, v ...interface{}) {
	//logLevel := envUtils.GetString(consts.PAAS_Log_Level)
	//if len(logLevel) > 0 && logLevel == "debug" {
	msg := fmt.Sprintf(format, v...)
	slog.Debug(msg)
	//}
}

func Infof(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	slog.Info(msg)
}

func Warnf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	slog.Warn(msg)
}

func Errorf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	slog.Error(msg)
}

func Fatalf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	slog.Error(msg)
	os.Exit(1)
}

func Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func Println(v ...interface{}) {
	log.Println(v...)
}

type Message struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Content string `json:"content"`
	} `json:"markdown"`
}
