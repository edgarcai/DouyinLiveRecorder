<script setup>
import {reactive, onMounted, ref} from 'vue'
import {GetConfig} from '../wailsjs/go/main/App.js'

const config = reactive({
  RecordingSettings: {},
  PushSettings: {},
  Cookies: {},
  Authorization: {},
  Accounts: {}
})

const loading = reactive({value: true})
const activeTab = ref('general')

onMounted(() => {
  GetConfig().then(result => {
    if (result) {
      Object.assign(config, result)
    }
    loading.value = false
  })
})
</script>

<template>
  <div class="config-panel">
    <h2>{{ $t('config.title') }}</h2>
    <div v-if="loading.value" class="loading">Loading...</div>
    <div v-else class="config-content">
      
      <div class="tabs">
        <button :class="{active: activeTab === 'general'}" @click="activeTab = 'general'">{{ $t('config.tabs.general') }}</button>
        <button :class="{active: activeTab === 'cookies'}" @click="activeTab = 'cookies'">{{ $t('config.tabs.cookies') }}</button>
        <button :class="{active: activeTab === 'push'}" @click="activeTab = 'push'">{{ $t('config.tabs.push') }}</button>
        <button :class="{active: activeTab === 'accounts'}" @click="activeTab = 'accounts'">{{ $t('config.tabs.accounts') }}</button>
      </div>

      <!-- General Settings -->
      <div v-if="activeTab === 'general'" class="card">
        <h3>{{ $t('config.tabs.general') }}</h3>
        <div class="form-grid">
          <div class="form-group">
            <label>{{ $t('config.general.save_path') }}</label>
            <input v-model="config.RecordingSettings.SavePath" placeholder="Default: ./downloads" />
          </div>
          <div class="form-group">
            <label>{{ $t('config.general.video_quality') }}</label>
            <select v-model="config.RecordingSettings.VideoQuality">
              <option>原画</option>
              <option>蓝光</option>
              <option>超清</option>
              <option>高清</option>
              <option>标清</option>
              <option>流畅</option>
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
            <label>Split Recording</label>
            <select v-model="config.RecordingSettings.SplitRecording">
              <option>是</option>
              <option>否</option>
            </select>
          </div>
          <div class="form-group">
            <label>Split Duration (s)</label>
            <input type="number" v-model="config.RecordingSettings.SplitDuration" />
          </div>
          <div class="form-group">
            <label>Generate Subtitles</label>
            <select v-model="config.RecordingSettings.GenerateTimeSubtitle">
              <option>是</option>
              <option>否</option>
            </select>
          </div>
          <div class="form-group">
            <label>{{ $t('config.general.loop_interval') }}</label>
            <input type="number" v-model="config.RecordingSettings.LoopInterval" />
          </div>
          <div class="form-group">
            <label>{{ $t('config.general.max_threads') }}</label>
            <input type="number" v-model="config.RecordingSettings.MaxThreads" />
          </div>
          <div class="form-group">
            <label>{{ $t('config.general.use_proxy') }}</label>
            <select v-model="config.RecordingSettings.UseProxy">
              <option>是</option>
              <option>否</option>
            </select>
          </div>
          <div class="form-group full-width">
            <label>{{ $t('config.general.proxy_addr') }}</label>
            <input v-model="config.RecordingSettings.ProxyAddress" placeholder="http://127.0.0.1:7890" />
          </div>
        </div>
      </div>
      
      <!-- Cookies -->
      <div v-if="activeTab === 'cookies'" class="card">
        <h3>{{ $t('config.tabs.cookies') }}</h3>
        <div class="form-grid">
          <div class="form-group" v-for="(value, key) in config.Cookies" :key="key">
            <label>{{ key }}</label>
            <textarea v-model="config.Cookies[key]" placeholder="Paste cookie here..."></textarea>
          </div>
        </div>
      </div>

      <!-- Push Settings -->
      <div v-if="activeTab === 'push'" class="card">
        <h3>{{ $t('config.tabs.push') }}</h3>
        <div class="form-grid">
          <div class="form-group">
            <label>{{ $t('config.push.channels') }}</label>
            <input v-model="config.PushSettings.PushChannels" placeholder="dingtalk,wechat,bark" />
          </div>
          <div class="form-group">
            <label>{{ $t('config.push.on_start') }}</label>
            <select v-model="config.PushSettings.PushOnStart">
              <option>是</option>
              <option>否</option>
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
      <div v-if="activeTab === 'accounts'" class="card">
        <h3>{{ $t('config.tabs.accounts') }}</h3>
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

    </div>
  </div>
</template>

<style scoped>
.config-panel {
  max-width: 1200px;
  margin: 0 auto;
}

.tabs {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
  border-bottom: 1px solid #eee;
  padding-bottom: 10px;
}

.tabs button {
  padding: 8px 16px;
  border: none;
  background: none;
  cursor: pointer;
  font-weight: 500;
  color: #666;
  border-radius: 4px;
  transition: all 0.2s;
}

.tabs button.active {
  background-color: #fff0f2;
  color: var(--primary-color);
  font-weight: bold;
}

.tabs button:hover:not(.active) {
  background-color: #f5f5f5;
}

.card {
  background: white;
  padding: 25px;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.02);
  margin-bottom: 25px;
  border: 1px solid #f0f0f0;
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

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.form-group {
  margin-bottom: 15px;
}

.form-group.full-width {
  grid-column: 1 / -1;
}

label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #606266;
  font-size: 14px;
}

input, textarea, select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  font-size: 14px;
  transition: all 0.2s;
  background: #fff;
  color: #333;
  box-sizing: border-box;
}

input:focus, textarea:focus, select:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 2px rgba(255, 44, 85, 0.1);
}

textarea {
  height: 80px;
  resize: vertical;
  font-family: monospace;
}

.loading {
  text-align: center;
  padding: 40px;
  color: #999;
}
</style>
