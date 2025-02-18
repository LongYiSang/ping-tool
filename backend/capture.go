package backend

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// Packet 定义单个数据包的结构
type Packet struct {
	Timestamp time.Time `json:"timestamp"` // 捕获时间
	Protocol  string    `json:"protocol"`  // 协议
	SrcIP     string    `json:"srcIP"`     // 源IP
	DstIP     string    `json:"dstIP"`     // 目标IP
	SrcPort   int       `json:"srcPort"`   // 源端口
	DstPort   int       `json:"dstPort"`   // 目标端口
	Length    int       `json:"length"`    // 数据包长度
	Info      string    `json:"info"`      // 附加信息
	Payload   string    `json:"payload"`   // 数据包内容
	RawData   string    `json:"rawData"`   // 原始数据的十六进制表示
	HTTPInfo  *HTTPInfo `json:"httpInfo"`  // HTTP信息
}

// HTTPInfo 定义HTTP协议相关信息
type HTTPInfo struct {
	Method      string `json:"method"`      // 请求方法
	Path        string `json:"path"`        // 请求路径
	Version     string `json:"version"`     // HTTP版本
	StatusCode  int    `json:"statusCode"`  // 状态码
	StatusText  string `json:"statusText"`  // 状态描述
	ContentType string `json:"contentType"` // 内容类型
	Host        string `json:"host"`        // 主机
	IsRequest   bool   `json:"isRequest"`   // 是否为请求
}

// PacketCapture 定义抓包任务的结构
type PacketCapture struct {
	Interface    string       // 网络接口名称
	Filter       string       // BPF过滤规则
	IsRunning    bool         // 是否正在运行
	StopChan     chan bool    // 停止信号
	PacketChan   chan *Packet // 数据包通道
	handle       *pcap.Handle // pcap句柄
	mu           sync.Mutex   // 互斥锁
	PacketBuffer []*Packet    // 数据包缓冲区
	MaxPackets   int          // 最大保存的数据包数量
}

// CaptureStats 定义抓包统计信息
type CaptureStats struct {
	TotalPackets int   `json:"totalPackets"` // 总包数
	TCPPackets   int   `json:"tcpPackets"`   // TCP包数
	UDPPackets   int   `json:"udpPackets"`   // UDP包数
	ICMPPackets  int   `json:"icmpPackets"`  // ICMP包数
	TotalBytes   int64 `json:"totalBytes"`   // 总字节数
	StartTime    int64 `json:"startTime"`    // 开始时间
}

// GetInterfaces 获取可用的网络接口列表
func (a *App) GetInterfaces() ([]string, error) {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return nil, fmt.Errorf("获取网络接口失败: %v", err)
	}

	var interfaces []string
	for _, device := range devices {
		// 添加接口名称和描述
		desc := device.Description
		if desc == "" {
			desc = "无描述"
		}
		interfaces = append(interfaces, fmt.Sprintf("%s (%s)", device.Name, desc))
	}

	return interfaces, nil
}

// StartCapture 开始抓包
func (a *App) StartCapture(interfaceName, filter string) error {
	a.captureMu.Lock()
	defer a.captureMu.Unlock()

	if a.capture.IsRunning {
		return fmt.Errorf("抓包已在运行中")
	}

	// 从接口名称中提取实际的设备名
	deviceName := strings.Split(interfaceName, " ")[0]

	// 打开网络接口
	handle, err := pcap.OpenLive(deviceName, 65535, true, pcap.BlockForever)
	if err != nil {
		return fmt.Errorf("打开网络接口失败: %v", err)
	}

	// 设置过滤器
	if filter != "" {
		if err := handle.SetBPFFilter(filter); err != nil {
			handle.Close()
			return fmt.Errorf("设置过滤器失败: %v", err)
		}
	}

	a.capture.handle = handle
	a.capture.Interface = deviceName
	a.capture.Filter = filter
	a.capture.IsRunning = true
	a.capture.PacketBuffer = make([]*Packet, 0)
	a.capture.StopChan = make(chan bool)

	// 启动抓包协程
	go a.capturePackets()

	return nil
}

// StopCapture 停止抓包
func (a *App) StopCapture() error {
	a.captureMu.Lock()
	defer a.captureMu.Unlock()

	if !a.capture.IsRunning {
		return fmt.Errorf("抓包未在运行")
	}

	close(a.capture.StopChan)
	if a.capture.handle != nil {
		a.capture.handle.Close()
	}
	a.capture.IsRunning = false
	return nil
}

// GetPackets 获取已捕获的数据包
func (a *App) GetPackets() []*Packet {
	a.capture.mu.Lock()
	defer a.capture.mu.Unlock()

	// 返回数据包的副本
	packets := make([]*Packet, len(a.capture.PacketBuffer))
	copy(packets, a.capture.PacketBuffer)
	return packets
}

// parseHTTPInfo 解析HTTP协议信息
func parseHTTPInfo(payload []byte) *HTTPInfo {
	if len(payload) == 0 {
		return nil
	}

	lines := strings.Split(string(payload), "\r\n")
	if len(lines) == 0 {
		return nil
	}

	info := &HTTPInfo{}
	firstLine := lines[0]

	// 判断是请求还是响应
	if strings.HasPrefix(firstLine, "HTTP/") {
		// 解析HTTP响应
		parts := strings.SplitN(firstLine, " ", 3)
		if len(parts) >= 3 {
			info.Version = parts[0]
			statusCode, err := strconv.Atoi(parts[1])
			if err == nil {
				info.StatusCode = statusCode
			}
			info.StatusText = parts[2]
			info.IsRequest = false
		}
	} else {
		// 解析HTTP请求
		parts := strings.SplitN(firstLine, " ", 3)
		if len(parts) >= 3 {
			info.Method = parts[0]
			info.Path = parts[1]
			info.Version = parts[2]
			info.IsRequest = true
		}
	}

	// 解析其他HTTP头
	for _, line := range lines[1:] {
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			key := strings.ToLower(parts[0])
			value := parts[1]
			switch key {
			case "content-type":
				info.ContentType = value
			case "host":
				info.Host = value
			}
		}
	}

	// 如果没有解析出有效信息，返回nil
	if info.Method == "" && info.StatusCode == 0 {
		return nil
	}

	return info
}

