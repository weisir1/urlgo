package cmd

import (
	"flag"
	"fmt"
	"github.com/gookit/color"
	"os"
)

var Update = "2025.01.02"
var XUpdate string

var (
	H  bool
	I  bool
	M  int
	S  string
	U  string
	D  string
	C  string
	A  string
	B  string
	F  string
	FF string
	O  string
	//nO string
	X  string
	T  = 50
	TI = 50
	MA = 99999
	Z  int
)

func init() {
	flag.StringVar(&A, "a", "", "set user-agent\n设置user-agent请求头")
	flag.StringVar(&B, "b", "", "set baseurl\n设置baseurl路径")
	flag.StringVar(&C, "c", "", "set cookie\n设置cookie")
	//flag.StringVar(&D, "d", "", "set domainName\n指定获取的域名,支持正则表达式")
	flag.StringVar(&F, "f", "", "set urlFile\n批量抓取url,指定文件路径,默认url.txt")
	//flag.StringVar(&FF, "ff", "", "set urlFile one\n与-f区别：全部抓取的数据,视为同一个url的结果来处理（只打印一份结果 | 只会输出一份结果）")
	flag.BoolVar(&H, "h", false, "this help\n帮助信息")
	flag.BoolVar(&I, "i", true, "set configFile\n加载yaml配置文件（不存在时,会在当前目录创建一个默认yaml配置文件）")
	flag.IntVar(&M, "m", 1, "set mode\n抓取模式 \n   1 normal\n     正常请求（默认,爬取的path路径不进行请求 \n   2 thorough\n     全面抓取（默认,对爬取的path路径进行请求,速度较慢） \n  ")
	flag.IntVar(&MA, "max", 99999, "set maximum\n最大抓取链接数")
	flag.StringVar(&O, "o", ".", "set outFile\n结果导出到xlsx文件,需指定导出文件目录及xlsx后缀,否则生成到当前目录下以时间命名")
	//flag.StringVar(&O, "o", ".", "废弃")
	//flag.StringVar(&O, "no", "", "不进行输出,打印到控制台")
	flag.StringVar(&S, "s", "", "set Status\n显示指定状态码,all为显示全部（多个状态码用,隔开）")
	flag.IntVar(&T, "t", 50, "set Thread\n设置线程数（默认50）")
	flag.IntVar(&TI, "time", 5, "set Timeout\n设置超时时间（默认5,单位秒）")
	flag.StringVar(&U, "u", "", "set Url\n目标URL")
	flag.StringVar(&X, "x", "", "set Proxy\n设置代理,格式: http://username:password@127.0.0.1:8809")
	//flag.IntVar(&Z, "z", 0, "set Fuzz\n对404链接进行fuzz(只对主域名下的链接生效,需要与 -s 一起使用） \n   1 decreasing\n     目录递减fuzz \n   2 2combination\n     2级目录组合fuzz（适合少量链接使用） \n   3 3combination\n     3级目录组合fuzz（适合少量链接使用） ")
	Z = 0
	// 改变默认的 Usage
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: URLGo [-a user-agent] [-b baseurl] [-c cookie][-f urlFile] [-h help]  [-i configFile]  [-m mode] [-max maximum] [-o outFile]  [-s Status] [-t thread] [-time timeout] [-u url] [-x proxy] 

Options:
`)
	flag.PrintDefaults()
}

func Parse() {
	color.LightCyan.Printf("                    $$\\                     \n                    $$ |                    \n$$\\   $$\\  $$$$$$\\  $$ | $$$$$$\\   $$$$$$\\  \n$$ |  $$ |$$  __$$\\ $$ |$$  __$$\\ $$  __$$\\ \n$$ |  $$ |$$ |  \\__|$$ |$$ /  $$ |$$ /  $$ |\n$$ |  $$ |$$ |      $$ |$$ |  $$ |$$ |  $$ |\n\\$$$$$$  |$$ |      $$ |\\$$$$$$$ |\\$$$$$$  |\n \\______/ \\__|      \\__| \\____$$ | \\______/ \n                        $$\\   $$ |          \n                        \\$$$$$$  |          \n                         \\______/           \n")
	flag.Parse()
}
