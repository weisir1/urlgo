package config

import (
	"fmt"
	"github.com/weisir1/URLGo/cmd"
	"github.com/weisir1/URLGo/mode"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strings"
	"sync"
)

var Conf mode.Config
var Progress = 1
var FuzzNum int

var (
	Risks = []string{"remove", "delete", "insert", "update", "logout"}

	JsFuzzPath = []string{
		"login.js",
		"app.js",
		"main.js",
		"config.js",
		"admin.js",
		"info.js",
		"open.js",
		"user.js",
		"input.js",
		"list.js",
		"upload.js",
	}
	JsFind = []string{
		"(https{0,1}:[-a-zA-Z0-9（）@:%_\\+.~#?&//=]{2,250}?[-a-zA-Z0-9（）@:%_\\+.~#?&//=]{3}[.]js)",
		//"[\"'‘“`]{0,1}\\s{0,6}(/{0,1}[-a-zA-Z0-9（）@:%_\\+.~#?&//=]{2,250}[.]js)[\"'‘“`]{0,1}",
		"[\"'‘“`]\\s{0,6}(/{0,1}[-a-zA-Z0-9（）@:%_\\+.~#?&//=]{2,250}?[-a-zA-Z0-9（）@:%_\\+.~#?&//=]{3}[.]js)",
		"=\\s{0,6}[\",',’,”]{0,1}\\s{0,6}(/{0,1}[-a-zA-Z0-9（）@:%_\\+.~#?&//=]{2,250}?[-a-zA-Z0-9（）@:%_\\+.~#?&//=]{3}[.]js)",
		//`.{0,70}\{(?:"?[a-zA-Z0-9_\-\/]+"?:"?[a-f0-9]+"?,?\s*)+\}.{0,60}\+.{0,30}".{0,16}\.js"`,
	}

	UrlFind = []string{
		//"[\"'‘“`]\\s{0,6}(https{0,1}:[-a-zA-Z0-9()@:%_\\+.~#?&//={}]{2,250}?)\\s{0,6}[\"'‘“`]",
		//path
		`['"]((?:\/|\.\.\/|\.\/)[^\/\>\< \)\(\{\}\,\'\"\\]([^\>\< \)\(\{\}\,\'\"\\])*?)['"]`,
		`['"]([^\/\>\< \)\(\{\}\,\'\"\\][\w\/]*?\/[\w\/]*?)['"]`,
		//url
		`["'](\s{0,6}(https{0,1}:[-a-zA-Z0-9()@:%_\+.~#?&//={}]{3,250}?)\s{0,6})["']`,
		// `"([-a-zA-Z0-9()@:%_\+.~#?&={}/]+?[/]{1}[-a-zA-Z0-9()@:%_\+.~#?&={}]+?)"`,
		//"\"([-a-zA-Z0-9()@:%_\\+.~#?&={}/]+?[/]{1}[-a-zA-Z0-9()@:%_\\+.~#?&={}]+?)\"",
		//"\"([-a-zA-Z0-9()@:%_\\+.~#?&={}/]+?[/]{1}[-a-zA-Z0-9()@:%_\\+.~#?&={}]+?)\"",
		//"=\\s{0,6}(https{0,1}:[-a-zA-Z0-9()@:%_\\+.~#?&//={}]{2,250})",
		//"href\\s{0,6}=\\s{0,6}[\"'‘“`]{0,1}\\s{0,6}([-a-zA-Z0-9()@:%_\\+.~#?&//={}]{2,250})|action\\s{0,6}=\\s{0,6}[\"'‘“`]{0,1}\\s{0,6}([-a-zA-Z0-9()@:%_\\+.~#?&//={}]{2,250})",
	}

	JsFiler = []string{
		"www\\.w3\\.org",
		"example\\.com",
	}
	UrlFiler = []string{
		"\\.js\\?|\\.css\\?|\\.jpeg\\?|\\.jpg\\?|\\.png\\?|.gif\\?|www\\.w3\\.org|example\\.com|\\<|\\>|\\{|\\}|\\[|\\]|\\||\\^|;|/js/|\\.src|\\.replace|\\.url|\\.att|\\.href|location\\.href|javascript:|location:|application/x-www-form-urlencoded|application/json|\\.createObject|:location|\\.path|\\*#__PURE__\\*|\\*\\$0\\*|\\n|text/css|text/javascript|text/xml|text/plain|text/html|image/jpeg|image/png|.*\\.js$|.*\\.css$|.*\\.scss$|.*,$|.*\\.jpeg$|.*\\.jpg$|.*\\.png$|.*\\.gif$|.*\\.ico$|.*\\.svg$|.*\\.vue$|.*\\.ts$|/text/css$|text/javascript$|(?i)M/D/yy$",
	}

	Phone  = []string{"['\"](1(3([0-35-9]\\d|4[1-8])|4[14-9]\\d|5([\\d]\\d|7[1-79])|66\\d|7[2-35-8]\\d|8\\d{2}|9[89]\\d)\\d{7})['\"]"}
	Email  = []string{"['\"]([\\w!#$%&'*+=?^_`{|}~-]+(?:\\.[\\w!#$%&'*+=?^_`{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[\\w](?:[\\w-]*[\\w])?)['\"]"}
	IDcard = []string{"['\"]((\\d{8}(0\\d|10|11|12)([0-2]\\d|30|31)\\d{3}$)|(\\d{6}(18|19|20)\\d{2}(0[1-9]|10|11|12)([0-2]\\d|30|31)\\d{3}(\\d|X|x)))['\"]"}
	Jwt    = []string{
		//auth
		`[Bb]earer\s+[a-zA-Z0-9\-=._+/\\]{20,500}|[Bb]asic\s+[A-Za-z0-9+/]{18,}={0,2}|(ey[A-Za-z0-9_-]{10,}\.[A-Za-z0-9._-]{10,}|ey[A-Za-z0-9_\/+-]{10,}\.[A-Za-z0-9._\/+-]{10,})`,
		`["']?[Aa]uthorization["']?\s*[:=]\s*['"]?\b(?:[Tt]oken\s+)?[a-zA-Z0-9\-_+/]{10,500}['"]?`,
		//`([Bb]earer\s+[a-zA-Z0-9\-=._+/\\]{20,500})|([Bb]asic\s+[A-Za-z0-9+/]{18,}={0,2})|(eyJrIjoi[a-zA-Z0-9\-_+/]{50,100}={0,2})|(["''\[]*[Aa]uthorization["''\]]*\s*[:=]\s*[''"]?\b(?:[Tt]oken\s+)?[a-zA-Z0-9\-_+/]{20,500}[''"]?)`,
	}
	//Other  = []string{"(access.{0,1}key|access.{0,1}Key|access.{0,1}Id|access.{0,1}id|.{0,5}密码|.{0,5}账号|默认.{0,5}|加密|解密|password:.{0,10}|username:.{0,10})"}
	//Other = []string{
	//	`["']?(admin[_-]?email|app[_-]?id|access[_-]?key[_-]?id|account[_-]?sid|access[_-]?token|access[_-]?secret|app[_-]?key|access[_-]?key|password|username|cameraindexcode|username|pwd|user|encryptkey|bucket|app[_-]?token|app[_-]?secret|secret)["']?\s*[:=]\s*["']?([^"',\s]+|"[^"]*"|'[^']*')["']?`,
	//	`["']?[\w_-]*?(?:password|username|accesskey|token)["']?[^\S\r\n]*[=:][^\S\r\n]*["']?[\w-]+["']?`,
	//	//,`["']?[-]+BEGIN \w+ PRIVATE KEY[-]+`,
	//	`["']?huawei\.oss\.(ak|sk|bucket\.name|endpoint|local\.path)["']?[^\S\r\n]*[=:][^\S\r\n]*["']?[\w-]+["']?`,
	//	`["']?private[_-]?key[_-]?(id)?["']?[^\S\r\n]*[=:][^\S\r\n]*["']?[\w-]+["']?`,
	//	`["']?account[_-]?(name|key)?["']?[^\S\r\n]*[=:][^\S\r\n]*["']?[\w-]+["']?`,
	//}
	Other = []string{
		//敏感
		//`["']?(admin[_-]?email|app[_-]?id|username|account|account[_-]?(name|key)?|account[_-]?sid|(?i)[\w_-]*?token[\w_-]*?|[\w_-]*?secret[\w_-]*?|private[_-]?key[_-]?|app[_-]?key|[\w_-]*access[_-]?key[\w_-]*|cameraindexcode|user|encryptkey|[\w_-]*?bucket[\w_-]*?|[\w_-]*?api[_-]?key[\w_-]*?)["']?\s*[:=]\s*["']?[a-z0-9!@#$%&*]["']?`,
		`["']?(admin[_-]?email|app[_-]?id|username|account|account[_-]?(name|key)?|account[_-]?sid|(?i)[\w_-]*?token[\w_-]*?|[\w_-]*?secret[\w_-]*?|private[_-]?key[_-]?|app[_-]?key|[\w_-]*access[_-]?key[\w_-]*|cameraindexcode|user|encryptkey|[bB]ucket|[\w_-]*?api[_-]?key[\w_-]*?)["']?\s*[:=]\s*["']?[\p{Han}a-zA-Z0-9/\-\_]{2,}["']?`,
		//密码
		`(?i)(?:admin_?pass|password|[a-z]{3,15}_?password|user_?pass|user_?pwd|admin_?pwd|pwd)\\?['"]*\s*[:=]\s*\\?['"][a-z0-9!@#$%&*=]{5,20}\\?['"]`,

		//云key
		`LTAI[A-Za-z\d]{12,30}|AKID[A-Za-z\d]{13,40}|JDC_[0-9A-Z]{25,40}|APID[a-zA-Z0-9]{32,42}|AIza[0-9A-Za-z_\-]{35}|AKLT[a-zA-Z0-9_\-]{16,28}|AKTP[a-zA-Z0-9_\-]{16,28}`,

		//wxid ghid
		`["'](wx[a-z0-9]{15,18})|(ww[a-z0-9]{15,18})|(gh_[a-z0-9]{11,13})["']`,
		`(oWebControl.JS_RequestInterface)|(jsWebControl)`,
	}
)

