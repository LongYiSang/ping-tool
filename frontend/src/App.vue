<script>
import PingMonitor from './components/PingMonitor.vue'
import TCPTester from './components/TCPTester.vue'
import PacketCapture from './components/PacketCapture.vue'
import { Greet } from '../wailsjs/go/backend/App'

export default {
  components: {
    PingMonitor,
    TCPTester,
    PacketCapture
  },
  data() {
    return {
      activeTab: 'ping'
    }
  }
}
</script>

<template>
  <div class="container">
    <div class="tabs">
      <button 
        @click="activeTab = 'ping'"
        :class="{ active: activeTab === 'ping' }"
        class="tab-button"
      >
        Ping监控
      </button>
      <button 
        @click="activeTab = 'tcp'"
        :class="{ active: activeTab === 'tcp' }"
        class="tab-button"
      >
        TCP测试
      </button>
      <button 
        @click="activeTab = 'capture'"
        :class="{ active: activeTab === 'capture' }"
        class="tab-button"
      >
        抓包分析
      </button>
    </div>

    <div v-if="activeTab === 'ping'">
      <PingMonitor />
    </div>

    <div v-if="activeTab === 'tcp'">
      <TCPTester />
    </div>

    <div v-if="activeTab === 'capture'">
      <PacketCapture />
    </div>
  </div>
</template>

<style>
.container {
  padding: 20px;
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.tabs {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.tab-button {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  background-color: #f0f0f0;
  cursor: pointer;
  font-size: 14px;
}

.tab-button.active {
  background-color: #4CAF50;
  color: white;
}
</style> 