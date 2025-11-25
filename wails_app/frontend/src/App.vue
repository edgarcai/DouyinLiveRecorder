<script setup>
import {ref} from 'vue'
import ConfigPanel from './components/ConfigPanel.vue'
import StatusPanel from './components/StatusPanel.vue'
import LogPanel from './components/LogPanel.vue'

const currentTab = ref('status')
</script>

<template>
  <div id="app">
    <div class="sidebar">
      <div class="logo">🔴 LiveRecorder</div>
      <nav>
        <button :class="{active: currentTab === 'status'}" @click="currentTab = 'status'">
          <span class="icon">📹</span> {{ $t('nav.status') }}
        </button>
        <button :class="{active: currentTab === 'config'}" @click="currentTab = 'config'">
          <span class="icon">⚙️</span> {{ $t('nav.config') }}
        </button>
        <button :class="{active: currentTab === 'logs'}" @click="currentTab = 'logs'">
          <span class="icon">📝</span> {{ $t('nav.logs') }}
        </button>
      </nav>
      <div class="lang-switch">
        <select v-model="$i18n.locale">
          <option value="zh">🇨🇳 中文</option>
          <option value="en">🇺🇸 English</option>
        </select>
      </div>
    </div>
    <div class="main-content">
      <transition name="fade" mode="out-in">
        <StatusPanel v-if="currentTab === 'status'" />
        <ConfigPanel v-else-if="currentTab === 'config'" />
        <LogPanel v-else-if="currentTab === 'logs'" />
      </transition>
    </div>
  </div>
</template>

<style>
:root {
  --primary-color: #ff2c55; /* Douyin Red */
  --bg-color: #f8f9fa;
  --sidebar-bg: #ffffff;
  --text-color: #333;
  --border-color: #e1e4e8;
}

body {
  margin: 0;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  background-color: var(--bg-color);
  color: var(--text-color);
}

#app {
  display: flex;
  height: 100vh;
}

.sidebar {
  width: 250px;
  background: var(--sidebar-bg);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
}

.logo {
  padding: 20px;
  font-size: 20px;
  font-weight: bold;
  color: var(--primary-color);
  border-bottom: 1px solid var(--border-color);
}

.lang-switch {
  padding: 20px;
  border-top: 1px solid var(--border-color);
  margin-top: auto;
}

.lang-switch select {
  width: 100%;
  padding: 8px;
  border-radius: 4px;
  border: 1px solid #ddd;
  background: white;
  cursor: pointer;
}

nav {
  padding: 20px 0;
}

nav button {
  display: block;
  width: 100%;
  padding: 15px 20px;
  border: none;
  background: none;
  text-align: left;
  font-size: 16px;
  color: #666;
  cursor: pointer;
  transition: all 0.2s;
}

nav button:hover {
  background-color: #f0f0f0;
  color: var(--primary-color);
}

nav button.active {
  background-color: #fff0f2;
  color: var(--primary-color);
  border-right: 3px solid var(--primary-color);
}

.icon {
  margin-right: 10px;
}

.main-content {
  flex: 1;
  padding: 30px;
  overflow-y: auto;
}

/* Global Component Styles */
h2 {
  margin-top: 0;
  border-bottom: 2px solid var(--primary-color);
  padding-bottom: 10px;
  display: inline-block;
  margin-bottom: 20px;
}

button.primary-btn {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
  transition: opacity 0.2s;
}

button.primary-btn:hover {
  opacity: 0.9;
}

input[type="text"], input[type="number"], textarea {
  border: 1px solid #ddd;
  border-radius: 6px;
  padding: 10px;
  font-size: 14px;
  transition: border-color 0.2s;
}

input:focus, textarea:focus {
  outline: none;
  border-color: var(--primary-color);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
