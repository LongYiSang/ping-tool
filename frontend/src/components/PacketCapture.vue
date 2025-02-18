<script>
import { GetInterfaces, StartCapture, StopCapture, GetPackets, GetCaptureStats } from '../../wailsjs/go/backend/App'

export default {
  data() {
    return {
      interfaces: [],          // 网络接口列表
      selectedInterface: '',   // 选中的接口
      filterRules: [],        // 过滤规则列表
      isCapturing: false,     // 是否正在抓包
      packets: [],            // 捕获的数据包
      stats: {                // 统计信息
        totalPackets: 0,
        tcpPackets: 0,
        udpPackets: 0,
        icmpPackets: 0,
        totalBytes: 0
      },
      updateTimer: null,      // 更新定时器
      maxDisplayPackets: 1000, // 最大显示包数
      // 过滤规则选项
      filterOptions: {
        type: [
          { label: '协议', value: 'protocol' },
          { label: '源IP', value: 'src_ip' },
          { label: '目标IP', value: 'dst_ip' },
          { label: '源端口', value: 'src_port' },
          { label: '目标端口', value: 'dst_port' }
        ],
        protocol: [
          { label: 'TCP', value: 'tcp' },
          { label: 'UDP', value: 'udp' },
          { label: 'ICMP', value: 'icmp' }
        ],
        operator: [
          { label: '等于', value: '=' },
          { label: '包含', value: 'contains' }
        ]
      },
      encodings: [
        { label: '自动检测', value: 'auto' },
        { label: 'UTF-8', value: 'utf8' },
        { label: 'UTF-16', value: 'utf16' },
        { label: 'GB2312', value: 'gb2312' },
        { label: 'GBK', value: 'gbk' },
        { label: 'Big5', value: 'big5' },
        { label: '十六进制', value: 'hex' }
      ],
      packetEncodings: {}, // 存储每个数据包的当前编码设置
      httpDisplayFormats: [
        { label: 'UTF-8', value: 'utf8' },
        { label: 'GB2312', value: 'gb2312' },
        { label: 'GBK', value: 'gbk' },
        { label: 'Big5', value: 'big5' },
        { label: '十六进制', value: 'hex' }
      ],
      packetHttpFormats: {} // 存储每个数据包的 HTTP 显示格式
    }
  },
  async mounted() {
    await this.loadInterfaces()
  },
  beforeDestroy() {
    if (this.updateTimer) {
      clearInterval(this.updateTimer)
    }
    if (this.isCapturing) {
      this.stopCapture()
    }
  },
  methods: {
    async loadInterfaces() {
      try {
        this.interfaces = await GetInterfaces()
        if (this.interfaces.length > 0) {
          this.selectedInterface = this.interfaces[0]
        }
      } catch (err) {
        alert('获取网络接口失败: ' + err)
      }
    },
    async startCapture() {
      if (!this.selectedInterface) {
        alert('请选择网络接口')
        return
      }

      try {
        const filterString = this.generateFilterString()
        await StartCapture(this.selectedInterface, filterString)
        this.isCapturing = true
        this.updateTimer = setInterval(this.updateData, 1000)
      } catch (err) {
        alert('启动抓包失败: ' + err)
      }
    },
    async stopCapture() {
      try {
        await StopCapture()
        this.isCapturing = false
        if (this.updateTimer) {
          clearInterval(this.updateTimer)
          this.updateTimer = null
        }
      } catch (err) {
        alert('停止抓包失败: ' + err)
      }
    },
    async updateData() {
      try {
        // 更新数据包列表
        const newPackets = await GetPackets()
        
        // 保持之前的显示状态和编码设置
        this.packets = newPackets.slice(-this.maxDisplayPackets).map(packet => {
          const existingPacket = this.packets.find(p => p.timestamp === packet.timestamp)
          return {
            ...packet,
            showPayload: existingPacket ? existingPacket.showPayload : false
          }
        })

        // 更新统计信息
        this.stats = await GetCaptureStats()
      } catch (err) {
        console.error('更新数据失败:', err)
      }
    },
    formatBytes(bytes) {
      if (bytes < 1024) return bytes + ' B'
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
      return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
    },
    formatTimestamp(timestamp) {
      return new Date(timestamp).toLocaleTimeString()
    },
    clearPackets() {
      this.packets = []
    },
    // 添加新的过滤规则
    addFilterRule() {
      this.filterRules.push({
        type: '',
        operator: '=',
        value: '',
        protocol: ''
      })
    },
    // 删除过滤规则
    removeFilterRule(index) {
      this.filterRules.splice(index, 1)
    },
    // 生成BPF过滤规则字符串
    generateFilterString() {
      if (this.filterRules.length === 0) return ''

      return this.filterRules.map(rule => {
        switch (rule.type) {
          case 'protocol':
            return rule.protocol.toLowerCase()
          case 'src_ip':
            return `src host ${rule.value}`
          case 'dst_ip':
            return `dst host ${rule.value}`
          case 'src_port':
            return `src port ${rule.value}`
          case 'dst_port':
            return `dst port ${rule.value}`
          default:
            return ''
        }
      }).filter(rule => rule !== '').join(' and ')
    },
    togglePayload(packet) {
      if (!packet.hasOwnProperty('showPayload')) {
        this.$set(packet, 'showPayload', true)
      } else {
        packet.showPayload = !packet.showPayload
      }
    },
    // 切换数据包编码
    toggleEncoding(packet, encoding) {
      if (!packet.rawData) return;
      
      try {
        const hexString = packet.rawData;
        const bytes = new Uint8Array(hexString.match(/.{1,2}/g).map(byte => parseInt(byte, 16)));
        
        if (encoding === 'hex') {
          packet.payload = `[HEX] 0x${packet.rawData}`;
          return;
        }
        
        if (encoding === 'auto') {
          packet.payload = this.autoDetectEncoding(bytes);
          return;
        }
        
        // 使用 TextDecoder 进行解码
        const decoder = new TextDecoder(encoding, { fatal: true });
        packet.payload = decoder.decode(bytes);
      } catch (err) {
        packet.payload = `[解码失败: ${encoding}] 0x${packet.rawData}`;
      }
    },
    
    // 自动检测编码
    autoDetectEncoding(bytes) {
      // 检查是否为 UTF-16
      if (bytes.length >= 2) {
        if ((bytes[0] === 0xFF && bytes[1] === 0xFE) || (bytes[0] === 0xFE && bytes[1] === 0xFF)) {
          try {
            const decoder = new TextDecoder('utf-16');
            return decoder.decode(bytes);
          } catch (e) {}
        }
      }
      
      // 尝试 UTF-8
      try {
        const decoder = new TextDecoder('utf-8', { fatal: true });
        return decoder.decode(bytes);
      } catch (e) {}
      
      // 如果都失败了，返回十六进制
      return `[Binary data] 0x${Array.from(bytes).map(b => b.toString(16).padStart(2, '0')).join('')}`;
    },
    // 切换 HTTP 信息显示格式
    toggleHttpFormat(packet, format) {
      if (!packet.httpInfo) return;
      
      try {
        // 保存当前选择的格式
        this.$set(this.packetHttpFormats, packet.timestamp, format);
        
        const decodeWithFormat = (text, format) => {
          if (!text) return '';
          if (format === 'hex') {
            const bytes = new TextEncoder().encode(text);
            return `0x${Array.from(bytes).map(b => b.toString(16).padStart(2, '0')).join('')}`;
          }
          
          try {
            // 先将文本转换为 UTF-8 字节
            const utf8Bytes = new TextEncoder().encode(text);
            // 然后用选定的编码格式解码
            const decoder = new TextDecoder(format);
            return decoder.decode(utf8Bytes);
          } catch (e) {
            return `[解码失败] ${text}`;
          }
        };

        // 创建新的 HTTP 信息对象
        const newHttpInfo = {
          ...packet.httpInfo,
          method: packet.httpInfo.isRequest ? decodeWithFormat(packet.httpInfo.method, format) : packet.httpInfo.method,
          path: packet.httpInfo.isRequest ? decodeWithFormat(packet.httpInfo.path, format) : packet.httpInfo.path,
          statusText: !packet.httpInfo.isRequest ? decodeWithFormat(packet.httpInfo.statusText, format) : packet.httpInfo.statusText,
          host: packet.httpInfo.host ? decodeWithFormat(packet.httpInfo.host, format) : packet.httpInfo.host,
          contentType: packet.httpInfo.contentType ? decodeWithFormat(packet.httpInfo.contentType, format) : packet.httpInfo.contentType
        };

        // 使用 Vue 的响应式系统更新 HTTP 信息
        this.$set(packet, 'httpInfo', newHttpInfo);
      } catch (err) {
        console.error('HTTP 信息格式转换失败:', err);
      }
    }
  }
}
</script>