var (
	UrlSteps = 2
	JsSteps  = 4
	Module   = 0
)

var (
	Lock sync.Mutex
	Wg   sync.WaitGroup
	Mux  sync.Mutex
	//Ch    = make(chan int, 50)
	Jsch  chan int
	Urlch chan int

	JsFindRegexps   []*regexp.Regexp
	UrlFindRegexps  []*regexp.Regexp
	InfoFindRegexps map[string][]*regexp.Regexp
)

func Init(threadNum int) {
	// 初始化Channel（如果还需要全局channel）
	//Ch = make(chan int, threadNum)
	Jsch = make(chan int, threadNum*3/10)
	Urlch = make(chan int, threadNum*7/10)

	//  预编译JS查找正则
	JsFindRegexps = make([]*regexp.Regexp, len(JsFind))
	for i, re := range JsFind {
		JsFindRegexps[i] = regexp.MustCompile(re)
	}

	//  预编译URL查找正则
	UrlFindRegexps = make([]*regexp.Regexp, len(UrlFind))
	for i, re := range UrlFind {
		UrlFindRegexps[i] = regexp.MustCompile(re)
	}

	//  预编译信息查找正则
	InfoFindRegexps = make(map[string][]*regexp.Regexp)
	for key, patterns := range map[string][]string{
		"Phone":  Phone,
		"Email":  Email,
		"IDcard": IDcard,
		"Jwt":    Jwt,
		"Other":  Other,
	} {
		regexps := make([]*regexp.Regexp, len(patterns))
		for i, pattern := range patterns {
			regexps[i] = regexp.MustCompile(pattern)
		}
		InfoFindRegexps[key] = regexps
	}

	fmt.Println("配置初始化完成")
}

