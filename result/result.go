package result

import (
	"bufio"
	_ "embed"
	"fmt"
	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/mode"
	"github.com/pingc0y/URLFinder/queue"
	"github.com/pingc0y/URLFinder/util"
	"github.com/tealeg/xlsx"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed report.html
var html string
var FileCreatetime string = "202412171452"

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
	WriteCh  chan string
	Redirect map[string]bool
	fileMap  = make(map[string]*os.File)
)

type ResultProcessor struct {
	baseURLs  []string
	fileMap   map[string]*os.File
	fileMutex sync.Mutex
	tempFile  string
	resultDir string
}
type Scan struct {
	UrlQueue   *queue.Queue
	Ch         chan []string
	Wg         sync.WaitGroup
	Thread     int
	Output     string
	Proxy      string
	Pakeris    map[string]bool
	Endurl     map[string][]string
	Visited    sync.Map
	JsResult   map[string][]mode.Link
	UrlResult  map[string][]mode.Link
	InfoResult map[string][]mode.Info
}

func Initfilecreatename() {
	now := time.Now()
	// 定义所需的时间格式
	const layout = "200601021504"
	// 格式化时间
	FileCreatetime = now.Format(layout)
}

func writeRow(sheet *xlsx.Sheet, rowData []string) {
	row := sheet.AddRow()
	for _, cellData := range rowData {
		cell := row.AddCell()
		cell.SetString(cellData)
	}
}

func CleanResultDir(dirname string) error {
	// 删除现有目录（如果存在）
	if _, err := os.Stat(dirname); !os.IsNotExist(err) {
		if err := os.RemoveAll(dirname); err != nil {
			log.Printf("旧目录清理失败: %v\n", err)
			return err
		}
		log.Printf("已清理旧目录: %s", dirname)
	}

	// 创建全新目录
	return os.MkdirAll(dirname, 0755)
}

