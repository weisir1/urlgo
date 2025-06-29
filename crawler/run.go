package crawler

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/gookit/color"
	"github.com/weisir1/URLGo/cmd"
	"github.com/weisir1/URLGo/config"
	"github.com/weisir1/URLGo/mode"
	"github.com/weisir1/URLGo/queue"
	"github.com/weisir1/URLGo/result"
	"github.com/weisir1/URLGo/util"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var client *http.Client

func load() {
	//可以直接指定config.yaml,默认
	if cmd.I {
		config.GetConfig("config.yaml")
	}
	cmd.Parse()
	if cmd.H {
		flag.Usage()
		os.Exit(0)
	}
	if cmd.U == "" && cmd.F == "" && cmd.FF == "" {
		fmt.Println("至少使用 -u -f -ff 指定一个url")
		os.Exit(0)
	}
	u, ok := url.Parse(cmd.U)
	if cmd.U != "" && ok != nil {
		fmt.Println("url格式错误,请填写正确url")
		os.Exit(0)
	}

	cmd.U = u.String()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   time.Second * 30,
			KeepAlive: time.Second * 30,
		}).DialContext,
		MaxIdleConns:          cmd.T / 2,
		MaxIdleConnsPerHost:   cmd.T + 10,
		IdleConnTimeout:       time.Second * 90,
		TLSHandshakeTimeout:   time.Second * 90,
		ExpectContinueTimeout: time.Second * 10,
	}

	if cmd.X != "" {
		tr.DisableKeepAlives = true
		proxyUrl, parseErr := url.Parse(cmd.X)
		if parseErr != nil {
			fmt.Println("代理地址错误: \n" + parseErr.Error())
			os.Exit(1)
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
	}
	if cmd.I {
		util.SetProxyConfig(tr)
	}

	client = &http.Client{
		Timeout:   (time.Duration(cmd.TI) * time.Second),
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("Too many redirects")
			}
			if len(via) > 0 {
				if via[0] != nil && via[0].URL != nil {
					AddRedirect(via[0].URL.String())
				} else {
					AddRedirect(req.URL.String())
				}
			}
			return nil
		},
	}

	result.Initfilecreatename()

}
func LocalFile(filename string) (urls []string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Println("Local file read error:", err)
		color.RGBStyleFromString("237,64,35").Println("[error] the input file is wrong!!!")
		os.Exit(1)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		text = strings.TrimSpace(text)
		if strings.Contains(text, "http") {
			urls = append(urls, text)
		} else {
			urls = append(urls, "https://"+text)
		}
	}
	return
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Run() {

	load()
	if cmd.F != "" {
		// 创建句柄
		Initialization()
		urls := LocalFile(cmd.F)
		i := len(urls)
		//s := NewScan(urls, min(i, cmd.T))
		s := NewScan(urls, cmd.T)
		fmt.Println("加载目标target数量: ", i)
		//r := bufio.NewReader(fi) // 创建 Reader

		StartScan(s)
		Res(s)
		fmt.Println("----------------------------------------")
		return
	}

	Initialization()
	cmd.U = util.GetProtocol(cmd.U)
	s := NewScan([]string{cmd.U}, 1)
	StartScan(s)
	Res(s)
}

func Res(s *result.Scan) {
	if len(s.JsResult) == 0 && len(s.UrlResult) == 0 && len(s.InfoResult) == 0 {
		fmt.Println("未获取到数据")
		return
	}
	//打印还是输出
	if len(cmd.O) > 0 {
		if strings.HasSuffix(cmd.O, ".xlsx") {
			result.OutFilecXlsx(cmd.O, s)
		} else {
			result.OutFilecXlsx("", s)
		}
	} else {
		result.OutFilecXlsx("", s)
	}
	//else {
	//	UrlToRedirect()
	//}
}