// 读取配置文件
func GetConfig(path string) {
	if f, err := os.Open(path); err != nil {
		if strings.Contains(err.Error(), "The system cannot find the file specified") || strings.Contains(err.Error(), "no such file or directory") {
			Conf.Headers = map[string]string{"Cookie": cmd.C, "Accept": "*/*"}
			Conf.Proxy = ""
			Conf.JsFind = JsFind
			Conf.UrlFind = UrlFind
			Conf.JsFiler = JsFiler
			Conf.UrlFiler = UrlFiler
			Conf.JsFuzzPath = JsFuzzPath
			Conf.Module = cmd.G
			Conf.JsSteps = JsSteps
			Conf.UrlSteps = UrlSteps
			Conf.Risks = Risks
			Conf.Timeout = cmd.TI
			Conf.Thread = cmd.T
			Conf.Max = cmd.MA
			Conf.InfoFind = map[string][]string{"Phone": Phone, "Email": Email, "IDcard": IDcard, "Jwt": Jwt, "Other": Other}
			data, err2 := yaml.Marshal(Conf)
			err2 = os.WriteFile(path, data, 0644)
			if err2 != nil {
				fmt.Println(err)
			} else {
				fmt.Println("未找到配置文件,已在当面目录下创建配置文件: config.yaml,请重新运行命令(程序运行时首先以config.yaml参数加载)")
			}
		} else {
			fmt.Println("配置文件错误,请尝试重新生成配置文件")
			fmt.Println(err)
		}
		os.Exit(1)
	} else {
		yaml.NewDecoder(f).Decode(&Conf)
		JsFind = Conf.JsFind
		UrlFind = Conf.UrlFind
		JsFiler = Conf.JsFiler
		UrlFiler = Conf.UrlFiler
		JsFuzzPath = Conf.JsFuzzPath
		Phone = Conf.InfoFind["Phone"]
		Email = Conf.InfoFind["Email"]
		IDcard = Conf.InfoFind["IDcard"]
		Jwt = Conf.InfoFind["Jwt"]
		Other = Conf.InfoFind["Other"]
		JsSteps = Conf.JsSteps
		cmd.G = Conf.Module
		UrlSteps = Conf.UrlSteps
		Risks = Conf.Risks
		cmd.T = Conf.Thread
		cmd.MA = Conf.Max
		cmd.TI = Conf.Timeout
	}

}
