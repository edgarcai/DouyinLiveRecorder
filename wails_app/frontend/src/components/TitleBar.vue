<script setup>
import {WindowMinimise, WindowToggleMaximise, Quit} from '../wailsjs/runtime/runtime.js'

const minimise = () => WindowMinimise()
const maximise = () => WindowToggleMaximise()
const quit = () => Quit()
</script>

<template>
  <div class="titlebar" style="--wails-draggable:drag">
    <div class="title-content">
      <div class="app-icon">🔴</div>
      <div class="app-title">{{ $t('app.title') }}</div>
    </div>
    <div class="window-controls">
      <button class="control-btn min" @click="minimise">─</button>
      <button class="control-btn max" @click="maximise">□</button>
      <button class="control-btn close" @click="quit">✕</button>
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

.window-controls {
  display: flex;
  height: 100%;
  -webkit-app-region: no-drag; /* Important: buttons must be clickable */
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
