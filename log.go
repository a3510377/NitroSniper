package main

import (
	"time"

	"github.com/gookit/color"
)

func getTime() string {
	return time.Now().Format("15:04:05")
}

func withTimeFail(msg string) {
	color.Println("<magenta>" + getTime() + "</> <red>" + msg + "</>")
}

func withTimeLog(msg string) {
	color.Println("<magenta>" + getTime() + " </>" + msg)
}
