<script setup>
import {reactive, onMounted, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {GetConfig, UpdateConfig} from '../wailsjs/go/main/App.js'

const { t } = useI18n()

const config = reactive({
  RecordingSettings: {},
  PushSettings: {},
  Cookies: {},
  Authorization: {},
  Accounts: {}
})

const loading = reactive({value: true})
const activeTab = ref('general')
const saveStatus = ref('')

const debounce = (fn, delay) => {
  let timeoutId
  return (...args) => {
    clearTimeout(timeoutId)
    timeoutId = setTimeout(() => fn(...args), delay)
  }
}

const saveConfig = debounce(() => {
  saveStatus.value = t('config.saving')
  UpdateConfig(config).then(result => {
    saveStatus.value = result // "Saved" or error message
    setTimeout(() => {
      if (saveStatus.value === 'Saved') {
        saveStatus.value = ''
      }
    }, 2000)
  })
}, 1000)

onMounted(() => {
  GetConfig().then(result => {
    if (result) {
      Object.assign(config, result)
    }
    loading.value = false
    
    // Start watching after initial load
    watch(config, () => {
      saveConfig()
    }, {deep: true})
  })
})
</script>

<template>
  <div class="config-panel">
    <div class="header-row">
      <h2>{{ $t('config.title') }}</h2>
      <transition name="fade">
        <span v-if="saveStatus" class="save-status" :class="{error: saveStatus.includes('Error')}">
          {{ saveStatus === 'Saved' ? '✅ ' + $t('config.saved') : saveStatus }}
        </span>
      </transition>
    </div>
    
    <div v-if="loading.value" class="loading-state">
      <div class="spinner"></div>
      <p>Loading configuration...</p>
    </div>
    
    <div v-else class="config-content">
      <div class="tabs-container">
        <div class="tabs">
          <button 
            v-for="tab in ['general', 'cookies', 'push', 'accounts']" 
            :key="tab"
            :class="{active: activeTab === tab}" 
            @click="activeTab = tab"
          >
            {{ $t(`config.tabs.${tab}`) }}
          </button>
        </div>
      </div>

      <transition name="fade" mode="out-in">
        <!-- General Settings -->
        <div v-if="activeTab === 'general'" class="card settings-card">
          <div class="card-header">
            <h3>{{ $t('config.tabs.general') }}</h3>
            <p class="subtitle">{{ $t('config.subtitle.general') }}</p>
          </div>
          
          <div class="form-grid">
            <div class="form-group full-width">
              <label>{{ $t('config.general.save_path') }}</label>
              <div class="input-wrapper">
                <input v-model="config.RecordingSettings.SavePath" :placeholder="$t('config.placeholder.save_path')" />
                <span class="input-icon">📂</span>
              </div>
            </div>
            
            <div class="form-group">
              <label>{{ $t('config.general.video_quality') }}</label>
              <select v-model="config.RecordingSettings.VideoQuality">
                <option value="原画">{{ $t('config.options.quality.origin') }}</option>
                <option value="蓝光">{{ $t('config.options.quality.blue') }}</option>
                <option value="超清">{{ $t('config.options.quality.super') }}</option>
                <option value="高清">{{ $t('config.options.quality.high') }}</option>
                <option value="标清">{{ $t('config.options.quality.standard') }}</option>
                <option value="流畅">{{ $t('config.options.quality.fluent') }}</option>
              </select>
            </div>
            
            <div class="form-group">
              <label>{{ $t('config.general.video_format') }}</label>
              <select v-model="config.RecordingSettings.VideoFormat">
                <option>ts</option>
                <option>mkv</option>
                <option>flv</option>
                <option>mp4</option>
                <option>mp3</option>
                <option>m4a</option>
              </select>
            </div>
            
            <div class="form-group">
              <label>{{ $t('config.general.split_recording') }}</label>
              <select v-model="config.RecordingSettings.SplitRecording">
                <option value="是">{{ $t('config.options.yes') }}</option>
                <option value="否">{{ $t('config.options.no') }}</option>
              </select>
            </div>
            
            <div class="form-group">
              <label>{{ $t('config.general.split_duration') }}</label>
              <input type="number" v-model.number="config.RecordingSettings.SplitDuration" />
            </div>
            
            <div class="form-group">
              <label>{{ $t('config.general.generate_subtitles') }}</label>
              <select v-model="config.RecordingSettings.GenerateTimeSubtitle">
                <option value="是">{{ $t('config.options.yes') }}</option>
                <option value="否">{{ $t('config.options.no') }}</option>
              </select>
            </div>
            
            <div class="form-group">
              <label>{{ $t('config.general.loop_interval') }}</label>
              <input type="number" v-model.number="config.RecordingSettings.LoopInterval" />
            </div>
            
            <div class="form-group">
              <label>{{ $t('config.general.max_threads') }}</label>
              <input type="number" v-model.number="config.RecordingSettings.MaxThreads" />
            </div>
            
            <div class="form-group">
              <label>{{ $t('config.general.use_proxy') }}</label>
              <select v-model="config.RecordingSettings.UseProxy">
                <option value="是">{{ $t('config.options.yes') }}</option>
                <option value="否">{{ $t('config.options.no') }}</option>
              </select>
            </div>
            
            <div class="form-group full-width">
              <label>{{ $t('config.general.proxy_addr') }}</label>
              <input v-model="config.RecordingSettings.ProxyAddress" :placeholder="$t('config.placeholder.proxy')" />
            </div>
          </div>
        </div>
        
        <!-- Cookies -->
        <div v-else-if="activeTab === 'cookies'" class="card settings-card">
          <div class="card-header">
            <h3>{{ $t('config.tabs.cookies') }}</h3>
            <p class="subtitle">{{ $t('config.subtitle.cookies') }}</p>
          </div>
          <div class="form-grid">
            <div class="form-group full-width" v-for="(value, key) in config.Cookies" :key="key">
              <label>{{ key }}</label>
              <textarea v-model="config.Cookies[key]" :placeholder="$t('config.placeholder.cookie')"></textarea>
            </div>
          </div>
        </div>

        <!-- Push Settings -->
        <div v-else-if="activeTab === 'push'" class="card settings-card">
          <div class="card-header">
            <h3>{{ $t('config.tabs.push') }}</h3>
            <p class="subtitle">{{ $t('config.subtitle.push') }}</p>
          </div>
          <div class="form-grid">
            <div class="form-group full-width">
              <label>{{ $t('config.push.channels') }}</label>
              <input v-model="config.PushSettings.PushChannels" :placeholder="$t('config.placeholder.channels')" />
              <small class="hint">{{ $t('config.hint.channels') }}</small>
            </div>
            <div class="form-group">
              <label>{{ $t('config.push.on_start') }}</label>
              <select v-model="config.PushSettings.PushOnStart">
                <option value="是">{{ $t('config.options.yes') }}</option>
                <option value="否">{{ $t('config.options.no') }}</option>
              </select>
            </div>
            <div class="form-group full-width">
              <label>{{ $t('config.push.dingtalk_url') }}</label>
              <input v-model="config.PushSettings.DingTalkUrl" />
            </div>
            <div class="form-group full-width">
              <label>{{ $t('config.push.bark_url') }}</label>
              <input v-model="config.PushSettings.BarkUrl" />
            </div>
          </div>
        </div>

        <!-- Accounts -->
        <div v-else-if="activeTab === 'accounts'" class="card settings-card">
          <div class="card-header">
            <h3>{{ $t('config.tabs.accounts') }}</h3>
            <p class="subtitle">{{ $t('config.subtitle.accounts') }}</p>
          </div>
          <div class="form-grid">
            <div class="form-group">
              <label>Sooplive Account</label>
              <input v-model="config.Accounts.SoopliveAccount" />
            </div>
            <div class="form-group">
              <label>Sooplive Password</label>
              <input type="password" v-model="config.Accounts.SooplivePassword" />
            </div>
            <div class="form-group">
              <label>Flextv Account</label>
              <input v-model="config.Accounts.FlextvAccount" />
            </div>
            <div class="form-group">
              <label>Flextv Password</label>
              <input type="password" v-model="config.Accounts.FlextvPassword" />
            </div>
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<style scoped>
.config-panel {
  max-width: 1000px;
  margin: 0 auto;
}

.header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-row h2 {
  margin: 0;
}

.save-status {
  font-size: 14px;
  color: #67c23a;
  font-weight: 500;
}

.save-status.error {
  color: #f56c6c;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px;
  color: var(--text-secondary);
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #f3f3f3;
  border-top: 3px solid var(--primary-color);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.tabs-container {
  margin-bottom: 24px;
  border-bottom: 1px solid var(--border-color);
}

.tabs {
  display: flex;
  gap: 24px;
}

.tabs button {
  padding: 12px 0;
  border: none;
  background: none;
  cursor: pointer;
  font-weight: 500;
  color: var(--text-secondary);
  font-size: 15px;
  position: relative;
  transition: color 0.2s;
}

.tabs button:hover {
  color: var(--primary-color);
}

.tabs button.active {
  color: var(--primary-color);
  font-weight: 600;
}

.tabs button.active::after {
  content: '';
  position: absolute;
  bottom: -1px;
  left: 0;
  width: 100%;
  height: 2px;
  background: var(--primary-color);
  border-radius: 2px 2px 0 0;
}

.card {
  background: white;
  border-radius: 12px;
  box-shadow: var(--card-shadow);
  border: 1px solid var(--border-color);
  overflow: hidden;
}

.settings-card {
  padding: 32px;
}

.card-header {
  margin-bottom: 32px;
}

.card-header h3 {
  margin: 0 0 8px 0;
  font-size: 18px;
  color: var(--text-primary);
  border: none;
  padding: 0;
}

.subtitle {
  margin: 0;
  color: var(--text-secondary);
  font-size: 14px;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 24px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group.full-width {
  grid-column: 1 / -1;
}

label {
  font-weight: 500;
  color: var(--text-primary);
  font-size: 14px;
}

.input-wrapper {
  position: relative;
}

.input-icon {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  opacity: 0.5;
  pointer-events: none;
}

input, textarea, select {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 14px;
  transition: all 0.2s;
  background: #fcfcfc;
  color: var(--text-primary);
  box-sizing: border-box;
}

input:hover, textarea:hover, select:hover {
  background: white;
  border-color: #c0c4cc;
}

input:focus, textarea:focus, select:focus {
  outline: none;
  border-color: var(--primary-color);
  background: white;
  box-shadow: 0 0 0 3px rgba(255, 44, 85, 0.1);
}

textarea {
  min-height: 100px;
  resize: vertical;
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
  line-height: 1.5;
}

.hint {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 4px;
}

/* Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
