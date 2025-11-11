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
	"github.com/weisir1/URLGo/result"
	"github.com/weisir1/URLGo/util"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var client *http.Client

func load() {
	//可以直接指定config.yaml,默认
	if cmd.I {
		config.GetConfig("conf/config.yaml")
	}
	config.GetFingerConfig()
	//cmd.T = 100
	cmd.Parse()
	config.Init(cmd.T)
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
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			//ClientSessionCache: tls.NewLRUClientSessionCache(2048)
		}, //缓存2048域名的握手环节
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   time.Second * 30,
			KeepAlive: time.Second * 100,
		}).DialContext,
		MaxIdleConns:        100, //总空闲
		MaxIdleConnsPerHost: 10,  //单最大空闲
		MaxConnsPerHost:     50,  //单主机最大连接数(空闲与活跃)

		IdleConnTimeout: time.Second * 25,
		//TLSHandshakeTimeout:   time.Second * 10,
		ExpectContinueTimeout: time.Second * 1,
		//ResponseHeaderTimeout: time.Duration(cmd.TI) * time.Second,
		DisableCompression: false, //设置请求响应压缩
	}

	if cmd.X != "" {
		//tr.DisableKeepAlives = true
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

}

func LocalFile(filename string) (urls []string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Println("Local file read error:", err)
		color.RGBStyleFromString("237,64,35").Println("[error] the input file is wrong!!!")
		os.Exit(1)
	}
	defer file.Close()
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
func max(a, b int) int {
	if a > b {
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
	s := NewScan([]string{cmd.U}, cmd.T)
	StartScan(s)
	Res(s)

}

func Res(s *result.Scan) {
	if len(s.JsResult) == 0 && len(s.UrlResult) == 0 && len(s.InfoResult) == 0 && len(s.FingerResult) == 0 {
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
	fmt.Printf("启动 %d 个爬虫协程，队列缓冲大小：%d\n", s.Thread, cap(s.UrlQueue))
	// ✅ 设置信号处理（Ctrl+C）
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	// ✅ 启动信号监听协程
	go func() {
		<-sigChan
		fmt.Println("\n\n 收到退出信号（Ctrl+C），正在保存结果...")
		s.Stop() // 通知所有协程退出
	}()

	for i := 0; i < s.Thread; i++ {
		s.Wg.Add(1)
		go func(id int) {
			defer s.Wg.Done()
			Spider(s, id)
		}(i)
	}
	go monitorProgress(s)
	s.Wg.Wait()
	// ✅ 保存结果
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
	bufferSize := thread * 1000 //
	//if bufferSize > 5000 {
	//	bufferSize = 5000 //设置上限
	//}
	if bufferSize < 3000 {
		bufferSize = 3000 // 最小100
	}
	s := &result.Scan{
		UrlQueue: make(chan []string, bufferSize),
		Done:     make(chan struct{}),
		//Ch:       make(chan []string, thread),
		Wg:      sync.WaitGroup{},
		Thread:  thread,
		Endurl:  map[string][]string{},
		Pakeris: map[string]bool{},
		//Output:     output,
		Visited:     sync.Map{},
		PendingURLs: &sync.Map{},
		BatchSize:   100,
		BatchTicker: time.NewTicker(200 * time.Millisecond),

		JsResult:     make(map[string][]mode.Link),
		UrlResult:    make(map[string][]mode.Link),
		InfoResult:   make(map[string][]mode.Info),
		FingerResult: make(map[string][]mode.Link),
	}

	for _, url := range urls {
		s.AddURL([]string{url, "0", url})
		//s.UrlQueue <- []string{url, "0", url}
		result.Baseurl = append(result.Baseurl, url)
	}

	// ✅ 启动单个批量处理协程
	//s.Wg.Add(1)
	//go s.BatchProcessor()
	return s
}

func Spider(s *result.Scan, id int) {
	/*	ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		idleTime := time.Duration(0)
		maxIdleTime := 5 * time.Second // 5秒内没有新任务则退出
	*/
	for {

		select {
		case urls := <-s.UrlQueue:

			//idleTime = 0 // 重置空闲时间

			// 增加活跃计数
			atomic.AddInt32(&s.ActiveCount, 1)

			//  处理URL（同步执行）
			func() {
				defer atomic.AddInt32(&s.ActiveCount, -1)
				processURL(s, urls)
			}()

		/*case <-ticker.C:
		// 定期检查是否应该退出
		queueLen := len(s.UrlQueue)
		active := atomic.LoadInt32(&s.ActiveCount)

		if queueLen == 0 && active == 0 {
			idleTime += 200 * time.Millisecond
			if idleTime >= maxIdleTime {
				// log.Printf("协程 %d: 空闲超过 %v，退出\n", id, maxIdleTime)
				return
			}
		} else {
			idleTime = 0
		}
		*/
		case <-s.Done:
			//  收到退出信号
			a := 1
			log.Printf("协程 %d: 收到退出信号\n", id, a)
			return
		}
	}
}

// 监控进度
func monitorProgress(s *result.Scan) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			queueLen := len(s.UrlQueue)
			active := s.GetActiveCount()
			visited := countVisited(&s.Visited)

			fmt.Printf("\r[监控] 队列: %d | 活跃: %d | 已访问: %d\n",
				queueLen, active, visited)

			if queueLen == 0 && active == 0 {
				//队列为空,活跃为空 通知关闭
				close(s.Done)
				return
			}

		case <-s.Done:
			return
		}
	}
}

// 统计访问数
func countVisited(m *sync.Map) int {
	count := 0
	m.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
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

//
//func Init() {
//	// 初始化channel
//
//	// 启动写入goroutine
//	go func() {
//
//		if err := createResultsDir(); err != nil {
//			log.Printf("创建results目录失败：%v", err)
//		}
//		// 打开文件
//		f, err := os.Create("tmp1.txt")
//		if err != nil {
//			fmt.Printf("写入临时文件错误-----------------%v", err)
//		}
//		defer f.Close()
//
//		for line := range result.WriteCh {
//			_, err := f.WriteString(line)
//			if err != nil {
//				log.Println("写入文件错误：", err)
//			}
//		}
//	}()
//}
//func createResultsDir() error {
//	return os.MkdirAll("results", os.ModePerm)
//}

//result.WriteCh <- fmt.Sprintf("----Baseurl: %s, ----Source: %s, ----jsPath: %s\n", baseurl, source, js)
