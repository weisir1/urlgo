package crawler

import (
	"github.com/weisir1/URLGo/config"
	"github.com/weisir1/URLGo/util"
	"net/url"
	"regexp"
	"strings"
)

func jsFilter(str []string) []string {
	//对不需要的数据过滤
	for i := range str {
		str[i], _ = url.QueryUnescape(str[i])
		str[i] = strings.TrimSpace(str[i])
		str[i] = strings.Replace(str[i], " ", "", -1)
		str[i] = strings.Replace(str[i], "\\/", "/", -1)
		str[i] = strings.Replace(str[i], "%3A", ":", -1)
		str[i] = strings.Replace(str[i], "%2F", "/", -1)
		str[i] = strings.Replace(str[i], "'", "", -1)
		str[i] = strings.Replace(str[i], "=", "", -1)
		str[i] = strings.Replace(str[i], `"`, ``, -1)
		//去除不是.js的链接
		if !strings.HasSuffix(str[i], ".js") && !strings.Contains(str[i], ".js?") {
			str[i] = ""
			continue
		}
		if strings.HasPrefix(str[i], "./") {
			str[i] = strings.Replace(str[i], "./", "/", -1)
		}
		//对提取的js进行去重

		//过滤配置的黑名单
		//for i2 := range config.JsFiler {
		//	re := regexp.MustCompile(config.JsFiler[i2])
		//	is := re.MatchString(str[i])
		//	if is {
		//		str[i] = ""
		//		break
		//	}
		//}

	}
	str = util.RemoveDuplicates(str)

	return str

}

// 过滤URL
func urlFilter(str [][]string) [][]string {
	Risks := []string{"remove", "delete", "insert", "update", "logout"}
	//对不需要的数据过滤
	for i := range str {
		if str[i][1] == "" {
			continue
		}

		//fmt.Println(str[i])
		//str[i][1], _ = url.QueryUnescape(str[i][1])
		str[i][1], _ = url.QueryUnescape(str[i][1])
		str[i][1] = strings.TrimSpace(str[i][1])
		str[i][1] = strings.Replace(str[i][1], ":id", "1", -1)
		str[i][1] = strings.Replace(str[i][1], "\\/", "/", -1)
		str[i][1] = strings.Replace(str[i][1], "%3A", ":", -1)
		str[i][1] = strings.Replace(str[i][1], "%2F", "/", -1)
		str[i][1] = strings.Replace(str[i][1], "'", "", -1)
		str[i][1] = strings.Replace(str[i][1], `"`, "", -1)
		if str[i][1] == "/" || str[i][1] == "//" {
			str[i][1] = ""
		}
		if strings.HasPrefix(str[i][1], "./") {
			str[i][1] = strings.Replace(str[i][1], `./`, "/", -1)
		}

		for i2 := range config.UrlFiler {
			re := regexp.MustCompile(config.UrlFiler[i2])
			is := re.MatchString(str[i][1])
			if is {
				str[i][1] = ""
				break
			}
		}
		//对抓到的域名做处理
		//re := regexp.MustCompile("([a-z0-9\\-]+\\.)+([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?").FindAllString(str[i][0], 1)
		//if len(re) != 0 && !string s.HasPrefix(str[i][1], "http") && !strings.HasPrefix(str[i][1], "/") {
		//	str[i][1] = "http://" + str[i][1]
		//}

		for _, v := range Risks {
			if strings.Contains(str[i][1], v) {
				str[i][1] = ""
			}
		}

		//过滤配置的黑名单

	}
	str = util.RemoveDuplicate(str)

	return str
}
