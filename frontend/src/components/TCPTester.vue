<script>
import { TestTCPConnection } from '../../wailsjs/go/backend/App'

export default {
  data() {
    return {
      tcpHost: '',
      tcpPort: 80,
      tcpTimeout: 2000,
      tcpResult: null,
      isTesting: false
    }
  },
  methods: {
    async testTCP() {
      if (!this.tcpHost || !this.tcpPort) {
        alert('请输入有效的主机地址和端口')
        return
      }

      this.isTesting = true
      try {
        this.tcpResult = await TestTCPConnection(this.tcpHost, this.tcpPort, this.tcpTimeout)
      } catch (err) {
        alert('TCP测试失败: ' + err)
      } finally {
        this.isTesting = false
      }
    }
  }
}
</script>

<template>
  <div class="tcp-container">
    <div class="control-panel">
      <div class="input-group">
        <input v-model="tcpHost" placeholder="输入主机地址" class="input" />
        <input v-model.number="tcpPort" type="number" placeholder="端口" class="input" />
        <input v-model.number="tcpTimeout" type="number" placeholder="超时(ms)" class="input" value="2000" />
        <button @click="testTCP" :disabled="isTesting" class="button">
          {{ isTesting ? '测试中...' : '测试连接' }}
        </button>
      </div>
    </div>
    
    <div class="tcp-results" v-if="tcpResult">
      <div class="result-card" :class="{ success: tcpResult.Success, error: !tcpResult.Success }">
        <h3>测试结果</h3>
        <p><strong>状态:</strong> {{ tcpResult.Success ? '成功' : '失败' }}</p>
        <p><strong>目标地址:</strong> {{ tcpHost }}:{{ tcpPort }}</p>
        <p><strong>解析IP:</strong> {{ tcpResult.IP }}</p>
        <p><strong>连接时间:</strong> {{ (tcpResult.ConnectTime / 1000000).toFixed(2) }}ms</p>
        <p v-if="tcpResult.Error"><strong>错误信息:</strong> {{ tcpResult.Error }}</p>
        <p><strong>测试时间:</strong> {{ new Date(tcpResult.Timestamp).toLocaleString() }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tcp-container {
  padding: 20px;
}

.control-panel {
  margin-bottom: 20px;
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

.tcp-results {
  margin-top: 20px;
}

.result-card {
  padding: 20px;
  border-radius: 4px;
  margin-bottom: 10px;
  background-color: #fff;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.result-card.success {
  background-color: #e8f5e9;
  border: 1px solid #4caf50;
}

.result-card.error {
  background-color: #ffebee;
  border: 1px solid #f44336;
}

.result-card h3 {
  margin-top: 0;
  margin-bottom: 15px;
  color: #333;
}

.result-card p {
  margin: 8px 0;
  color: #666;
}

.result-card strong {
  color: #333;
  margin-right: 8px;
  min-width: 80px;
  display: inline-block;
}
</style> 