func StartScan(s *result.Scan) {

	for i := 0; i <= s.Thread; i++ {
		s.Wg.Add(1)
		go func() {
			defer s.Wg.Done()
			Spider(s)
		}()
	}
	s.Wg.Wait()
	fmt.Println("\nAll Target Spider Complete!!")
	if cmd.S != "" {
		fmt.Println("正在对目标进行状态检测")

		for _, baseurl := range result.Baseurl {
			links := s.UrlResult[baseurl]
			ResultUrl := util.RemoveRepeatElement(links)
			list := s.JsResult[baseurl]
			ResultJs := util.RemoveRepeatElement(list)
			fmt.Printf("\rStart %d Validate...", len(ResultUrl)+len(ResultJs))
			fmt.Printf("\r                    ")
			//验证JS状态
			for i, js := range ResultJs {
				config.Wg.Add(1)
				config.Jsch <- 1
				go JsState(s, js.Url, i, ResultJs[i].Source, baseurl)
			}
			//验证URL状态
			for i, ul := range ResultUrl {
				config.Wg.Add(1)
				config.Urlch <- 1
				go UrlState(s, ul.Url, i, baseurl)
			}
			config.Wg.Wait()

			//time.Sleep(1 * time.Second)
			fmt.Printf("\r                                           ")
			fmt.Printf("\rValidate OK \n\n")

			if cmd.Z != 0 {
				time.Sleep(1 * time.Second)
			}
		}

	}
}

func NewScan(urls []string, thread int) *result.Scan {
	s := &result.Scan{
		UrlQueue: queue.NewQueue(),
		Ch:       make(chan []string, thread),
		Wg:       sync.WaitGroup{},
		Thread:   thread,
		Endurl:   map[string][]string{},
		Pakeris:  map[string]bool{},
		//Output:     output,
		Visited:    sync.Map{},
		JsResult:   make(map[string][]mode.Link),
		UrlResult:  make(map[string][]mode.Link),
		InfoResult: make(map[string][]mode.Info),
	}

	for _, url := range urls {
		s.UrlQueue.Push([]string{url, "0", url})
		result.Baseurl = append(result.Baseurl, url)
	}
	return s
}

func AppendEndUrl(s *result.Scan, url string, baseurl string) {
	config.Lock.Lock()
	defer config.Lock.Unlock()
	for _, eachItem := range s.Endurl[baseurl] {
		if eachItem == url {
			return
		}
	}
	s.Endurl[baseurl] = append(s.Endurl[baseurl], url)
}

func GetEndUrl(s *result.Scan, url string, baseurl string) bool {
	config.Lock.Lock()
	defer config.Lock.Unlock()
	for _, eachItem := range s.Endurl[baseurl] {
		if eachItem == url {
			return true
		}
	}
	return false
}

func AddRedirect(url string) {
	config.Lock.Lock()
	defer config.Lock.Unlock()
	result.Redirect[url] = true
}

//func UrlToRedirect() {
//	for i := range result.ResultJs {
//		if result.ResultJs[i].Status == "302" {
//			result.ResultJs[i].Url = result.ResultJs[i].Url + " -> " + result.ResultJs[i].Redirect
//		}
//	}
//	for i := range result.ResultUrl {
//		if result.ResultUrl[i].Status == "302" {
//			 result.ResultUrl[i].Url = result.ResultUrl[i].Url + " -> " + result.ResultUrl[i].Redirect
//		}
//	}
//
//}

func Initialization() {
	//result.ResultsPacker = mode.Link{}
	//result.ResultJs = []mode.Link{}
	//result.ResultUrl = []mode.Link{}
	//result.Fuzzs = []mode.Link{}
	//result.Infos = []mode.Info{}
	//result.EndUrl = []string{}
	result.Baseurl = []string{}
	result.Domains = []string{}
	//result.Jsinurl = make(map[string]string)
	//result.Jstourl = make(map[string]string)
	//result.Urltourl = make(map[string]string)
	result.Redirect = make(map[string]bool)

}
