package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type Reserve struct {
	loginPage      string
	url           string
	submitURL     string
	loginURL      string
	client        *http.Client
	sleepTime     float64
}

var loginHeaders = http.Header{
	"Accept":           []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3"},
	"Accept-Language":  []string{"zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7"},
	"User-Agent":       []string{"Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_1 like Mac OS X) AppleWebKit/603.1.3 (KHTML, like Gecko) Version/10.0 Mobile/14E304 Safari/602.1 wechatdevtools/1.05.2109131 MicroMessenger/8.0.5 Language/zh_CN webview/16364215743155638"},
	"X-Requested-With": []string{"XMLHttpRequest"},
	"Content-Type":     []string{"application/x-www-form-urlencoded; charset=UTF-8"},
	"Host":             []string{"passport2.chaoxing.com"},
}

var defaultHeaders = http.Header{
        "User-Agent": []string{"Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_1 like Mac OS X) AppleWebKit/603.1.3 (KHTML, like Gecko) Version/10.0 Mobile/14E304 Safari/602.1 wechatdevtools/1.05.2109131 MicroMessenger/8.0.5 Language/zh_CN webview/16364215743155638"},
        "Host":       []string{"reserve.chaoxing.com"},
}

// 可选的服务器代理URL，默认为空表示不使用代理
var ProxyServer string = ""

var DEBUGMODE = false

func NewReserve(sleepTime float64) *Reserve {
    jar, _ := cookiejar.New(nil)
    tr := &http.Transport{
        TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
        DisableCompression:  false,                  // 启用压缩
        ForceAttemptHTTP2:   true,                  // 启用 HTTP/2
    }

    // 仅当设置了代理服务器时才使用代理
    if ProxyServer != "" {
        if proxy, err := url.Parse(ProxyServer); err == nil {
            tr.Proxy = http.ProxyURL(proxy)
        }
    }

	client := &http.Client{
		Transport: tr,
		Jar:      jar,
	}

	return &Reserve{
		loginPage:    "https://passport2.chaoxing.com/mlogin?loginType=1&newversion=true&fid=",
		url:         "https://reserve.chaoxing.com/front/third/apps/seatengine/select?id=%s&day=%s&backLevel=2&seatId=796",
		submitURL:   "https://reserve.chaoxing.com/data/apps/seatengine/submit",
		loginURL:    "https://passport2.chaoxing.com/fanyalogin",
		client:      client,
		sleepTime:   sleepTime,
	}
}

func (r *Reserve) GetURL() string {
	return r.url
}

func (r *Reserve) GetLoginStatus() error {
	req, err := http.NewRequest("GET", r.loginPage, nil)

	if err != nil {
		return err
	}

	req.Header = loginHeaders
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// 登录相关方法
func (r *Reserve) Login(username, password string) (bool, string) {
	// 加密用户名和密码
	encUsername := AES_Encrypt(username)
	encPassword := AES_Encrypt(password)

	params := url.Values{
		"fid":      {"-1"},
		"uname":    {encUsername},
		"password": {encPassword},
		"refer":    {"http%3A%2F%2Foffice.chaoxing.com%2Ffront%2Fthird%2Fapps%2Fseat%2Fcode%3Fid%3D4219%26seatNum%3D380"},
		"t":        {"true"},
	}

	// 发送登录请求
	req, err := http.NewRequest("POST", r.loginURL, strings.NewReader(params.Encode()))

	if err != nil {
		return false, err.Error()
	}

	req.Header = loginHeaders
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := r.client.Do(req)
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()

	// 解析响应
	var result struct {
		Status bool   `json:"status"`
		Msg2   string `json:"msg2"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err.Error()
	}

	return result.Status, result.Msg2
}

// 提交预约
func (r *Reserve) Submit(times []string, roomID string, seatIDs []string, token string) bool {
    day := time.Now().In(cstZone).Format("2006-01-02")

    
    // 轮询每个座位
    for _, seatID := range seatIDs {
        submitParams := map[string]string{
            "roomId":    roomID,
            "startTime": times[0],
            "endTime":   times[1],
            "day":       day,
            "seatNum":   seatID,
            "captcha":   "",
            "token":     token,
        }

        // 添加加密参数
        submitParams["enc"] = enc(submitParams)

        // 构建请求参数
        values := url.Values{}
        for key, value := range submitParams {
            values.Add(key, value)
        }

        // 构建URL
        urlWithParams := r.submitURL + "?" + values.Encode()

        // 创建请求
        req, err := http.NewRequest("POST", urlWithParams, nil)

        if err != nil {
            continue
        }

		getHeaders := defaultHeaders.Clone()
		getHeaders.Set("Content-Type", "application/json")

        // 设置请求头
        req.Header = getHeaders

        // 发送请求
        resp, err := r.client.Do(req)
        if err != nil {
            continue
        }

        // 读取响应内容
        bodyBytes, err := io.ReadAll(resp.Body)
        if err != nil {
            log.Printf("Failed to read response body: %v", err)
            resp.Body.Close()
            continue
        }

        // 解析响应
        var result struct {
            Success bool `json:"success"`
        }
        if err := json.Unmarshal(bodyBytes, &result); err != nil {
			// 输出当前尝试信息
			currentTime := time.Now().In(cstZone).Format("2006-01-02 15:04:05")
			log.Printf("当前座位: %s, 当前时间: %s, 成功状态: %v", seatID, currentTime, "false")
			if (DEBUGMODE) {
				log.Printf("请求返回内容: %s", string(bodyBytes))
			} else {
				log.Printf("请求返回内容: %s", "输出非json异常内容 请将reserve.go的DEBUGMODE改为true查看信息")
			}

            resp.Body.Close()

			// 添加短暂延时，避免请求过于频繁
			time.Sleep(time.Duration(r.sleepTime) * time.Second)

            continue
        }
        resp.Body.Close()

        // 输出当前尝试信息
        currentTime := time.Now().In(cstZone).Format("2006-01-02 15:04:05")
        log.Printf("当前座位: %s, 当前时间: %s, 成功状态: %v", seatID, currentTime, result.Success)
        log.Printf("请求返回内容: %s", string(bodyBytes))

        if result.Success {
            return true
        }

		// 添加短暂延时，避免请求过于频繁
		time.Sleep(time.Duration(r.sleepTime) * time.Second)

    }

	log.Printf("程序运行结束")
	os.Exit(0)
    return false
}

// 将正则表达式编译为全局变量，避免重复编译
var tokenRegexp = regexp.MustCompile(`token = '(.*?)'`)

// 获取页面token
func (r *Reserve) getPageToken(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return "", err
	}

	req.Header = defaultHeaders.Clone()
	resp, err := r.client.Do(req)
    if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

    matches := tokenRegexp.FindStringSubmatch(string(body))
	if len(matches) > 1 {
		return matches[1], nil
	}
	return "", fmt.Errorf("未找到token")
}