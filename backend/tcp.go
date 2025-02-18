package backend

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

// TCPResult 定义TCP连接测试的结果
type TCPResult struct {
	Timestamp   time.Time
	ConnectTime time.Duration
	Success     bool
	Error       string
	IP          string
}

// TestTCPConnection 测试TCP连接
func (a *App) TestTCPConnection(host string, port int, timeout int) (*TCPResult, error) {
	result := &TCPResult{
		Timestamp: time.Now(),
		Success:   false,
	}

	// 预处理主机地址
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		if u, err := url.Parse(host); err == nil {
			host = u.Hostname()
		}
	}

	// 解析域名获取IP
	ips, err := net.LookupHost(host)
	if err != nil {
		result.Error = fmt.Sprintf("DNS解析失败: %v", err)
		return result, nil
	}
	result.IP = ips[0]

	// 开始计时
	startTime := time.Now()

	// 尝试建立TCP连接
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), time.Duration(timeout)*time.Millisecond)
	if err != nil {
		result.Error = fmt.Sprintf("连接失败: %v", err)
		return result, nil
	}
	defer conn.Close()

	// 计算连接时间
	result.ConnectTime = time.Since(startTime)
	result.Success = true

	return result, nil
}
