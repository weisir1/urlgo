package main

import (
	"github.com/weisir1/URLGo/crawler"
	"io"
	"log"
)

func main() {
	log.SetOutput(io.Discard)
	//util.GetUpdate()
	//config.JsSteps = 1
	//config.UrlSteps = 1
	//cmd.M = 2
	//cmd.F = "url.txt"
	//cmd.S = "all"
	//cmd.M = 2
	//cmd.Parse()
	crawler.Run()
}