func initFile() {
	CleanResultDir("results")
	for _, url := range Baseurl {
		safeName := sanitizeFileName(url)
		f, err := os.OpenFile(
			filepath.Join("results", safeName+".txt"),
			os.O_WRONLY|os.O_CREATE|os.O_APPEND,
			0644,
		)
		if err != nil {
			log.Printf("%s创建失败---%v", safeName, err)
		}
		fileMap[safeName] = f
	}

}
func ClassifyResults() {
	initFile()
	defer closeAll()

	// 读取临时文件
	file, err := os.Open("tmp.txt")
	defer file.Close()

	if err != nil {
		log.Printf("读取临时文件失败: %v", err)
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) < 1 {
			continue
		}

		baseURL := sanitizeFileName(parts[0])
		if f, exists := fileMap[baseURL]; exists {
			if _, err := f.WriteString(line + "\n"); err != nil {
				log.Printf("写入%s失败: %v", baseURL, err)
			}
		} else {
			log.Printf("发现未预创建的URL: %s", baseURL)
		}
	}

	// 清理临时文件
	//if err := os.Remove("tmp.txt"); err != nil {
	//	log.Printf("删除临时文件失败: %v", err)
	//}
}
func closeAll() {
	for _, f := range fileMap {
		f.Close()
	}
}
func sanitizeFileName(name string) string {
	return strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(name)
}
func OutFilecXlsx(out string, s *Scan) {

	var fileName string
	//var filejwtName string
	if out != "" {
		if strings.HasPrefix(out, ".") {
			fileName = out
		} else {
			fileName = "./" + out
		}
	} else {
		fileName = "./" + FileCreatetime + "_result.xlsx"
	}

	file, err := xlsx.OpenFile(fileName)
	if err != nil {
		fmt.Println("File does not exist, creating a new one...")
		file = xlsx.NewFile()
	}
	urlsheet, err := file.AddSheet("url")
	jssheet, err := file.AddSheet("js")
	infosheet, err := file.AddSheet("info")

	if err != nil {
		panic(err)
	}
	if cmd.S == "" {
		writeRow(infosheet, []string{"info", "", "", "", "Source"})
		writeRow(jssheet, []string{"jsurl", "Source"})
		writeRow(urlsheet, []string{"url", "Source"})

	} else {
		writeRow(infosheet, []string{"info", "", "", "", "Source"})
		writeRow(jssheet, []string{"jsurl", "Status", "Size", "Title", "Redirect", "Source"})
		writeRow(urlsheet, []string{"url", "Status", "Size", "Title", "Redirect", "Source"})
	}
	//saveInterval := 100
	for _, url := range Baseurl {
		urlres := s.UrlResult[url]
		jsres := s.JsResult[url]
		infores := s.InfoResult[url]
		if urlres == nil {
			urlres = []mode.Link{}
		}
		if jsres == nil {
			jsres = []mode.Link{}
		}
		if infores == nil {
			infores = []mode.Info{}
		}
		urlres = util.RemoveDuplicatesLink(urlres) // 去重
		jsres = util.RemoveDuplicatesLink(jsres)   // 去重

		if cmd.S != "" {
			urlres = util.SelectSort(urlres)
			jsres = util.SelectSort(jsres)
		}
		ResultJsHost, _ := util.UrlDispose(jsres)
		ResultUrlHost, ResultUrlOther := util.UrlDispose(urlres)
		Domains = util.GetDomains(util.MergeArray(jsres, urlres))
		writeRow(jssheet, []string{"", ""})
		writeRow(jssheet, []string{"", ""})
		writeRow(jssheet, []string{"baseurl:   " + url})
		writeRow(jssheet, []string{strconv.Itoa(len(ResultJsHost)) + " JS to " + util.GetHost(cmd.U)})

		for _, j := range ResultJsHost {
			if cmd.S != "" {
				writeRow(jssheet, []string{j.Url, j.Status, j.Size, "", j.Redirect, j.Source})
			} else {
				writeRow(jssheet, []string{j.Url, j.Source})
			}
		}

		writeRow(urlsheet, []string{"", ""})
		writeRow(urlsheet, []string{"", ""})
		writeRow(urlsheet, []string{"baseurl:   " + url})
		writeRow(urlsheet, []string{strconv.Itoa(len(ResultUrlHost)) + " URL to " + util.GetHost(cmd.U)})

		for _, u := range ResultUrlHost {
			if cmd.S != "" {
				writeRow(urlsheet, []string{u.Url, u.Status, u.Size, u.Title, u.Redirect, u.Source})
			} else {
				writeRow(urlsheet, []string{u.Url, u.Source})
			}
		}

		writeRow(urlsheet, []string{""})
		writeRow(urlsheet, []string{""})
		writeRow(urlsheet, []string{strconv.Itoa(len(ResultUrlOther)) + " Other URL to " + util.GetHost(cmd.U)})

		for _, u := range ResultUrlOther {
			if cmd.S != "" {
				writeRow(urlsheet, []string{u.Url, u.Status, u.Size, u.Title, u.Redirect, u.Source})
			} else {
				writeRow(urlsheet, []string{u.Url, u.Source})
			}
		}
		//
		writeRow(urlsheet, []string{""})
		writeRow(urlsheet, []string{strconv.Itoa(len(Domains)) + " Domain"})
		for _, u := range Domains {
			writeRow(urlsheet, []string{u})
		}

		writeRow(infosheet, []string{"", ""})
		writeRow(infosheet, []string{"", ""})
		writeRow(infosheet, []string{"BaseUrl:     " + url})
		writeRow(infosheet, []string{"Phone"})
		for i := range infores {
			for i2 := range infores[i].Phone {
				writeRow(infosheet, []string{infores[i].Phone[i2], "", "", "", infores[i].Source})
			}
		}
		writeRow(infosheet, []string{""})
		writeRow(infosheet, []string{"Email"})
		for i := range infores {
			for i2 := range infores[i].Email {
				writeRow(infosheet, []string{infores[i].Email[i2], "", "", "", infores[i].Source})
			}
		}
		writeRow(infosheet, []string{""})
		writeRow(infosheet, []string{"IDcard"})
		for i := range infores {
			for i2 := range infores[i].IDcard {
				writeRow(infosheet, []string{infores[i].IDcard[i2], "", "", "", infores[i].Source})
			}
		}
		writeRow(infosheet, []string{""})
		writeRow(infosheet, []string{"JWT"})
		for i := range infores {
			for i2 := range infores[i].JWT {
				writeRow(infosheet, []string{infores[i].JWT[i2], "", "", "", infores[i].Source})
			}
		}
		writeRow(infosheet, []string{""})
		writeRow(infosheet, []string{"Other"})
		tmps := ""
		for i := range infores {
			for i2 := range infores[i].Other {
				if strings.Contains(tmps, infores[i].Other[i2]) {
					continue
				}
				tmps += infores[i].Other[i2]
				writeRow(infosheet, []string{infores[i].Other[i2], "", "", "", infores[i].Source})
			}
		}
		//if i > 0 && i%saveInterval == 0 {
		//	err := file.Save(fileName)
		//	if err != nil {
		//		log.Fatalf("Failed to save file: %v", err)
		//	}
		//}
	}
	err = file.Save(fileName)
	if err != nil {
		log.Fatalf("Failed to save file: %v", err)
	}
	fmt.Println(" out to --> ", fileName)
	return
}
