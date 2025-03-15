package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// 定义全局变量 cstZone
var cstZone *time.Location

func init() {
	var err error
	cstZone, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		fmt.Printf("加载时区失败: %v\n", err)
		// 如果加载时区失败，程序无法正常运行，可以选择退出
		panic("无法加载中国标准时间时区")
	}
}

const (
	SLEEPTIME        = 0.0     // 每次抢座间隔
)

// Config 用户配置结构体
type Config struct {
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	Time      []string `json:"time"`
	RoomID     string   `json:"roomid"`
	SeatID     []string `json:"seatid"`
	DaysOfWeek []string `json:"daysofweek"`
	StartTime  string   `json:"starttime"` // 开始时间字段
	EndTime    string   `json:"endtime"`   // 结束时间字段，将通过计算得到
	LoginTime  string   `json:"logintime"` // 登录时间字段，将通过计算得到
}

type ConfigFile struct {
	Reserve []Config `json:"reserve"`
}

func getCurrentTime() string {
	return time.Now().In(cstZone).Format("15:04:05")
}

func getCurrentDayOfWeek() string {
	return time.Now().In(cstZone).Weekday().String()
}

func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var configFile ConfigFile
	if err := json.Unmarshal(data, &configFile); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	if len(configFile.Reserve) == 0 {
		return nil, fmt.Errorf("配置文件中没有预约配置")
	}

	cfg := &configFile.Reserve[0]

	// 解析开始时间
	startTime, err := time.ParseInLocation("15:04:05", cfg.StartTime, cstZone)
	if err != nil {
		return nil, fmt.Errorf("解析开始时间失败: %v", err)
	}

	// 计算结束时间和登录时间
	endTime := startTime.Add(time.Minute) // 结束时间为开始时间后一分钟
	loginTime := startTime.Add(-time.Minute) // 登录时间为开始时间前一分钟

	// 格式化时间字符串
	cfg.EndTime = endTime.Format("15:04:05")
	cfg.LoginTime = loginTime.Format("15:04:05")

	return cfg, nil
}

func main() {

	// 设置日志格式
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetPrefix("[座位预约] ")

	// 解析命令行参数
	configPath := flag.String("u", "config.json", "配置文件路径")
	flag.Parse()

	// 获取配置文件的绝对路径
	absConfigPath, err := filepath.Abs(*configPath)
	if err != nil {
		log.Fatalf("无法解析配置文件路径: %v", err)
	}

	// 加载配置
	cfg, err := loadConfig(absConfigPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 检查当前时间
	currentTime := getCurrentTime()
	currentDay := getCurrentDayOfWeek()
	log.Printf("当前时间: %s, 星期: %s", currentTime, currentDay)

	// 检查是否是预约日
	if !contains(cfg.DaysOfWeek, currentDay) {
		log.Println("今天不是预约日")
		return
	}

	// 创建预约实例
	r := NewReserve(SLEEPTIME)

	// 执行预约
	if err := runReserve(r, cfg); err != nil {
		log.Fatalf("预约失败: %v", err)
	}
}

func runReserve(r *Reserve, cfg *Config) error {
	var STARTTIME string = cfg.StartTime

    // 解析固定的开始和结束时间
    start, err := time.ParseInLocation("15:04:05", STARTTIME, cstZone)
    if err != nil {
        return fmt.Errorf("解析开始时间失败: %v", err)
    }

    end, err := time.ParseInLocation("15:04:05", cfg.EndTime, cstZone)
    if err != nil {
        return fmt.Errorf("解析结束时间失败: %v", err)
    }

	login, err := time.ParseInLocation("15:04:05", cfg.LoginTime, cstZone)
	if err != nil {
		return fmt.Errorf("解析登录时间失败: %v", err)
	}

    // 获取当前时间
    now := time.Now().In(cstZone)

    // 设置今天的开始执行时间
    LoginTime := time.Date(
        now.Year(), now.Month(), now.Day(),
        login.Hour(), login.Minute(), login.Second(),
        0, cstZone,
    )
    
    // 设置今天的开始执行时间
    scheduledTime := time.Date(
        now.Year(), now.Month(), now.Day(),
        start.Hour(), start.Minute(), start.Second(),
        0, cstZone,
    )

    // 设置今天的结束时间
    todayEndTime := time.Date(
        now.Year(), now.Month(), now.Day(),
        end.Hour(), end.Minute(), end.Second(),
        0, cstZone,
    )

    // 如果当前时间已经超过今天的结束时间，直接返回
    if now.After(todayEndTime) {
		log.Printf("今天已到达截至时间 请手动修改配置后重试")
		os.Exit(0)
    }

	currentTime := getCurrentTime()
	log.Printf("当前时间: %s", currentTime)

	// 如果还没到登录时间，等待
	if now.Before(LoginTime) {
		waitDuration := login.Sub(now)
		log.Printf("将在 %s 登录系统", LoginTime.Format("15:04:05"))
		time.Sleep(waitDuration)
	}

	// 获取登录状态
	if err := r.GetLoginStatus(); err != nil {
		return fmt.Errorf("获取登录状态失败: %v", err)
	}

	// 登录系统
	LoginSuccess, msg := r.Login(cfg.Username, cfg.Password)
	if !LoginSuccess {
		return fmt.Errorf("登录失败: %s", msg)
	}


    // 如果还没到开始时间，等待
    if now.Before(scheduledTime) {
        waitDuration := scheduledTime.Sub(now)
        log.Printf("账号 %s 登录成功 定时任务将在 %s 开始执行", cfg.Username, scheduledTime.Format("15:04:05"))
        time.Sleep(waitDuration)
    }

    log.Printf("开始执行预约任务")

	// 计算预约日期
	day := time.Now().In(cstZone)
	// 获取token
	url := fmt.Sprintf(r.GetURL(), cfg.RoomID, day.Format("2006-01-02"))
	token, err := r.getPageToken(url)
	if err != nil {
		return fmt.Errorf("获取token失败: %v", err)
	}
	
	log.Printf("获取到token: %s", token)

    success := r.Submit(cfg.Time, cfg.RoomID, cfg.SeatID, token)

    if success {
        log.Printf("预约成功! 时间: %s", time.Now().In(cstZone).Format("15:04:05"))
		log.Printf("程序结束")
        return nil
    }

    return fmt.Errorf("预约失败: 达到结束时间 %s", end.Format("15:04:05"))
}