package crawler

import (
	"encoding/json"
	"fmt"
	"github.com/weisir1/URLGo/cmd"
	"github.com/weisir1/URLGo/config"
	"github.com/weisir1/URLGo/mode"
	"github.com/weisir1/URLGo/result"
	"regexp"
	"strconv"
	"strings"
)

func getPrejspath(data string) string {
	var Prematch []string
	Preregex := regexp.MustCompile(`"([^"]+)"`)
	if len(data) > 50 {
		Prematch = Preregex.FindAllString(data[:50], -1)
	}

	//fmt.Println(Prematch)
	for _, prematch := range Prematch {
		prematch = strings.ReplaceAll(prematch, "\"", "")
		if strings.HasSuffix(prematch, "/") {
			return prematch
		}
	}
	return ""
}
func Jsonppp(par string) map[string]string {
	re := regexp.MustCompile(`([{,])\s*([^",{\s][^,{:]*?)\s*:`)
	fixedInput := re.ReplaceAllString(par, `$1"$2":`)
	var data map[string]string
	err := json.Unmarshal([]byte(fixedInput), &data)
	if err != nil {
		return map[string]string{}
	}

	return data
}

func IsWebpackechunkjs(data string) string {

	re := regexp.MustCompile(`\w\.p\s*\+\s*\"([\s\S]*?)\.js"`)
	matchs := re.FindAllStringSubmatch(data, -1)

	if len(matchs) == 0 {
		return ""
	}

	WebpackArray := []string{}
	for _, m := range matchs {
		WebpackArray = append(WebpackArray, m[0])
	}

	return strings.Join(WebpackArray, "")
}
func WebpackJsTiQu(data string) []string {
	if data == "" {
		return nil
	}
	//提取前缀如static/js/
	prejspath := getPrejspath(data)

	regex := regexp.MustCompile(`(\{[^\}]+\})`)
	regexJsSuff := regexp.MustCompile(`\+\s*?"(.{0,10}\.js)"`)
	// FindAllStringSubmatch 返回所有匹配的字符串以及捕获的子组
	// 处理数字键： 将数字键用引号包裹
	reNumberKeys := regexp.MustCompile(`(\d+):`)
	data = reNumberKeys.ReplaceAllString(data, `"$1":`)

	match := regex.FindAllString(data, -1)
	matchSuff := regexJsSuff.FindAllStringSubmatch(data, -1)
	if matchSuff == nil {
		return nil
	}
	var result1 []string
	var result map[string]string
	for _, v := range match {
		err := json.Unmarshal([]byte(v), &result)
		if err != nil {
			parse := Jsonppp(v)
			if len(parse) > 0 {
				result = parse
				break
			}
			continue
		}
		break
	}
	if len(result) == 0 {
		return nil
	}

	if prejspath == "" {
		for k, v := range result {
			result1 = append(result1, k+"."+v+matchSuff[0][1])
		}
	} else {
		for k, v := range result {
			result1 = append(result1, prejspath+k+"."+v+matchSuff[0][1])
		}
	}
	return result1
}

