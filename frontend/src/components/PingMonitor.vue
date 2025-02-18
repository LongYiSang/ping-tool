<script>
import { StartPing, StopPing, GetPingResults } from '../../wailsjs/go/backend/App'
import * as echarts from 'echarts'

export default {
  data() {
    return {
      target: '',
      interval: 1000,
      isPinging: false,
      chart: null,
      stats: null,
      updateTimer: null,
      chartData: {
        times: [],
        rtts: []
      },
      pingLogs: [],
      maxLogs: 100
    }
  },
  mounted() {
    this.initChart()
    window.addEventListener('resize', this.resizeChart)
  },
  beforeDestroy() {
    if (this.updateTimer) {
      clearInterval(this.updateTimer)
    }
    window.removeEventListener('resize', this.resizeChart)
  },
  methods: {
    initChart() {
      this.chart = echarts.init(this.$refs.chartContainer)
      const option = {
        title: {
          text: 'Ping监控图表',
          left: 'center'
        },
        tooltip: {
          trigger: 'axis',
          formatter: function(params) {
            const data = params[0]
            return `时间: ${data.name}<br/>延迟: ${data.value !== null ? data.value.toFixed(2) : '超时'}ms`
          }
        },
        grid: {
          top: 60,
          right: 20,
          bottom: 60,
          left: 50
        },
        xAxis: {
          type: 'category',
          data: [],
          axisLabel: {
            rotate: 45,
            fontSize: 11,
            color: '#666'
          }
        },
        yAxis: {
          type: 'value',
          name: '延迟(ms)',
          nameTextStyle: {
            color: '#666'
          },
          min: 0,
          splitLine: {
            show: true,
            lineStyle: {
              type: 'dashed',
              color: '#ddd'
            }
          }
        },
        series: [{
          data: [],
          type: 'line',
          smooth: false,
          name: '延迟',
          showSymbol: true,
          symbol: 'emptyCircle',
          symbolSize: 8,
          itemStyle: {
            color: '#4CAF50',
            borderWidth: 2
          },
          lineStyle: {
            width: 3,
            color: '#4CAF50'
          },
          emphasis: {
            itemStyle: {
              color: '#4CAF50',
              borderColor: '#4CAF50',
              borderWidth: 3,
              shadowColor: 'rgba(0, 0, 0, 0.2)',
              shadowBlur: 10
            }
          },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              {
                offset: 0,
                color: 'rgba(76, 175, 80, 0.3)'
              },
              {
                offset: 1,
                color: 'rgba(76, 175, 80, 0.1)'
              }
            ])
          }
        }]
      }
      this.chart.setOption(option)
    },
    resizeChart() {
      if (this.chart) {
        this.chart.resize()
      }
    },
    async startPing() {
      if (!this.target || !this.interval || this.interval < 100) {
        alert('请输入有效的目标地址和间隔时间（最小100ms）')
        return
      }

      try {
        await StartPing(this.target, this.interval)
        this.isPinging = true
        this.updateTimer = setInterval(this.updateChart, 1000)
      } catch (err) {
        alert('启动Ping监控失败: ' + err)
      }
    },
    async stopPing() {
      try {
        await StopPing(this.target)
        this.isPinging = false
        if (this.updateTimer) {
          clearInterval(this.updateTimer)
          this.updateTimer = null
        }
        this.pingLogs = []
      } catch (err) {
        alert('停止Ping监控失败: ' + err)
      }
    },
    async updateChart() {
      try {
        const results = await GetPingResults(this.target)
        if (!results || results.length === 0) return

        // 最多显示30个数据点
        const maxDataPoints = 30
        const displayResults = results.slice(-maxDataPoints)

        this.chartData.times = displayResults.map(r => new Date(r.Timestamp).toLocaleTimeString())
        this.chartData.rtts = displayResults.map(r => r.Success ? r.RTT / 1000000 : null)

        // 更新统计信息
        const successResults = this.chartData.rtts.filter(rtt => rtt !== null)
        if (successResults.length > 0) {
          this.stats = {
            min: Math.min(...successResults).toFixed(2),
            max: Math.max(...successResults).toFixed(2),
            avg: (successResults.reduce((a, b) => a + b, 0) / successResults.length).toFixed(2),
            lossRate: ((this.chartData.rtts.length - successResults.length) / this.chartData.rtts.length * 100).toFixed(2)
          }
        }

        // 更新日志
        const latestResult = results[results.length - 1]
        if (latestResult) {
          const logEntry = {
            timestamp: latestResult.Timestamp,
            target: this.target,
            ip: latestResult.IP,
            success: latestResult.Success,
            rtt: latestResult.Success ? latestResult.RTT / 1000000 : null,
            error: latestResult.Error
          }
          this.pingLogs.unshift(logEntry)
          if (this.pingLogs.length > this.maxLogs) {
            this.pingLogs = this.pingLogs.slice(0, this.maxLogs)
          }
        }

        // 更新图表
        this.chart.setOption({
          xAxis: {
            data: this.chartData.times
          },
          series: [{
            data: this.chartData.rtts.map(rtt => rtt !== null ? parseFloat(rtt.toFixed(2)) : null)
          }]
        })
      } catch (err) {
        console.error('更新图表失败:', err)
      }
    },
    clearLogs() {
      this.pingLogs = []
    }
  }
}
</script>

