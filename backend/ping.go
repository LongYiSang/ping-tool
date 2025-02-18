package backend

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/go-ping/ping"
)

// PingTask 定义Ping任务的结构
type PingTask struct {
	Target    string
	Interval  time.Duration
	IsRunning bool
	Results   []PingResult
	StopChan  chan bool
}

// PingResult 定义单次Ping的结果
type PingResult struct {
	Timestamp time.Time
	RTT       time.Duration
	Success   bool
	Error     string
	IP        string
}

// runPingTask 运行Ping任务
func (a *App) runPingTask(task *PingTask) {
	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	// 预处理目标地址
	target := task.Target
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		if u, err := url.Parse(target); err == nil {
			target = u.Hostname()
		}
	}

	for {
		select {
		case <-task.StopChan:
			return
		case <-ticker.C:
			result := PingResult{
				Timestamp: time.Now(),
				Success:   false,
			}

			// 先尝试解析域名
			ips, err := net.LookupHost(target)
			if err != nil {
				result.Error = fmt.Sprintf("DNS解析失败: %v", err)
				a.mutex.Lock()
				task.Results = append(task.Results, result)
				if len(task.Results) > 1000 {
					task.Results = task.Results[1:]
				}
				a.mutex.Unlock()
				continue
			}

			// 使用第一个解析到的IP地址
			pingTarget := ips[0]
			result.IP = pingTarget

			// 创建新的 pinger
			pinger, err := ping.NewPinger(pingTarget)
			if err != nil {
				result.Error = fmt.Sprintf("创建Ping失败: %v", err)
				a.mutex.Lock()
				task.Results = append(task.Results, result)
				if len(task.Results) > 1000 {
					task.Results = task.Results[1:]
				}
				a.mutex.Unlock()
				continue
			}

			// Windows系统需要设置特权模式为true
			pinger.SetPrivileged(true)
			// 只发送一个包
			pinger.Count = 1
			// 设置超时时间为2秒
			pinger.Timeout = time.Second * 2
			// 设置ICMP包大小
			pinger.Size = 32

			err = pinger.Run()
			if err != nil {
				result.Error = fmt.Sprintf("Ping执行失败: %v", err)
			} else {
				stats := pinger.Statistics()
				if stats.PacketsRecv > 0 {
					result.Success = true
					result.RTT = stats.AvgRtt
				} else {
					result.Error = "请求超时"
				}
			}

			a.mutex.Lock()
			task.Results = append(task.Results, result)
			if len(task.Results) > 1000 {
				task.Results = task.Results[1:]
			}
			a.mutex.Unlock()
		}
	}
}

// StartPing 开始对指定目标进行Ping监控
func (a *App) StartPing(target string, interval int) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if task, exists := a.pingTasks[target]; exists && task.IsRunning {
		return nil
	}

	task := &PingTask{
		Target:    target,
		Interval:  time.Duration(interval) * time.Millisecond,
		IsRunning: true,
		StopChan:  make(chan bool),
		Results:   make([]PingResult, 0),
	}

	a.pingTasks[target] = task
	go a.runPingTask(task)
	return nil
}

// StopPing 停止对指定目标的Ping监控
func (a *App) StopPing(target string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if task, exists := a.pingTasks[target]; exists && task.IsRunning {
		task.IsRunning = false
		close(task.StopChan)
		delete(a.pingTasks, target)
	}
}

// GetPingResults 获取指定目标的Ping结果
func (a *App) GetPingResults(target string) []PingResult {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	if task, exists := a.pingTasks[target]; exists {
		return task.Results
	}
	return nil
}