// 分析内容中的js
func jsFind(s *result.Scan, cont, baseurl string, host, scheme, path, source string, num int) {

	domain := host
	var cata string
	care := regexp.MustCompile("/.*/{1}|/")
	catae := care.FindAllString(path, -1)
	if len(catae) == 0 {
		cata = "/"
	} else {
		cata = catae[0]
	}
	host = scheme + "://" + host
	var jss = []string{}
	for _, re := range config.JsFind {
		reg := regexp.MustCompile(re)
		jssi := reg.FindAllStringSubmatch(cont, -1)
		for _, js := range jssi {
			if js[0] == "" {
				continue
			}

			jss = append(jss, js[1])
		}
	}

	ispake := s.Pakeris[baseurl]

	if !ispake {
		respacker := IsWebpackechunkjs(cont)

		if respacker != "" {
			fmt.Println(host+cata+path, "发现使用webpacker打包,进行解压...")
			tmp := WebpackJsTiQu(respacker)
			if tmp != nil {
				jss = append(jss, tmp...)
				fmt.Println(host, "解压装填完毕")
			}
			config.Lock.Lock()
			s.Pakeris[baseurl] = true
			config.Lock.Unlock()
		}
	}

	jss = jsFilter(jss)

	//js匹配正则
	for _, js := range jss {
		if js == "" {
			continue
		}

		if strings.Contains(js, "https:") || strings.Contains(js, "http:") {
			if !(strings.Contains(js, domain)) {
				continue
			}

			if cmd.G == 0 {
				config.Lock.Lock()
				s.JsResult[baseurl] = append(s.JsResult[baseurl], mode.Link{Url: js, Source: source, Baseurl: baseurl})
				config.Lock.Unlock()
			}

			if num < config.JsSteps {
				s.UrlQueue.Push([]string{js, strconv.Itoa(num + 1), baseurl})
			}

		} else if strings.HasPrefix(js, "//") {
			//不是当前扫描地址的js直接过滤
			if !(strings.Contains(js, domain)) {
				continue
			}
			if cmd.G == 0 {
				config.Lock.Lock()
				s.JsResult[baseurl] = append(s.JsResult[baseurl], mode.Link{Url: scheme + ":" + js, Source: source, Baseurl: baseurl})
				config.Lock.Unlock()
			}

			if num < config.JsSteps {
				s.UrlQueue.Push([]string{scheme + ":" + js, strconv.Itoa(num + 1), baseurl})
			}
		} else if strings.HasPrefix(js, "/") {
			if cmd.G == 0 {
				config.Lock.Lock()
				s.JsResult[baseurl] = append(s.JsResult[baseurl], mode.Link{Url: host + js, Source: source, Baseurl: baseurl})
				config.Lock.Unlock()
			}
			if num < config.JsSteps {
				s.UrlQueue.Push([]string{host + js, strconv.Itoa(num + 1), baseurl})
			}
		} else {
			if cmd.G == 0 {
				config.Lock.Lock()
				s.JsResult[baseurl] = append(s.JsResult[baseurl], mode.Link{Url: host + cata + js, Source: source, Baseurl: baseurl})
				config.Lock.Unlock()
			}
			if num < config.JsSteps {
				s.UrlQueue.Push([]string{host + cata + js, strconv.Itoa(num + 1), baseurl})
			}
		}
	}

}

