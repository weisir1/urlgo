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

func processURL(s *result.Scan, urls []string) {

	if len(urls) < 3 {
		return
	}
	num, err := strconv.Atoi(urls[1])
	if err != nil {
		return
	}
	u, _ := url.QueryUnescape(urls[0])

	if _, loaded := s.Visited.LoadOrStore(u, true); loaded {
		return
	}

	fmt.Printf("正在请求url: %v\n", u)
	for _, risk := range config.Risks {
		if strings.Contains(u, risk) {
			return
		}
	}

	//AppendEndUrl(s, u, urls[2])
	//start := time.Now()
	// 创建请求
	request, err := http.NewRequest("GET", u, nil)
	if err != nil {

		return
	}

	if cmd.I {
		util.SetHeadersConfig(&request.Header)
	}

	request.Header.Set("User-Agent", util.GetUserAgent())
	request.Header.Set("Accept", "*/*")

	if cmd.He != "" {
		parts := strings.SplitN(cmd.He, ":", 2)
		if len(parts) == 2 {
			request.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	//增加header选项
	if cmd.C != "" {
		request.Header.Set("Cookie", cmd.C)
	}

	response, err := client.Do(request)
	if err != nil {
		return
	}

	defer response.Body.Close()
	//读取响应体
	result, err := readResponseBody(response)
	//elapsed := time.Since(start).Milliseconds()
	//fmt.Printf("请求地址耗时: %d ms 请求地址: %v\n", elapsed, u)
	if err != nil {
		fmt.Printf("读取响应体失败: %v 地址%v\n", err, u)
		return
	}

	// 解析URL信息
	base1 := urls[2]
	host, scheme, path := parseURLInfo(base1, response)
	source := scheme + "://" + host + path

	if baseURL := extractBaseTag(result); baseURL != "" {
		base1 = baseURL
		host, scheme, path = parseURLInfo(base1, response)
	}

	num++
	jsFind(s, result, base1, host, scheme, path, u, num)
	urlFind(s, result, base1, host, scheme, path, u, num)
	infoFind(s, result, base1, source)

}

func readResponseBody(response *http.Response) (string, error) {
	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)
	buffer.Reset()

	var reader io.Reader = response.Body

	if response.Header.Get("Content-Encoding") == "gzip" {
		gr, err := gzip.NewReader(response.Body)
		if err != nil {
			fmt.Errorf("读取响应体失败 (Content-Length: %d, 已读: %d): %w")
			return "", err
		}
		defer gr.Close()
		reader = gr
	}

	_, err := io.Copy(buffer, reader)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func parseURLInfo(baseURL string, response *http.Response) (host, scheme, path string) {
	if matches := regexp.MustCompile(`http.*?//([^/]+)`).FindStringSubmatch(baseURL); len(matches) > 1 {
		host = matches[1]
	}
	if matches := regexp.MustCompile(`(http.*?)://`).FindStringSubmatch(baseURL); len(matches) > 1 {
		scheme = matches[1]
	}
	path = response.Request.URL.Path
	return
}

func extractBaseTag(content string) string {
	re := regexp.MustCompile(`base.{1,5}href.{1,5}(http.+?//[^\s]+?)["'']`)
	if matches := re.FindStringSubmatch(content); len(matches) > 1 {
		return matches[1]
	}
	return ""
}