<template>
  <div class="capture-container">
    <!-- 控制面板 -->
    <div class="control-panel">
      <div class="input-group">
        <select v-model="selectedInterface" class="input">
          <option value="">选择网络接口</option>
          <option v-for="iface in interfaces" :key="iface" :value="iface">
            {{ iface }}
          </option>
        </select>
      </div>

      <!-- 过滤规则面板 -->
      <div class="filter-panel">
        <div class="filter-header">
          <h4>过滤规则</h4>
          <button @click="addFilterRule" class="add-button">
            <span class="plus-icon">+</span>
            添加规则
          </button>
        </div>
        <div class="filter-rules">
          <div v-for="(rule, index) in filterRules" :key="index" class="filter-rule">
            <select v-model="rule.type" class="filter-select">
              <option value="">选择类型</option>
              <option v-for="opt in filterOptions.type" 
                      :key="opt.value" 
                      :value="opt.value">
                {{ opt.label }}
              </option>
            </select>

            <template v-if="rule.type === 'protocol'">
              <select v-model="rule.protocol" class="filter-select">
                <option value="">选择协议</option>
                <option v-for="opt in filterOptions.protocol" 
                        :key="opt.value" 
                        :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
            </template>

            <template v-else-if="rule.type">
              <select v-model="rule.operator" class="filter-select">
                <option v-for="opt in filterOptions.operator" 
                        :key="opt.value" 
                        :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
              <input v-model="rule.value" 
                     :placeholder="rule.type.includes('port') ? '输入端口号' : '输入值'"
                     :type="rule.type.includes('port') ? 'number' : 'text'"
                     class="filter-input" />
            </template>

            <button @click="removeFilterRule(index)" class="remove-button">
              ×
            </button>
          </div>
        </div>
      </div>

      <div class="button-group">
        <button @click="startCapture" :disabled="isCapturing" class="button">
          开始抓包
        </button>
        <button @click="stopCapture" :disabled="!isCapturing" class="button stop">
          停止抓包
        </button>
        <button @click="clearPackets" class="button clear">
          清除数据
        </button>
      </div>
    </div>

    <!-- 统计信息面板 -->
    <div class="stats-panel" v-if="stats.totalPackets > 0">
      <div class="stat-item">
        <span class="label">总包数:</span>
        <span class="value">{{ stats.totalPackets }}</span>
      </div>
      <div class="stat-item">
        <span class="label">TCP包数:</span>
        <span class="value">{{ stats.tcpPackets }}</span>
      </div>
      <div class="stat-item">
        <span class="label">UDP包数:</span>
        <span class="value">{{ stats.udpPackets }}</span>
      </div>
      <div class="stat-item">
        <span class="label">ICMP包数:</span>
        <span class="value">{{ stats.icmpPackets }}</span>
      </div>
      <div class="stat-item">
        <span class="label">总流量:</span>
        <span class="value">{{ formatBytes(stats.totalBytes) }}</span>
      </div>
    </div>

    <!-- 数据包列表 -->
    <div class="packet-panel">
      <div class="panel-header">
        <h3>抓包日志</h3>
        <button @click="clearPackets" class="clear-button">清除日志</button>
      </div>
      <div class="packet-list">
        <div v-for="(packet, index) in packets.slice().reverse()" :key="index" 
             class="packet-item"
             :class="{ 'expanded': packet.showPayload }">
          <div class="packet-header" @click="togglePayload(packet)">
            <div class="packet-time">{{ formatTimestamp(packet.timestamp) }}</div>
            <div class="packet-content">
              <span class="protocol" :class="packet.protocol.toLowerCase()">{{ packet.protocol }}</span>
              <span class="address">{{ packet.srcIP }}:{{ packet.srcPort }}</span>
              <span class="arrow">→</span>
              <span class="address">{{ packet.dstIP }}:{{ packet.dstPort }}</span>
              <span class="length">[{{ packet.length }} bytes]</span>
              <span v-if="packet.info" class="info">{{ packet.info }}</span>
              <span class="expand-icon" :class="{ 'expanded': packet.showPayload }">▼</span>
            </div>
          </div>
          <div v-if="packet.showPayload" class="packet-payload">
            <div class="encoding-controls">
              <span class="encoding-label">编码:</span>
              <div class="encoding-buttons">
                <button v-for="enc in encodings" 
                        :key="enc.value"
                        @click="toggleEncoding(packet, enc.value)"
                        :class="{ active: packetEncodings[packet.timestamp] === enc.value }"
                        class="encoding-button">
                  {{ enc.label }}
                </button>
              </div>
            </div>
            <div class="payload-content">{{ packet.payload }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.capture-container {
  padding: 20px;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.control-panel {
  flex-shrink: 0;
  margin-bottom: 0;
}

.input-group {
  display: flex;
  gap: 10px;
  align-items: center;
}

.input {
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  flex: 1;
}

.button {
  padding: 8px 16px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  white-space: nowrap;
}

.button:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}

.button.stop {
  background-color: #f44336;
}

.button.clear {
  background-color: #9e9e9e;
}

.stats-panel {
  flex-shrink: 0;
  display: flex;
  justify-content: space-around;
  padding: 15px;
  background-color: #f5f5f5;
  border-radius: 4px;
  margin-bottom: 0;
}

.stat-item {
  text-align: center;
}

.label {
  font-weight: 500;
  color: #666;
  margin-right: 8px;
}

.value {
  color: #2196F3;
  font-weight: 600;
}

.packet-panel {
  height: 500px;
  flex-shrink: 0;
  background-color: #fff;
  border-radius: 4px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  display: flex;
  flex-direction: column;
}

.panel-header {
  flex-shrink: 0;
  padding: 15px 20px;
  border-bottom: 1px solid #eee;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #fff;
  border-radius: 4px 4px 0 0;
}

.panel-header h3 {
  margin: 0;
  color: #333;
}

.clear-button {
  padding: 4px 8px;
  background-color: #f5f5f5;
  border: 1px solid #ddd;
  border-radius: 4px;
  color: #666;
  cursor: pointer;
  font-size: 12px;
}

.clear-button:hover {
  background-color: #e0e0e0;
}

.packet-list {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
  background-color: #f8f9fa;
}

.packet-item {
  margin-bottom: 4px;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  transition: all 0.2s ease;
  background-color: #ffffff;
}

.packet-item:hover {
  border-color: #2196F3;
  box-shadow: 0 2px 4px rgba(33, 150, 243, 0.1);
}

.packet-item.expanded {
  border-color: #2196F3;
  background-color: #fff;
}

.packet-header {
  display: flex;
  cursor: pointer;
  padding: 8px 12px;
  transition: all 0.2s;
  border-radius: 4px;
}

.packet-header:hover {
  background-color: #e3f2fd;
}

.packet-time {
  color: #666;
  width: 100px;
  flex-shrink: 0;
  font-family: monospace;
}

.packet-content {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.protocol {
  padding: 2px 8px;
  border-radius: 3px;
  font-weight: 500;
  min-width: 50px;
  text-align: center;
  font-size: 12px;
}

.protocol.tcp {
  color: #fff;
  background-color: #2196F3;
}

.protocol.udp {
  color: #fff;
  background-color: #4CAF50;
}

.protocol.icmp {
  color: #fff;
  background-color: #FF9800;
}

.address {
  color: #333;
  font-family: monospace;
  background-color: #f5f5f5;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 12px;
}

.arrow {
  color: #999;
  font-weight: 300;
}

.length {
  color: #666;
  font-size: 0.9em;
  background-color: #f5f5f5;
  padding: 2px 6px;
  border-radius: 3px;
}

.info {
  color: #4CAF50;
  font-size: 0.9em;
  margin-left: auto;
  background-color: #e8f5e9;
  padding: 2px 6px;
  border-radius: 3px;
}

.packet-payload {
  padding: 8px 12px 8px 110px;
  background-color: #f8f9fa;
  border-top: 1px solid #e3f2fd;
  font-family: monospace;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
}

.payload-content {
  padding: 12px;
  background-color: #fff;
  border: 1px solid #e3f2fd;
  border-radius: 4px;
  max-height: 200px;
  overflow-y: auto;
  line-height: 1.5;
  color: #333;
  box-shadow: inset 0 1px 3px rgba(0,0,0,0.05);
}

.filter-panel {
  margin-top: 15px;
  background-color: #f8f9fa;
  border-radius: 4px;
  padding: 15px;
}

.filter-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.filter-header h4 {
  margin: 0;
  color: #333;
  font-size: 14px;
}

.add-button {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 6px 12px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.plus-icon {
  font-size: 16px;
  font-weight: bold;
}

.filter-rules {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.filter-rule {
  display: flex;
  gap: 10px;
  align-items: center;
  background-color: white;
  padding: 8px;
  border-radius: 4px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.filter-select {
  padding: 6px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 13px;
  min-width: 120px;
}

.filter-input {
  padding: 6px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 13px;
  flex: 1;
}

.remove-button {
  padding: 4px 8px;
  background-color: #f44336;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  line-height: 1;
}

.button-group {
  display: flex;
  gap: 10px;
  margin-top: 15px;
}

.expand-icon {
  cursor: pointer;
  margin-left: auto;
  color: #666;
  transition: transform 0.2s;
  width: 16px;
  height: 16px;
  text-align: center;
  line-height: 16px;
}

.expand-icon.expanded {
  transform: rotate(180deg);
}

.encoding-controls {
  margin-bottom: 10px;
  padding: 8px;
  background-color: #f5f5f5;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.encoding-label {
  font-weight: 500;
  color: #666;
  font-size: 12px;
}

.encoding-buttons {
  display: flex;
  gap: 5px;
  flex-wrap: wrap;
}

.encoding-button {
  padding: 4px 8px;
  font-size: 12px;
  border: 1px solid #ddd;
  border-radius: 3px;
  background-color: #fff;
  color: #666;
  cursor: pointer;
  transition: all 0.2s;
}

.encoding-button:hover {
  background-color: #e3f2fd;
  border-color: #2196F3;
  color: #2196F3;
}

.encoding-button.active {
  background-color: #2196F3;
  border-color: #2196F3;
  color: #fff;
}
</style> 