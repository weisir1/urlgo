## URLgo

URLgo是一款快速、全面、易用的页面信息提取工具

用于分析页面中的js与url,搜集js中的敏感信息

有什么需求或bug欢迎各位师傅提交lssues

## 快速使用
单url
```
显示全部状态码
URLgo.exe -u http://www.baidu.com -s all -m 2

显示200和403状态码
URLgo.exe -u http://www.baidu.com -s 200,403 -m 2
```
批量url
```
导出全部,默认文件名运行时时间,仅限导出xlsx文件
URLgo.exe -s all -f url.txt 
```
参数（更多参数使用 -i 配置）：
```
 -a string
        set user-agent
        设置user-agent请求头
  -b string
        set baseurl
        设置baseurl路径
  -c string
        set cookie
        设置cookie
  -f string
        set urlFile
        批量抓取url,指定文件路径,默认url.txt
  -h    this help
        帮助信息
  -i    set configFile
        加载yaml配置文件（不存在时,会在当前目录创建一个默认yaml配置文件）
  -m int
        set mode
        抓取模式
           1 normal
             正常请求（爬取的path路径不进行请求,首页/返回windows.location无法获取）
           2 thorough
             全面抓取（默认,对爬取的path路径进行请求,速度较慢）
           (default 2)
  -max int
        set maximum
        最大抓取链接数 (default 99999)
  -o string
        set outFile
        结果导出到xlsx文件,需指定导出文件目录及xlsx后缀,否则生成到当前目录下以时间命名 (default ".")
  -s string
        set Status
        显示指定状态码,all为显示全部（多个状态码用,隔开）
  -t int
        set Thread
        设置线程数（默认50） (default 50)
  -time int
        set Timeout
        设置超时时间（默认5,单位秒） (default 5)
  -u string
        set Url
        目标URL
  -x string
        set Proxy
        设置代理,格式: http://username:password@127.0.0.1:8809
```
## 使用截图

![image-20250106144238900](img\image-20250106144238900.png)

![image-20250106150739664](img\image-20250106150739664.png)
![image-20250106150935678](img\image-20250106150935678.png)

![image-20250106151139430](img\image-20250106151139430.png)

![image-20250106151439237](img\image-20250106151439237.png)

![image-20250106151533319](img\image-20250106151533319.png)

## 部分说明

fuzz功能是基于抓到的404目录和路径。将其当作字典,随机组合并碰撞出有效路径,从而解决路径拼接错误的问题

结果会优先显示输入的url顶级域名,其他域名不做区分显示在 other

结果会优先显示200,按从小到大排序（输入的域名最优先,就算是404也会排序在其他子域名的200前面）

为了更好的兼容和防止漏抓链接,放弃了低误报率,错误的链接会变多但漏抓概率变低,可通过 ‘-s 200’ 筛选状态码过滤无效的链接（但不推荐只看200状态码）

## 更新说明
2025/03/01 
工具首次上传

# 开发由来
对URLFinder表示感谢,在面对大量资产时,需要一款好用的工具来对web进行接口,敏感信息搜集,来提高漏洞挖掘效率,而URLFinder在使用-f参数请求时,采用是单线路爬取url,在攻防比赛中速度很慢,且不支持对webpack的解析提取,本工具对urlfinder做了彻底性的修改,能够更快速地在大量资产中提取其path、js、敏感信息，另外可以根据工具提取的path、js结果来在攻击前做初步的了解，如js多的、或path路径较多较敏感的优先进行测试

