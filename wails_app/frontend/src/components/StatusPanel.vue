<script setup>
import {reactive, onMounted, onUnmounted} from 'vue'
import {StartRecording, StopRecording, GetRecordingStatus} from '../../wailsjs/go/main/App'

const state = reactive({
  url: '',
  statusMap: {},
  message: ''
})

let timer = null

const updateStatus = () => {
  GetRecordingStatus().then(result => {
    state.statusMap = result
  })
}

const start = () => {
  if (!state.url) return
  StartRecording(state.url).then(result => {
    state.message = result
    updateStatus()
  })
}

const stop = (url) => {
  StopRecording(url).then(result => {
    state.message = result
    updateStatus()
  })
}

const getStatusClass = (status) => {
  if (status.includes('Recording')) return 'badge-recording'
  if (status.includes('Monitoring')) return 'badge-monitoring'
  if (status.includes('Stopped') || status.includes('Error')) return 'badge-stopped'
  return 'badge-starting'
}

onMounted(() => {
  updateStatus()
  timer = setInterval(updateStatus, 2000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <div class="status-panel">
    <h2>{{ $t('status.title') }}</h2>
    <div class="control-bar card">
      <input v-model="state.url" :placeholder="$t('status.placeholder')" type="text" />
      <button @click="start" class="primary-btn">{{ $t('status.start') }}</button>
    </div>
    <div v-if="state.message" class="message-box" :class="{error: state.message.includes('Error')}">
      {{ state.message }}
    </div>
    
    <div class="status-list card">
      <h3>{{ $t('status.active') }}</h3>
      <div v-if="Object.keys(state.statusMap).length === 0" class="empty-state">
        {{ $t('status.empty') }}
      </div>
      <ul v-else>
        <li v-for="(status, url) in state.statusMap" :key="url">
          <div class="info">
            <span class="url" :title="url">{{ url }}</span>
            <span class="badge" :class="getStatusClass(status)">{{ status }}</span>
          </div>
          <button @click="stop(url)" class="stop-btn">{{ $t('status.stop') }}</button>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
.card {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  margin-bottom: 20px;
}

.control-bar {
  display: flex;
  gap: 15px;
}

input {
  flex: 1;
}

.message-box {
  padding: 10px;
  background: #e3f2fd;
  color: #0d47a1;
  border-radius: 6px;
  margin-bottom: 20px;
}

.message-box.error {
  background: #ffebee;
  color: #c62828;
}

.empty-state {
  color: #999;
  text-align: center;
  padding: 40px;
  font-style: italic;
}

h3 {
  margin-top: 0;
  margin-bottom: 20px;
  color: #2c3e50;
  font-size: 16px;
  font-weight: 600;
  border-left: 4px solid var(--primary-color);
  padding-left: 10px;
}

.status-list ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.status-list li {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 0;
  border-bottom: 1px solid #eee;
}

.status-list li:last-child {
  border-bottom: none;
}

.info {
  display: flex;
  flex-direction: column;
  gap: 5px;
  overflow: hidden;
}

.url {
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 500px;
}

.badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: bold;
  width: fit-content;
}

.badge-recording {
  background-color: #e8f5e9;
  color: #2e7d32;
}

.badge-monitoring {
  background-color: #e3f2fd;
  color: #1565c0;
}

.badge-stopped {
  background-color: #ffebee;
  color: #c62828;
}

.badge-starting {
  background-color: #fff3e0;
  color: #ef6c00;
}

.stop-btn {
  padding: 6px 16px;
  border: 1px solid #ffcdd2;
  background: #fff;
  border-radius: 6px;
  cursor: pointer;
  color: #c62828;
  font-size: 13px;
  transition: all 0.2s;
}

.stop-btn:hover {
  background: #ffebee;
  border-color: #ef9a9a;
}
</style>
