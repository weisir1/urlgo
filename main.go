package main

import (
	"github.com/pingc0y/URLFinder/crawler"
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
	//cmd.T = 300
	//cmd.X = "http://127.0.0.1:8080"
	//cmd.M = 2
	//cmd.Parse()
	crawler.Run()
}
