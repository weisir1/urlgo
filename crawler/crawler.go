package crawler

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/weisir1/URLGo/cmd"
	"github.com/weisir1/URLGo/config"
	"github.com/weisir1/URLGo/result"
	"github.com/weisir1/URLGo/util"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func Spider(s *result.Scan) {
	for s.UrlQueue.Len() != 0 {
		dataface := s.UrlQueue.Pop()
		switch dataface.(type) {
		case []string:
			urls := dataface.([]string)

			num, _ := strconv.Atoi(urls[1])
			u, _ := url.QueryUnescape(urls[0])
			if strings.Contains(u, "/js/js/") {
				u = strings.Replace(u, "/js/js/", "/js/", -1)
			}
			if _, loaded := s.Visited.LoadOrStore(u, true); loaded {
				continue
			}
			fmt.Printf("Start  Spider Target  %s, 爬取步数%s\n", u, num)

			isRisk := -1
			for _, v := range config.Risks {
				if strings.Contains(u, v) {
					isRisk = 1
				}
			}
			if isRisk == 1 {
				continue
			}
			//}
			AppendEndUrl(s, u, urls[2]) //添加历史请求列表

			request, err := http.NewRequest("GET", u, nil)
			if err != nil {
				continue
			}

			//request.Header.Set("Accept-Encoding", "gzip") //使用gzip压缩传输数据让访问更快
			request.Header.Set("User-Agent", util.GetUserAgent())
			request.Header.Set("Accept", "*/*")
			//增加header选项
			if cmd.C != "" {
				request.Header.Set("Cookie", cmd.C)
			}
			//加载yaml配置(headers)
			if cmd.I {
				util.SetHeadersConfig(&request.Header)
			}
			response, err := client.Do(request)
			if err != nil {
				continue
			}
			num = num + 1
			defer response.Body.Close()
			result := ""

			if response.Header.Get("Content-Encoding") == "gzip" {

				buffer := bufferPool.Get().(*bytes.Buffer)
				defer bufferPool.Put(buffer)
				buffer.Reset()
				reader, err := gzip.NewReader(response.Body) // gzip解压缩
				if err != nil {
					continue
				}
				defer reader.Close()
				_, err = io.Copy(buffer, reader)
				//con, err := ioutil.ReadAll(reader)
				if err != nil {
					continue
				}
				result = buffer.String()

			} else {

				buffer := bufferPool.Get().(*bytes.Buffer)
				defer bufferPool.Put(buffer)
				buffer.Reset()
				//file, err := os.Create("output.txt")
				//if err != nil {
				//	fmt.Println("Failed to create file:", err)
				//	return
				//}
				//defer file.Close()
				//// 使用 io.TeeReader 将内容同时写入缓冲区和文件
				//teeReader := io.TeeReader(response.Body, &resultBuffer)
				//_, err = io.Copy(file, teeReader)
				//if err != nil {
				//	fmt.Println("Failed to copy response body:", err)
				//	return
				//}
				//
				//buf := make([]byte, 1024*1024) // 每次读取 1MB
				//for {
				//	n, err := response.Body.Read(buf)
				//	if err != nil && err != io.EOF {
				//		fmt.Println("Failed to read response body:", err)
				//		return
				//	}
				//	if n == 0 {
				//		break
				//	}
				//	resultBuffer.Write(buf[:n]) // 将读取的内容写入缓冲区
				//}
				_, err := io.Copy(buffer, response.Body)
				if err != nil {
					fmt.Printf("响应体读取错误-%v", err)
					continue
				}

				response.Body.Close()
				////将缓冲区内容转换为字符串
				result = buffer.String()
				//dataBytes, err := ioutil.ReadAll(response.Body)
				//if err != nil {
				//	return
				//}
				////字节数组 转换成 字符串
				//result = string(dataBytes)
				//result = results.String()
			}
			base1 := urls[2]
			host := regexp.MustCompile("http.*?//([^/]+)").FindAllStringSubmatch(base1, -1)[0][1]
			scheme := regexp.MustCompile("(http.*?)://").FindAllStringSubmatch(base1, -1)[0][1]
			//path := regexp.MustCompile("http.*?//.*?(/.*)").FindAllStringSubmatch(urls[1], -1)[0][1]
			path := response.Request.URL.Path
			//host := response.Request.URL.Host
			//scheme := response.Request.URL.Scheme
			source := scheme + "://" + host + path
			//处理base标签,如果有的站点前台地址后后台接口不一致时,可以进行切换
			re := regexp.MustCompile("base.{1,5}href.{1,5}(http.+?//[^\\s]+?)[\"'‘“]")
			base := re.FindAllStringSubmatch(result, -1)
			if len(base) > 0 {
				base1 = base[0][1]
				urls[2] = base1
				host = regexp.MustCompile("http.*?//([^/]+)").FindAllStringSubmatch(base1, -1)[0][1]
				scheme = regexp.MustCompile("(http.*?)://").FindAllStringSubmatch(base1, -1)[0][1]
				paths := regexp.MustCompile("http.*?//.*?(/.*)").FindAllStringSubmatch(base1, -1)
				if len(paths) > 0 {
					path = paths[0][1]
				} else {
					path = "/"
				}
			}
			//提取js
			jsFind(s, result, base1, host, scheme, path, u, num)
			//提取url,待定可添加参数,对参数进行判定是否进行请求
			urlFind(s, result, base1, host, scheme, path, u, num)

			//提取信息
			infoFind(s, result, base1, source)

		default:
			continue
		}
	}

}

// 打印Validate进度
//func PrintProgress() {
//	config.Mux.Lock()
//	num := len(result.ResultJs) + len(result.ResultUrl)
//	fmt.Printf("\rValidate %.0f%%", float64(config.Progress+1)/float64(num+1)*100)
//	config.Progress++
//	config.Mux.Unlock()
//}