<template>
  <div class="ping-container">
    <div class="control-panel">
      <div class="input-group">
        <input v-model="target" placeholder="输入IP地址或域名" class="input" />
        <input v-model.number="interval" type="number" placeholder="Ping间隔(ms)" class="input" min="100" />
        <button @click="startPing" :disabled="isPinging" class="button">开始监控</button>
        <button @click="stopPing" :disabled="!isPinging" class="button">停止监控</button>
      </div>
    </div>
    <div class="chart-container" ref="chartContainer"></div>
    <div class="stats-panel" v-if="stats">
      <div class="stat-item">
        <span class="label">最小延迟:</span>
        <span class="value">{{ stats.min }}ms</span>
      </div>
      <div class="stat-item">
        <span class="label">最大延迟:</span>
        <span class="value">{{ stats.max }}ms</span>
      </div>
      <div class="stat-item">
        <span class="label">平均延迟:</span>
        <span class="value">{{ stats.avg }}ms</span>
      </div>
      <div class="stat-item">
        <span class="label">丢包率:</span>
        <span class="value">{{ stats.lossRate }}%</span>
      </div>
    </div>
    <div class="ping-log-panel" v-if="pingLogs.length > 0">
      <div class="panel-header">
        <h3>Ping日志</h3>
        <button @click="clearLogs" class="clear-button">清除日志</button>
      </div>
      <div class="log-list">
        <div v-for="(log, index) in pingLogs" :key="index" 
             class="log-item" :class="{ 'log-success': log.success, 'log-error': !log.success }">
          <div class="log-time">{{ new Date(log.timestamp).toLocaleTimeString() }}</div>
          <div class="log-content">
            <template v-if="log.success">
              <span class="log-target">{{ log.target }}</span>
              <span class="log-ip">[{{ log.ip }}]</span>
              <span class="log-rtt">{{ log.rtt.toFixed(2) }}ms</span>
            </template>
            <template v-else>
              <span class="log-target">{{ log.target }}</span>
              <span class="log-ip">[{{ log.ip }}]</span>
              <span class="log-error-message">{{ log.error }}</span>
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ping-container {
  padding: 20px;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.control-panel {
  flex-shrink: 0;
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
}

.button {
  padding: 8px 16px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.button:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}

.chart-container {
  flex-shrink: 0;
  height: 400px;
}

.stats-panel {
  flex-shrink: 0;
  display: flex;
  justify-content: space-around;
  padding: 20px;
  background-color: #ffffff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  border: 1px solid #e8f5e9;
}

.stat-item {
  text-align: center;
}

.label {
  font-weight: 600;
  color: #333333;
  margin-right: 8px;
}

.value {
  font-weight: 600;
}

.stat-item:nth-child(1) .value {
  color: #2196F3;  /* 最小延迟 - 蓝色 */
}

.stat-item:nth-child(2) .value {
  color: #f44336;  /* 最大延迟 - 红色 */
}

.stat-item:nth-child(3) .value {
  color: #4CAF50;  /* 平均延迟 - 绿色 */
}

.stat-item:nth-child(4) .value {
  color: #ff9800;  /* 丢包率 - 橙色 */
}

.ping-log-panel {
  height: 300px;
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

.log-list {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
}

.log-item {
  display: flex;
  padding: 8px 12px;
  border-bottom: 1px solid #f5f5f5;
  font-size: 13px;
  align-items: center;
}

.log-item:last-child {
  border-bottom: none;
}

.log-time {
  color: #666;
  width: 100px;
  flex-shrink: 0;
}

.log-content {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
}

.log-target {
  color: #333;
  font-weight: 500;
}

.log-rtt {
  color: #4CAF50;
  font-weight: 500;
}

.log-error-message {
  color: #f44336;
}

.log-success {
  background-color: #f9fff9;
}

.log-error {
  background-color: #fff9f9;
}

.log-ip {
  color: #888;
  font-size: 0.9em;
  margin: 0 8px;
}
</style> 