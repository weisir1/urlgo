package crawler

import (
	"compress/gzip"
	"fmt"
	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/config"
	"github.com/pingc0y/URLFinder/result"
	"github.com/pingc0y/URLFinder/util"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// 蜘蛛抓取页面内容
func Spider(s *result.Scan) {
	for s.UrlQueue.Len() != 0 {
		dataface := s.UrlQueue.Pop()
		switch dataface.(type) {
		case []string:
			urls := dataface.([]string)

			num, _ := strconv.Atoi(urls[1])
			u, _ := url.QueryUnescape(urls[0])
			if GetEndUrl(s, u, urls[2]) {
				continue
			}
			fmt.Printf("\rStart  Spider Target %s", u)

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

			request.Header.Set("Accept-Encoding", "gzip") //使用gzip压缩传输数据让访问更快
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
				return
			}

			defer response.Body.Close()
			result := ""
			//var results strings.Builder
			//buffer := make([]byte, 1024*10) // 1KB 的缓冲区
			////解压
			if response.Header.Get("Content-Encoding") == "gzip" {
				reader, err := gzip.NewReader(response.Body) // gzip解压缩
				if err != nil {
					return
				}
				defer reader.Close()
				//for {
				//	n, err := reader.Read(buffer)
				//	if err != nil && err != io.EOF {
				//		log.Fatal(err)
				//	}
				//	if n == 0 {
				//		break
				//	}
				//	results.Write(buffer[:n])
				//	// 处理读取到的数据
				//}
				if err != nil {
					return
				}
				defer reader.Close()
				con, err := io.ReadAll(reader)
				if err != nil {
					return
				}
				result = string(con)
			} else {
				//for {
				//	n, err := response.Body.Read(buffer)
				//	if err != nil && err != io.EOF {
				//		log.Fatal(err)
				//	}
				//	if n == 0 {
				//		break
				//	}
				//	results.Write(buffer[:n])
				//	// 处理读取到的数据
				//}
				dataBytes, err := io.ReadAll(response.Body)
				if err != nil {
					return
				}
				//字节数组 转换成 字符串
				result = string(dataBytes)
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
