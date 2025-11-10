package main

import (
	"fmt"
	"github.com/weisir1/URLGo/crawler"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
)

//
//func detailedRequest(url string, timeout int) {
//	clientTimeout := time.Duration(timeout) * time.Second
//
//	tr := &http.Transport{
//		TLSClientConfig: &tls.Config{
//			InsecureSkipVerify: true,
//			ClientSessionCache: tls.NewLRUClientSessionCache(2048),
//		},
//		Proxy: http.ProxyFromEnvironment,
//		DialContext: (&net.Dialer{
//			Timeout:   30 * time.Second,
//			KeepAlive: 30 * time.Second,
//			Control: func(network, address string, c syscall.RawConn) error {
//				// 只允许 IPv4
//				if network == "tcp6" || network == "udp6" {
//					return fmt.Errorf("IPv6 disabled, only IPv4 allowed")
//				}
//				return nil
//			},
//		}).DialContext,
//
//		MaxIdleConns:        100,
//		MaxIdleConnsPerHost: 1,
//		MaxConnsPerHost:     1,
//
//		IdleConnTimeout:       90 * time.Second,
//		TLSHandshakeTimeout:   30 * time.Second,
//		ExpectContinueTimeout: 1 * time.Second,
//		DisableCompression:    false,
//
//		// ✅ 添加这个，自动关闭连接
//		DisableKeepAlives: false,
//	}
//
//	client := &http.Client{
//		Timeout:   clientTimeout,
//		Transport: tr,
//	}
//
//	// ✅ 添加详细的 trace
//	var (
//		dnsStart, connectStart, tlsStart, firstByteStart time.Time
//		requestStart                                     = time.Now()
//	)
//
//	trace := &httptrace.ClientTrace{
//		DNSStart: func(info httptrace.DNSStartInfo) {
//			dnsStart = time.Now()
//			fmt.Printf("[%.1fs] → DNS 查询开始\n", time.Since(requestStart).Seconds())
//		},
//		DNSDone: func(info httptrace.DNSDoneInfo) {
//			elapsed := time.Since(dnsStart)
//			fmt.Printf("[%.1fs] ✅ DNS 完成 耗时: %v\n",
//				time.Since(requestStart).Seconds(), elapsed)
//			if info.Err != nil {
//				fmt.Printf("       ❌ DNS 错误: %v\n", info.Err)
//			}
//		},
//
//		ConnectStart: func(network, addr string) {
//			connectStart = time.Now()
//			fmt.Printf("[%.1fs] → TCP 连接开始: %s\n",
//				time.Since(requestStart).Seconds(), addr)
//		},
//		ConnectDone: func(network, addr string, err error) {
//			elapsed := time.Since(connectStart)
//			if err != nil {
//				fmt.Printf("[%.1fs] ❌ TCP 连接失败: %v\n",
//					time.Since(requestStart).Seconds(), err)
//			} else {
//				fmt.Printf("[%.1fs] ✅ TCP 连接完成 耗时: %v\n",
//					time.Since(requestStart).Seconds(), elapsed)
//			}
//		},
//
//		TLSHandshakeStart: func() {
//			tlsStart = time.Now()
//			fmt.Printf("[%.1fs] → TLS 握手开始\n",
//				time.Since(requestStart).Seconds())
//		},
//		TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
//			elapsed := time.Since(tlsStart)
//			if err != nil {
//				fmt.Printf("[%.1fs] ❌ TLS 握手失败: %v\n",
//					time.Since(requestStart).Seconds(), err)
//			} else {
//				fmt.Printf("[%.1fs] ✅ TLS 握手完成 耗时: %v\n",
//					time.Since(requestStart).Seconds(), elapsed)
//			}
//		},
//
//		GotConn: func(info httptrace.GotConnInfo) {
//			fmt.Printf("[%.1fs] ✅ 获得连接 (复用: %v)\n",
//				time.Since(requestStart).Seconds(), info.Reused)
//		},
//
//		GotFirstResponseByte: func() {
//			firstByteStart = time.Now()
//			fmt.Printf("[%.1fs] ✅ 收到第一个响应字节\n",
//				time.Since(firstByteStart).Seconds())
//		},
//
//		Got100Continue: func() {
//			fmt.Printf("[%.1fs] ℹ️  收到 100 Continue\n",
//				time.Since(requestStart).Seconds())
//		},
//	}
//
//	// 创建请求
//	req, err := http.NewRequest("GET", url, nil)
//	if err != nil {
//		fmt.Printf("❌ 创建请求失败: %v\n", err)
//		return
//	}
//	if cmd.I {
//		util.SetHeadersConfig(&req.Header)
//	}
//
//	//req.Header.Set("User-Agent", util.GetUserAgent())
//	//req.Header.Set("Accept", "*/*")
//
//	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
//	req.Header.Set("Accept", "*/*")
//	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
//	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
//	req.Header.Set("Cache-Control", "no-cache")
//	// ✅ 添加 trace
//	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
//
//	// 发送请求
//	fmt.Printf("[%.1fs] → 发送请求\n", time.Since(requestStart).Seconds())
//	doStart := time.Now()
//	response, err := client.Do(req)
//	doElapsed := time.Since(doStart)
//
//	if err != nil {
//		fmt.Printf("[%.1fs] ❌ 请求失败: %v (耗时: %v)\n",
//			time.Since(requestStart).Seconds(), err, doElapsed)
//		return
//	}
//
//	fmt.Printf("[%.1fs] ✅ 收到响应 (耗时: %v)\n",
//		time.Since(requestStart).Seconds(), doElapsed)
//	fmt.Printf("       状态码: %d\n", response.StatusCode)
//	fmt.Printf("       Content-Length: %d\n", response.ContentLength)
//	fmt.Printf("       Transfer-Encoding: %v\n", response.TransferEncoding)
//
//	defer response.Body.Close()
//
//	// ✅ 逐步读取响应体，而不是一次性读取
//	fmt.Printf("[%.1fs] → 开始读取响应体\n", time.Since(requestStart).Seconds())
//
//	readStart := time.Now()
//	bytesRead := int64(0)
//	buffer := make([]byte, 256*1024) // 32KB 缓冲
//
//	for {
//		n, err := response.Body.Read(buffer)
//
//		if n > 0 {
//			bytesRead += int64(n)
//			elapsed := time.Since(readStart)
//			speed := float64(bytesRead) / 1024 / elapsed.Seconds()
//
//			// 每 100KB 打印一次进度
//			if bytesRead%(100*1024) < int64(n) {
//				fmt.Printf("[%.1fs] 已读: %d bytes (%.2f KB/s)\n",
//					time.Since(requestStart).Seconds(), bytesRead, speed)
//			}
//		}
//
//		if err != nil {
//			if err == io.EOF {
//				elapsed := time.Since(readStart)
//				fmt.Printf("[%.1fs] ✅ 读取完成 总字节: %d 耗时: %v\n",
//					time.Since(requestStart).Seconds(), bytesRead, elapsed)
//				break
//			} else {
//				elapsed := time.Since(readStart)
//				fmt.Printf("[%.1fs] ❌ 读取失败: %v (已读: %d bytes, 耗时: %v)\n",
//					time.Since(requestStart).Seconds(), err, bytesRead, elapsed)
//				break
//			}
//		}
//	}
//
//	totalElapsed := time.Since(requestStart)
//	fmt.Printf("\n" + "================================" + "\n")
//	fmt.Printf("总耗时: %v (%.1f 秒)\n", totalElapsed, totalElapsed.Seconds())
//	fmt.Printf("总字节: %d (%.2f MB)\n", bytesRead, float64(bytesRead)/1024/1024)
//	if bytesRead > 0 {
//		fmt.Printf("平均速度: %.2f KB/s\n", float64(bytesRead)/1024/totalElapsed.Seconds())
//	}
//	fmt.Printf("==================================" + "\n")
//}

func main() {
	log.SetOutput(io.Discard)
	//util.GetUpdate()
	//config.JsSteps = 1
	//config.UrlSteps = 1
	//cmd.M = 2
	//cmd.F = "url.txt"
	//cmd.X = "http://127.0.0.1:8080"
	//cmd.S = "all"
	//cmd.M = 2
	//cmd.G = 1c
	//cmd.U = "http://221.229.120.26"
	//cmd.Parse()
	go func() {
		fmt.Println("pprof server running at http://localhost:6060/debug/pprof/")
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			fmt.Println("pprof server error:", err)
		}
	}()
	crawler.Run()
	//detailedRequest("https://tz.jsfic.com.cn/jsfile/ezgo-components-pc/EZGOCOMPONENTS.umd.js", 300)

}
