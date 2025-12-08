<script setup>
import { ref, onMounted, watch } from 'vue'
import { CheckAppUpdate, StartAppUpdate } from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'

const props = defineProps({
  show: Boolean
})

const emit = defineEmits(['close'])

const checking = ref(false)
const hasUpdate = ref(false)
const newVersion = ref('')
const releaseNotes = ref('')
const downloadUrl = ref('')
const error = ref('')

const downloading = ref(false)
const progress = ref(0)
const downloadComplete = ref(false)
const downloadedFile = ref('')

const check = () => {
  checking.value = true
  error.value = ''
  hasUpdate.value = false
  downloadComplete.value = false
  
  CheckAppUpdate().then(res => {
    checking.value = false
    if (res.error) {
      error.value = res.error
      return
    }
    hasUpdate.value = res.hasUpdate
    if (hasUpdate.value) {
      newVersion.value = res.newVersion
      releaseNotes.value = res.releaseNotes
      downloadUrl.value = res.downloadUrl
    }
  })
}

const startDownload = () => {
  downloading.value = true
  StartAppUpdate(downloadUrl.value)
}

watch(() => props.show, (newVal) => {
  if (newVal) {
    check()
  }
})

onMounted(() => {
  EventsOn("update-progress", (p) => {
    progress.value = p
  })

  EventsOn("update-complete", (path) => {
    downloading.value = false
    downloadComplete.value = true
    downloadedFile.value = path
  })

  EventsOn("update-error", (err) => {
    downloading.value = false
    error.value = err
  })
})

const close = () => {
  emit('close')
}
</script>

<template>
  <div v-if="show" class="modal-overlay">
    <div class="modal-content">
      <div class="modal-header">
        <h3>{{ $t('update.title') }}</h3>
        <button @click="close" class="close-btn">×</button>
      </div>
      
      <div class="modal-body">
        <div v-if="checking" class="loading">
          <div class="spinner"></div>
          <p>{{ $t('update.checking') }}</p>
        </div>

        <div v-else-if="error" class="error-state">
          <p class="error-text">{{ error }}</p>
          <button @click="check" class="retry-btn">{{ $t('update.retry') }}</button>
        </div>

        <div v-else-if="hasUpdate && !downloadComplete" class="update-info">
          <div class="version-badge">New: {{ newVersion }}</div>
          <div class="release-notes">
            <h4>{{ $t('update.release_notes') }}</h4>
            <pre>{{ releaseNotes }}</pre>
          </div>
          
          <div v-if="downloading" class="progress-section">
            <div class="progress-bar">
              <div class="fill" :style="{width: progress + '%'}"></div>
            </div>
            <p>{{ progress }}%</p>
          </div>
          
          <div v-else class="actions">
            <button @click="startDownload" class="download-btn">{{ $t('update.download') }}</button>
            <button @click="close" class="cancel-btn">{{ $t('update.later') }}</button>
          </div>
        </div>

        <div v-else-if="downloadComplete" class="complete-state">
          <p class="success-icon">✅</p>
          <p>{{ $t('update.downloaded') }}</p>
          <p class="file-path">{{ downloadedFile }}</p>
          <p class="instruction">{{ $t('update.install_instruction') }}</p>
          <button @click="close" class="close-btn-main">{{ $t('update.close') }}</button>
        </div>

        <div v-else class="no-update">
          <p>{{ $t('update.latest_version') }}</p>
          <button @click="close" class="close-btn-main">{{ $t('update.close') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 500px;
  max-width: 90%;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  animation: modal-in 0.3s ease;
}

@keyframes modal-in {
  from { opacity: 0; transform: translateY(-20px); }
  to { opacity: 1; transform: translateY(0); }
}

.modal-header {
  padding: 16px 24px;
  border-bottom: 1px solid #eee;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  color: var(--text-primary);
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #999;
}

.close-btn:hover {
  color: #666;
}

.modal-body {
  padding: 24px;
}

.loading, .error-state, .no-update, .complete-state {
  text-align: center;
  padding: 20px;
  color: var(--text-secondary);
}

.spinner {
  border: 3px solid #f3f3f3;
  border-top: 3px solid var(--primary-color);
  border-radius: 50%;
  width: 30px;
  height: 30px;
  animation: spin 1s linear infinite;
  margin: 0 auto 16px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.error-text {
  color: #f56c6c;
  margin-bottom: 16px;
}

.version-badge {
  display: inline-block;
  background: #e6f7ff;
  color: #1890ff;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 14px;
  margin-bottom: 16px;
  font-weight: 500;
}

.release-notes h4 {
  margin: 0 0 8px 0;
  font-size: 14px;
  color: var(--text-primary);
}

.release-notes pre {
  background: #f5f7fa;
  padding: 12px;
  border-radius: 6px;
  white-space: pre-wrap;
  max-height: 200px;
  overflow-y: auto;
  font-family: inherit;
  font-size: 13px;
  color: var(--text-secondary);
  border: 1px solid #eee;
}

.progress-section {
  margin-top: 24px;
  text-align: center;
}

.progress-bar {
  height: 8px;
  background: #eee;
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 8px;
}

.fill {
  height: 100%;
  background: var(--primary-color);
  transition: width 0.3s;
}

.actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 24px;
}

.download-btn, .retry-btn, .close-btn-main {
  background: var(--primary-color);
  color: white;
  border: none;
  padding: 8px 20px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: opacity 0.2s;
}

.download-btn:hover, .retry-btn:hover, .close-btn-main:hover {
  opacity: 0.9;
}

.cancel-btn {
  background: white;
  border: 1px solid #dcdfe6;
  padding: 8px 20px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  color: #606266;
  transition: all 0.2s;
}

.cancel-btn:hover {
  border-color: #c0c4cc;
  color: #303133;
}

.success-icon {
  font-size: 48px;
  margin: 0 0 16px 0;
}

.file-path {
  background: #f5f7fa;
  padding: 8px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 12px;
  word-break: break-all;
  margin: 8px 0;
}

.instruction {
  font-size: 13px;
  color: #e6a23c;
  margin-bottom: 24px;
}
</style>
