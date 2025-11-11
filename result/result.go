package result

import (
	_ "embed"
	"fmt"
	"github.com/tealeg/xlsx"
	"github.com/weisir1/URLGo/cmd"
	"github.com/weisir1/URLGo/mode"
	"github.com/weisir1/URLGo/util"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

//go:embed report.html
var html string
var fileCreatetime string = "202412171452"

var (
	//ResultsPacker mode.Link
	//ResultJs      []mode.Link
	//ResultUrl     []mode.Link
	//Fuzzs []mode.Link
	//Infos   []mode.Info
	Baseurl []string
	//EndUrl  []string
	//Jsinurl       map[string]string
	//Jstourl       map[string]string
	//Urltourl      map[string]string
	Domains  []string
	Redirect map[string]bool
)

type Scan struct {
	UrlQueue chan []string
	//Ch         chan []string
	Wg           sync.WaitGroup
	Thread       int
	ActiveCount  int32
	Done         chan struct{}
	Output       string
	Proxy        string
	Pakeris      map[string]bool
	Endurl       map[string][]string
	JsResult     map[string][]mode.Link
	UrlResult    map[string][]mode.Link
	InfoResult   map[string][]mode.Info
	FingerResult map[string][]mode.Link
	Visited      sync.Map
	ResultMux    sync.Mutex
	PendingURLs  *sync.Map    //  sync.Map 天然安全
	BatchSize    int          // 只读，安全
	BatchTicker  *time.Ticker // 只在一个协程使用
}

// 新增：向队列添加URL的方法（带流控）
func (s *Scan) AddURL(url []string) {
	//退出优先：当两个分支都“就绪”时（比如队列有空位、同时 done 也已关闭），
	// ✅ LoadOrStore 是原子操作（并发安全）
	if _, loaded := s.Visited.LoadOrStore(url[0], struct{}{}); loaded {
		return
	}

	// ✅ Store 是原子操作（并发安全）
	//s.PendingURLs.Store(url, struct{}{})
	//select 随机选一支，不保证一定优先写入队列。
	//start := time.Now()
	select {
	case s.UrlQueue <- url:
	//case <-time.After(2 * time.Second):
	//	log.Printf("添加URL超时: %s", url[0])
	case <-s.Done:
		return
		//	select中是随机收到退出信号不在添加队列
	}
	//elapsed := time.Since(start).Milliseconds()
	//fmt.Printf("添加爬取资源耗时: %d ms\n", elapsed)
}

// ✅ 并发安全：单个协程运行
func (s *Scan) BatchProcessor() {
	defer s.Wg.Done()

	batch := make([]string, 0, s.BatchSize)

	for {
		select {
		case <-s.BatchTicker.C:
			// ✅ Range + Delete 是并发安全的
			s.PendingURLs.Range(func(key, value interface{}) bool {
				if len(batch) >= s.BatchSize {
					return false
				}

				url := key.(string)
				batch = append(batch, url)

				// ✅ Delete 是原子操作
				s.PendingURLs.Delete(url)
				return true
			})

			// ✅ 发送到 channel（并发安全）
			if len(batch) > 0 {
				select {
				case s.UrlQueue <- batch:
				case <-time.After(1 * time.Second):
					fmt.Printf("⚠️ 队列阻塞\n")
				case <-s.Done:
					return
				}
				batch = batch[:0]
			}

		case <-s.Done:
			return
		}
	}
}

// ✅ 安全地添加JS结果
func (s *Scan) AddJsResult(baseurl string, link mode.Link) {
	s.ResultMux.Lock()
	defer s.ResultMux.Unlock()
	s.JsResult[baseurl] = append(s.JsResult[baseurl], link)
}

// ✅ 安全地添加URL结果
func (s *Scan) AddUrlResult(baseurl string, link mode.Link) {
	s.ResultMux.Lock()
	defer s.ResultMux.Unlock()
	s.UrlResult[baseurl] = append(s.UrlResult[baseurl], link)
}

// ✅ 安全地添加Info结果
func (s *Scan) AddInfoResult(baseurl string, info mode.Info) {
	s.ResultMux.Lock()
	defer s.ResultMux.Unlock()
	s.InfoResult[baseurl] = append(s.InfoResult[baseurl], info)
}

// 新增：获取活跃任务数
func (s *Scan) GetActiveCount() int32 {
	return atomic.LoadInt32(&s.ActiveCount)
}

// 统计访问的URL数
func (s *Scan) CountVisited() int {
	count := 0
	s.Visited.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// 新增：通知所有协程退出
func (s *Scan) Stop() {
	select {
	case <-s.Done:
		// 已经关闭
	default:
		close(s.Done)
	}
}

func Initfilecreatename() {
	now := time.Now()
	// 定义所需的时间格式
	const layout = "200601021504"
	// 格式化时间
	fileCreatetime = now.Format(layout)
}

func writeRow(sheet *xlsx.Sheet, rowData []string) {
	row := sheet.AddRow()
	for _, cellData := range rowData {
		cell := row.AddCell()
		cell.SetString(cellData)
	}
}

// OutFileXlsx 导出结果到Excel文件
func OutFilecXlsx(out string, s *Scan) error {
	fileName := getFileName(out)

	file, err := createOrOpenFile(fileName)
	if err != nil {
		return err
	}

	// 初始化Sheet
	if err := initializeSheets(file); err != nil {
		return err
	}

	// 处理每个BaseURL
	for _, url := range Baseurl {
		processBaseURL(file, url, s)
	}

	// 保存文件
	if err := file.Save(fileName); err != nil {
		return fmt.Errorf("保存文件失败: %v", err)
	}

	fmt.Println("结果已导出 --> ", fileName)
	return nil
}

// getFileName 获取输出文件名
func getFileName(out string) string {
	Initfilecreatename()
	if out != "" {
		if strings.HasPrefix(out, ".") {
			return out
		}
		return "./" + out
	}
	return "./" + fileCreatetime + "_result.xlsx"
}

// createOrOpenFile 创建或打开Excel文件
func createOrOpenFile(fileName string) (*xlsx.File, error) {
	file, err := xlsx.OpenFile(fileName)
	if err != nil {
		log.Println("文件不存在，创建新文件...")
		file = xlsx.NewFile()
	}
	return file, nil
}

// initializeSheets 初始化Sheet的表头
func initializeSheets(file *xlsx.File) error {
	urlSheet, err := file.AddSheet("url")
	if err != nil {
		return err
	}

	jsSheet, err := file.AddSheet("js")
	if err != nil {
		return err
	}

	infoSheet, err := file.AddSheet("info")
	if err != nil {
		return err
	}

	fingerSheet, err := file.AddSheet("finger")
	if err != nil {
		return err
	}
	// 写入表头
	if cmd.S == "" {
		writeRow(infoSheet, []string{"info", "", "", "", "Source"})
		writeRow(jsSheet, []string{"jsurl", "Source"})
		writeRow(urlSheet, []string{"url", "Source"})
		writeRow(fingerSheet, []string{"finger", "match", "Source"})
	} else {
		writeRow(infoSheet, []string{"info", "", "", "", "Source"})
		writeRow(jsSheet, []string{"jsurl", "Status", "Size", "Title", "Redirect", "Source"})
		writeRow(urlSheet, []string{"url", "Status", "Size", "Title", "Redirect", "Source"})
		writeRow(fingerSheet, []string{"finger", "match", "Source"})
	}

	return nil
}

// processBaseURL 处理单个BaseURL的所有结果
func processBaseURL(file *xlsx.File, baseURL string, s *Scan) {
	urlSheet, _ := file.Sheet["url"]
	jsSheet, _ := file.Sheet["js"]
	infoSheet, _ := file.Sheet["info"]
	fingerSheet, _ := file.Sheet["finger"]

	// 获取结果
	urlRes := safeGetLinks(s.UrlResult[baseURL])
	jsRes := safeGetLinks(s.JsResult[baseURL])
	infoRes := safeGetInfos(s.InfoResult[baseURL])
	fingerRes := safeGetLinks(s.FingerResult[baseURL])

	// 处理URL和JS结果
	if len(urlRes) > 0 || len(jsRes) > 0 {
		writeURLAndJSResults(jsSheet, urlSheet, fingerSheet, baseURL, jsRes, urlRes, fingerRes)
	}

	// 处理Info结果
	if len(infoRes) > 0 {
		writeInfoResults(infoSheet, baseURL, infoRes)
	}
}

// writeURLAndJSResults 写入URL和JS结果
func writeURLAndJSResults(jsSheet, urlSheet, fingersheet *xlsx.Sheet, baseURL string,
	jsRes, urlRes, fingerRes []mode.Link) {

	// 去重
	jsRes = util.RemoveDuplicatesLink(jsRes)
	urlRes = util.RemoveDuplicatesLink(urlRes)
	fingerRes = util.RemoveDuplicatesLinkFinger(fingerRes)

	// 排序（如果需要）
	if cmd.S != "" {
		jsRes = util.SelectSort(jsRes)
		urlRes = util.SelectSort(urlRes)
	}

	// 分类处理
	jsHost, _ := util.UrlDispose(jsRes)
	urlHost, urlOther := util.UrlDispose(urlRes)

	// 更新域名列表
	allLinks := util.MergeArray(jsRes, urlRes)
	domains := util.GetDomains(allLinks)

	// 写入JS结果
	writeJSSection(jsSheet, baseURL, jsHost)

	//写入finger结果
	writeFingerSection(fingersheet, baseURL, fingerRes)

	// 写入URL结果
	writeURLSection(urlSheet, baseURL, urlHost, urlOther, domains)
}

// writeJSSection 写入JS部分
func writeJSSection(sheet *xlsx.Sheet, baseURL string, jsHost []mode.Link) {
	writeRow(sheet, []string{"", ""})
	writeRow(sheet, []string{"", ""})
	writeRow(sheet, []string{"baseurl:", baseURL})
	writeRow(sheet, []string{strconv.Itoa(len(jsHost)) + " JS"})

	for _, j := range jsHost {
		if cmd.S != "" {
			writeRow(sheet, []string{j.Url, j.Status, j.Size, "", j.Redirect, j.Source})
		} else {
			writeRow(sheet, []string{j.Url, j.Source})
		}
	}
}

func writeFingerSection(sheet *xlsx.Sheet, baseURL string, Finger []mode.Link) {
	writeRow(sheet, []string{"", ""})
	writeRow(sheet, []string{"", ""})
	writeRow(sheet, []string{"baseurl:", baseURL})
	writeRow(sheet, []string{strconv.Itoa(len(Finger)) + " finger "})

	for _, j := range Finger {
		writeRow(sheet, []string{j.Finger, j.MatchesN, j.Source})
	}
}

// writeURLSection 写入URL部分
func writeURLSection(sheet *xlsx.Sheet, baseURL string,
	urlHost, urlOther []mode.Link, domains []string) {

	// 写入同域URL
	writeRow(sheet, []string{"", ""})
	writeRow(sheet, []string{"", ""})
	writeRow(sheet, []string{"baseurl:", baseURL})
	writeRow(sheet, []string{strconv.Itoa(len(urlHost)) + " URL "})

	for _, u := range urlHost {
		writeURLRow(sheet, u)
	}

	// 写入跨域URL
	writeRow(sheet, []string{""})
	writeRow(sheet, []string{""})
	writeRow(sheet, []string{strconv.Itoa(len(urlOther)) + " Other URL "})

	for _, u := range urlOther {
		writeURLRow(sheet, u)
	}

	// 写入域名列表
	writeRow(sheet, []string{""})
	writeRow(sheet, []string{strconv.Itoa(len(domains)) + " Domain"})
	for _, domain := range domains {
		writeRow(sheet, []string{domain})
	}
}

// writeURLRow 写入单个URL行
func writeURLRow(sheet *xlsx.Sheet, link mode.Link) {
	if cmd.S != "" {
		writeRow(sheet, []string{link.Url, link.Status, link.Size, link.Title, link.Redirect, link.Source})
	} else {
		writeRow(sheet, []string{link.Url, link.Source})
	}
}

// writeInfoResults 写入信息结果
func writeInfoResults(sheet *xlsx.Sheet, baseURL string, infos []mode.Info) {
	writeRow(sheet, []string{"", ""})
	writeRow(sheet, []string{"", ""})
	writeRow(sheet, []string{"baseurl:", baseURL})

	// 写入各类信息（统一处理，避免重复代码）
	writeInfoCategory(sheet, "Phone", infos,
		func(info mode.Info) []string { return info.Phone })

	writeInfoCategory(sheet, "Email", infos,
		func(info mode.Info) []string { return info.Email })

	writeInfoCategory(sheet, "IDcard", infos,
		func(info mode.Info) []string { return info.IDcard })

	writeInfoCategory(sheet, "JWT", infos,
		func(info mode.Info) []string { return info.JWT })

	writeInfoCategory(sheet, "Other", infos,
		func(info mode.Info) []string { return info.Other })
}

// writeInfoCategory 写入单个信息类别（核心优化：消除重复代码）
func writeInfoCategory(sheet *xlsx.Sheet, title string, infos []mode.Info,
	extractor func(mode.Info) []string) {

	writeRow(sheet, []string{""})
	writeRow(sheet, []string{title})

	// 使用Map而不是字符串拼接去重（性能优化）
	seen := make(map[string]bool)

	for _, info := range infos {
		for _, item := range extractor(info) {
			if item != "" && !seen[item] {
				seen[item] = true
				writeRow(sheet, []string{item, "", "", "", info.Source})
			}
		}
	}
}

// safeGetLinks 安全获取Link切片
func safeGetLinks(links []mode.Link) []mode.Link {
	if links == nil {
		return []mode.Link{}
	}
	return links
}

// safeGetInfos 安全获取Info切片
func safeGetInfos(infos []mode.Info) []mode.Info {
	if infos == nil {
		return []mode.Info{}
	}
	return infos
}