// 分析内容中的url
func urlFind(s *result.Scan, cont, baseurl string, host, scheme, path, source string, num int) {

	domain := host
	var cata string
	care := regexp.MustCompile("/.*/{1}|/")
	catae := care.FindAllString(path, -1)
	if len(catae) == 0 {
		cata = "/"
	} else {
		cata = catae[0]
	}
	host = scheme + "://" + host
	for _, re := range config.UrlFind {
		reg := regexp.MustCompile(re)
		urls := reg.FindAllStringSubmatch(cont, -1)
		urls = urlFilter(urls)
		//循环提取url放到结果中
		for _, url := range urls {
			if url[1] == "" {
				continue
			}

			//对path进行爬取
			if cmd.M == 2 {
				if strings.Contains(url[1], "https:") || strings.Contains(url[1], "http:") {
					//host外的暂时不进行盲目请求,记录到文档中
					if cmd.G == 0 {
						config.Lock.Lock()
						s.UrlResult[baseurl] = append(s.UrlResult[baseurl], mode.Link{Url: url[1], Source: source, Baseurl: baseurl})
						config.Lock.Unlock()
					}
					if !strings.Contains(url[1], domain) {
						continue
					}
					if num < config.UrlSteps {
						s.UrlQueue.Push([]string{url[1], strconv.Itoa(num + 1), baseurl})
					}

				} else if strings.Contains(url[1], "//") {
					if cmd.G == 0 {
						config.Lock.Lock()
						s.UrlResult[baseurl] = append(s.UrlResult[baseurl], mode.Link{Url: scheme + ":" + url[1], Source: source, Baseurl: baseurl})
						config.Lock.Unlock()
					}
					if !strings.Contains(url[1], domain) {
						continue
					}
					if num < config.UrlSteps {
						s.UrlQueue.Push([]string{scheme + ":" + url[1], strconv.Itoa(num + 1), baseurl})
					}

				} else if strings.HasPrefix(url[1], "/") {

					urlz := ""
					if cmd.B != "" {
						urlz = cmd.B + url[1]
					} else {
						urlz = host + url[1]
					}
					if cmd.G == 0 {

						config.Lock.Lock()
						s.UrlResult[baseurl] = append(s.UrlResult[baseurl], mode.Link{Url: urlz, Source: source, Baseurl: baseurl})
						config.Lock.Unlock()
					}
					if num < config.UrlSteps {
						s.UrlQueue.Push([]string{urlz, strconv.Itoa(num + 1), baseurl})
					}
				} else {
					urlz := ""
					if cmd.B != "" {
						if strings.HasSuffix(cmd.B, "/") {
							urlz = cmd.B + url[1]
						} else {
							urlz = cmd.B + "/" + url[1]
						}
					} else {
						urlz = host + cata + url[1]
					}
					if cmd.G == 0 {

						config.Lock.Lock()
						s.UrlResult[baseurl] = append(s.UrlResult[baseurl], mode.Link{Url: urlz, Source: source, Baseurl: baseurl})
						config.Lock.Unlock()
					}
					if num < config.UrlSteps {
						s.UrlQueue.Push([]string{urlz, strconv.Itoa(num + 1), baseurl})
					}
				}
			} else {
				if strings.Contains(url[1], "https:") || strings.Contains(url[1], "http:") || strings.Contains(url[1], "//") {
					if !strings.Contains(url[1], domain) {
						continue
					}
				} else if strings.HasPrefix(url[1], "/") {
					urlz := ""
					if cmd.B != "" {
						urlz = cmd.B + url[1]
					} else {
						urlz = host + url[1]
					}
					if cmd.G == 0 {

						config.Lock.Lock()
						s.UrlResult[baseurl] = append(s.UrlResult[baseurl], mode.Link{Url: urlz, Source: source, Baseurl: baseurl})
						config.Lock.Unlock()
					}
				} else {
					urlz := ""
					if cmd.B != "" {
						if strings.HasSuffix(cmd.B, "/") {
							urlz = cmd.B + url[1]
						} else {
							urlz = cmd.B + "/" + url[1]
						}
					} else {
						urlz = host + cata + url[1]
					}
					if cmd.G == 0 {

						config.Lock.Lock()
						s.UrlResult[baseurl] = append(s.UrlResult[baseurl], mode.Link{Url: urlz, Source: source, Baseurl: baseurl})
						config.Lock.Unlock()
					}
				}
			}
		}
	}
}

// 分析内容中的敏感信息
func infoFind(s *result.Scan, cont, baseurl string, source string) {
	info := mode.Info{}
	//手机号码
	for i := range config.Phone {
		phones := regexp.MustCompile(config.Phone[i]).FindAllStringSubmatch(cont, -1)
		for i := range phones {
			info.Phone = append(info.Phone, phones[i][1])
		}
	}
	//for i := range config.Email {
	//	emails := regexp.MustCompile(config.Email[i]).FindAllStringSubmatch(cont, -1)
	//	for i := range emails {
	//		info.Email = append(info.Email, emails[i][1])
	//	}
	//}

	for i := range config.IDcard {
		IDcards := regexp.MustCompile(config.IDcard[i]).FindAllStringSubmatch(cont, -1)
		for i := range IDcards {
			info.IDcard = append(info.IDcard, IDcards[i][1])
		}
	}

	for i := range config.Jwt {
		Jwts := regexp.MustCompile(config.Jwt[i]).FindAllStringSubmatch(cont, -1)
		for i := range Jwts {
			info.JWT = append(info.JWT, Jwts[i][0])
		}
	}
	for i := range config.Other {

		Others := regexp.MustCompile(config.Other[i]).FindAllStringSubmatch(cont, -1)

		for i := range Others {
			if strings.Contains(Others[i][0], "function") {
				continue
			}
			if Others[i][0] == "" {
				continue
			}
			info.Other = append(info.Other, Others[i][0])
		}
	}

	info.Source = source
	info.Baseurl = baseurl
	if len(info.Phone) != 0 || len(info.IDcard) != 0 || len(info.JWT) != 0 || len(info.Email) != 0 || len(info.Other) != 0 {
		config.Lock.Lock()
		s.InfoResult[baseurl] = append(s.InfoResult[baseurl], info)
		config.Lock.Unlock()
	}
}
