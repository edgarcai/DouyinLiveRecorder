<script setup>
import {WindowMinimise, WindowToggleMaximise, Quit} from '../wailsjs/runtime/runtime.js'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  user: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['login-click', 'logout-click'])
const { t } = useI18n()

const minimise = () => WindowMinimise()
const maximise = () => WindowToggleMaximise()
const quit = () => Quit()
</script>

<template>
  <div class="titlebar" style="--wails-draggable:drag" @dblclick="maximise">
    <div class="title-content">
      <div class="app-icon">🔴</div>
      <div class="app-title">{{ $t('app.title') }}</div>
    </div>
    
    <div class="right-section">
      <div class="user-area" style="--wails-draggable:no-drag">
        <div v-if="user" class="user-info" @click="$emit('logout-click')">
          <span class="user-avatar">👤</span>
          <span class="username">{{ user.username }}</span>
        </div>
        <button v-else class="login-btn" @click="$emit('login-click')">
          {{ t('login.title') }}
        </button>
      </div>

      <div class="window-controls">
        <button class="control-btn min" @click="minimise">─</button>
        <button class="control-btn max" @click="maximise">□</button>
        <button class="control-btn close" @click="quit">✕</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.titlebar {
  height: 32px;
  background: #fff;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 0 0 16px;
  border-bottom: 1px solid var(--border-color);
  user-select: none;
}

.title-content {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.app-icon {
  font-size: 14px;
}

.right-section {
  display: flex;
  align-items: center;
  height: 100%;
}

.user-area {
  margin-right: 8px;
  display: flex;
  align-items: center;
}

.login-btn {
  background: transparent;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 2px 8px;
  font-size: 12px;
  cursor: pointer;
  color: var(--text-secondary);
  transition: all 0.2s;
}

.login-btn:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  padding: 2px 8px;
  border-radius: 4px;
  transition: background 0.2s;
}

.user-info:hover {
  background: #f0f2f5;
}

.user-avatar {
  font-size: 14px;
}

.username {
  font-size: 12px;
  color: var(--text-primary);
}

.window-controls {
  display: flex;
  height: 100%;
  -webkit-app-region: no-drag; /* 重要：按钮必须可点击 */
}

.control-btn {
  width: 46px;
  height: 100%;
  border: none;
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 14px;
  color: var(--text-secondary);
  transition: background-color 0.2s, color 0.2s;
}

.control-btn:hover {
  background-color: #f0f2f5;
  color: var(--text-primary);
}

.control-btn.close:hover {
  background-color: #e81123;
  color: white;
}
</style>
