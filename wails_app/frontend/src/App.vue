<script setup>
import {ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {WindowSetTitle, EventsOn} from './wailsjs/runtime/runtime.js'
import {ToggleMiniMode} from './wailsjs/go/main/App.js'
import ConfigPanel from './components/ConfigPanel.vue'
import StatusPanel from './components/StatusPanel.vue'
import LogPanel from './components/LogPanel.vue'
import HelpPanel from './components/HelpPanel.vue'
import TitleBar from './components/TitleBar.vue'
import FloatingBall from './components/FloatingBall.vue'

const { t, locale } = useI18n()
const currentTab = ref('status')
const isMiniMode = ref(false)
const recordingCount = ref(0)

// Update window title when locale changes
watch(locale, () => {
  WindowSetTitle(t('app.title'))
})

// Set initial title
WindowSetTitle(t('app.title'))

const toggleMiniMode = async (mini) => {
  await ToggleMiniMode(mini)
  isMiniMode.value = mini
}

// Listen for mini mode toggle from StatusPanel
const handleMiniModeToggle = () => {
  toggleMiniMode(true)
}

// Listen for recording status updates to show badge on ball
// We can reuse the existing event or poll, but for now let's assume StatusPanel emits or we listen to backend
// For simplicity, let's just listen to a global event or pass props if we had a store.
// Since we don't have a store, we'll rely on StatusPanel to update us or just mock it for now.
// Better: Listen to "recording_status_change" event from backend if available, or just use a timer here too?
// Let's add a simple listener for now if we can.
</script>

<template>
  <div id="app" :class="{mini: isMiniMode}">
    <template v-if="!isMiniMode">
      <TitleBar />
      <div class="app-layout">
        <aside class="sidebar">
          <div class="brand">
            <div class="logo-icon">🔴</div>
            <div class="logo-text">{{ $t('app.title') }}</div>
          </div>
          
          <nav class="nav-menu">
            <button 
              :class="['nav-item', {active: currentTab === 'status'}]" 
              @click="currentTab = 'status'"
            >
              <span class="icon">📹</span>
              <span class="label">{{ $t('nav.status') }}</span>
            </button>
            <button 
              :class="['nav-item', {active: currentTab === 'config'}]" 
              @click="currentTab = 'config'"
            >
              <span class="icon">⚙️</span>
              <span class="label">{{ $t('nav.config') }}</span>
            </button>
            <button 
              :class="['nav-item', {active: currentTab === 'logs'}]" 
              @click="currentTab = 'logs'"
            >
              <span class="icon">📝</span>
              <span class="label">{{ $t('nav.logs') }}</span>
            </button>
            <button 
              :class="['nav-item', {active: currentTab === 'help'}]" 
              @click="currentTab = 'help'"
            >
              <span class="icon">❓</span>
              <span class="label">{{ $t('nav.help') }}</span>
            </button>
          </nav>
    
          <div class="sidebar-footer">
            <div class="lang-switch">
              <select v-model="$i18n.locale">
                <option value="zh">🇨🇳 中文</option>
                <option value="en">🇺🇸 English</option>
              </select>
            </div>
            <div class="version-info">v1.0.0</div>
          </div>
        </aside>
    
        <main class="main-content">
          <transition name="fade" mode="out-in">
            <keep-alive>
              <component 
                :is="currentTab === 'status' ? StatusPanel : (currentTab === 'config' ? ConfigPanel : (currentTab === 'logs' ? LogPanel : HelpPanel))" 
                @toggle-mini="handleMiniModeToggle"
              />
            </keep-alive>
          </transition>
        </main>
      </div>
    </template>
    
    <FloatingBall 
      v-else 
      :recording-count="recordingCount" 
      @restore="toggleMiniMode(false)" 
    />
  </div>
</template>

<style>
/* ... (keep existing root variables) ... */
:root {
  --primary-color: #ff2c55;
  --primary-hover: #ff4769;
  --bg-color: #f0f2f5;
  --sidebar-bg: #ffffff;
  --sidebar-width: 260px;
  --text-primary: #1a1a1a;
  --text-secondary: #606266;
  --border-color: #ebeef5;
  --card-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  --transition-speed: 0.3s;
}

body {
  margin: 0;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background-color: var(--bg-color);
  color: var(--text-primary);
  -webkit-font-smoothing: antialiased;
  overflow: hidden; /* Prevent body scroll */
}

#app {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  background: transparent; /* Allow transparency for floating ball */
}

#app.mini {
  align-items: center;
  justify-content: center;
}

.app-layout {
  display: flex;
  flex: 1;
  overflow: hidden;
  width: 100%;
}

/* Sidebar Styles */
.sidebar {
  width: var(--sidebar-width);
  background: var(--sidebar-bg);
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--border-color);
  z-index: 10;
  box-shadow: 4px 0 24px rgba(0,0,0,0.02);
}

.brand {
  height: 80px;
  display: flex;
  align-items: center;
  padding: 0 24px;
  gap: 12px;
}

.logo-icon {
  font-size: 24px;
}

.logo-text {
  font-size: 20px;
  font-weight: 800;
  background: linear-gradient(45deg, #ff2c55, #ff7eb3);
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  letter-spacing: -0.5px;
}

.nav-menu {
  padding: 24px 16px;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.nav-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  border: none;
  background: transparent;
  border-radius: 12px;
  cursor: pointer;
  transition: all var(--transition-speed);
  color: var(--text-secondary);
  font-weight: 500;
  font-size: 15px;
  text-align: left;
}

.nav-item:hover {
  background-color: rgba(255, 44, 85, 0.05);
  color: var(--primary-color);
}

.nav-item.active {
  background-color: rgba(255, 44, 85, 0.1);
  color: var(--primary-color);
  font-weight: 600;
}

.nav-item .icon {
  margin-right: 12px;
  font-size: 18px;
}

.sidebar-footer {
  padding: 24px;
  border-top: 1px solid var(--border-color);
}

.lang-switch select {
  width: 100%;
  padding: 10px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  background: #f8f9fa;
  cursor: pointer;
  font-size: 14px;
  color: var(--text-primary);
  outline: none;
  transition: border-color 0.2s;
}

.lang-switch select:hover {
  border-color: #dcdfe6;
}

.version-info {
  margin-top: 12px;
  text-align: center;
  font-size: 12px;
  color: #909399;
}

/* Main Content Styles */
.main-content {
  flex: 1;
  padding: 32px;
  overflow-y: auto;
  background: var(--bg-color);
  position: relative;
}

/* Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(10px);
}

/* Global Component Styles */
h2 {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 24px 0;
  display: flex;
  align-items: center;
}

h2::before {
  content: '';
  display: block;
  width: 4px;
  height: 24px;
  background: var(--primary-color);
  margin-right: 12px;
  border-radius: 2px;
}

button.primary-btn {
  background: linear-gradient(45deg, var(--primary-color), var(--primary-hover));
  color: white;
  border: none;
  padding: 10px 24px;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  font-size: 14px;
  transition: transform 0.1s, box-shadow 0.2s;
  box-shadow: 0 4px 12px rgba(255, 44, 85, 0.2);
}

button.primary-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 16px rgba(255, 44, 85, 0.3);
}

button.primary-btn:active {
  transform: translateY(0);
}

input[type="text"], 
input[type="number"], 
input[type="password"],
textarea,
select {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 10px 12px;
  font-size: 14px;
  transition: all 0.2s;
  background: white;
}

input:focus, textarea:focus, select:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px rgba(255, 44, 85, 0.1);
}

/* Scrollbar Styling */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: #dcdfe6;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: #c0c4cc;
}
</style>