// capturePackets 实际的抓包过程
func (a *App) capturePackets() {
	packetSource := gopacket.NewPacketSource(a.capture.handle, a.capture.handle.LinkType())

	for {
		select {
		case <-a.capture.StopChan:
			return
		default:
			packet, err := packetSource.NextPacket()
			if err != nil {
				continue
			}

			// 解析数据包
			p := &Packet{
				Timestamp: packet.Metadata().Timestamp,
				Length:    packet.Metadata().Length,
			}

			// 解析网络层（IP）
			ipLayer := packet.Layer(layers.LayerTypeIPv4)
			if ipLayer != nil {
				ip, _ := ipLayer.(*layers.IPv4)
				p.SrcIP = ip.SrcIP.String()
				p.DstIP = ip.DstIP.String()
				p.Protocol = ip.Protocol.String()
			}

			// 解析传输层（TCP/UDP）
			tcpLayer := packet.Layer(layers.LayerTypeTCP)
			if tcpLayer != nil {
				tcp, _ := tcpLayer.(*layers.TCP)
				p.SrcPort = int(tcp.SrcPort)
				p.DstPort = int(tcp.DstPort)
				p.Protocol = "TCP"

				// 添加TCP标志信息
				flags := make([]string, 0)
				if tcp.SYN {
					flags = append(flags, "SYN")
				}
				if tcp.ACK {
					flags = append(flags, "ACK")
				}
				if tcp.FIN {
					flags = append(flags, "FIN")
				}
				if tcp.RST {
					flags = append(flags, "RST")
				}
				if len(flags) > 0 {
					p.Info = fmt.Sprintf("Flags: %s", strings.Join(flags, ","))
				}
			}

			udpLayer := packet.Layer(layers.LayerTypeUDP)
			if udpLayer != nil {
				udp, _ := udpLayer.(*layers.UDP)
				p.SrcPort = int(udp.SrcPort)
				p.DstPort = int(udp.DstPort)
				p.Protocol = "UDP"
			}

			// 解析应用层数据
			applicationLayer := packet.ApplicationLayer()
			if applicationLayer != nil {
				payload := applicationLayer.Payload()
				if len(payload) > 0 {
					// 保存原始数据的十六进制表示
					p.RawData = fmt.Sprintf("%x", payload)
					// 尝试解析为文本
					p.Payload = formatPayload(payload)
					// 尝试解析HTTP信息
					if p.Protocol == "TCP" && (p.DstPort == 80 || p.SrcPort == 80 ||
						p.DstPort == 443 || p.SrcPort == 443 ||
						p.DstPort == 8080 || p.SrcPort == 8080) {
						p.HTTPInfo = parseHTTPInfo(payload)
					}
				}
			}

			// 更新数据包缓冲区
			a.capture.mu.Lock()
			a.capture.PacketBuffer = append(a.capture.PacketBuffer, p)
			// 限制缓冲区大小
			if len(a.capture.PacketBuffer) > a.capture.MaxPackets {
				a.capture.PacketBuffer = a.capture.PacketBuffer[1:]
			}
			a.capture.mu.Unlock()
		}
	}
}

// formatPayload 格式化数据包内容
func formatPayload(payload []byte) string {
	// 如果数据为空，直接返回
	if len(payload) == 0 {
		return ""
	}

	// 尝试不同的编码方式
	// 1. 首先尝试 UTF-8
	if str := string(payload); isPrintableASCII(str) {
		return str
	}

	// 2. 检查是否为 UTF-16 编码
	if len(payload) >= 2 {
		// 检查 UTF-16 BOM
		if (payload[0] == 0xFF && payload[1] == 0xFE) || (payload[0] == 0xFE && payload[1] == 0xFF) {
			return fmt.Sprintf("[UTF-16 encoded data] 0x%x", payload)
		}
	}

	// 3. 检查是否为二进制数据
	isBinary := false
	for _, b := range payload {
		if b < 32 && b != '\t' && b != '\n' && b != '\r' {
			isBinary = true
			break
		}
	}

	if isBinary {
		return fmt.Sprintf("[Binary data] 0x%x", payload)
	}

	// 4. 如果看起来像是文本但包含一些不可打印字符
	return fmt.Sprintf("[Encoded/Corrupted text] 0x%x", payload)
}

// isPrintableASCII 检查字符串是否为可打印ASCII
func isPrintableASCII(s string) bool {
	for _, r := range s {
		if r > 127 || !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

// GetCaptureStats 获取抓包统计信息
func (a *App) GetCaptureStats() *CaptureStats {
	a.capture.mu.Lock()
	defer a.capture.mu.Unlock()

	stats := &CaptureStats{
		TotalPackets: len(a.capture.PacketBuffer),
		TCPPackets:   0,
		UDPPackets:   0,
		ICMPPackets:  0,
		TotalBytes:   0,
	}

	for _, p := range a.capture.PacketBuffer {
		stats.TotalBytes += int64(p.Length)
		switch p.Protocol {
		case "TCP":
			stats.TCPPackets++
		case "UDP":
			stats.UDPPackets++
		case "ICMP":
			stats.ICMPPackets++
		}
	}

	return stats
}
