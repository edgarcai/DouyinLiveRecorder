<script setup>
import {ref, onMounted} from 'vue'

const props = defineProps({
  recordingCount: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits(['restore'])

const restore = () => {
  emit('restore')
}
</script>

<template>
  <div class="floating-ball" @dblclick="restore" style="--wails-draggable:drag">
    <div class="ball-content" :class="{recording: recordingCount > 0}">
      <span class="icon">🔴</span>
      <span v-if="recordingCount > 0" class="badge">{{ recordingCount }}</span>
    </div>
  </div>
</template>

<style scoped>
.floating-ball {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  overflow: visible;
  transition: transform 0.2s;
  border: 2px solid white;
}

.floating-ball:hover {
  transform: scale(1.05);
}

.ball-content {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}

.ball-content.recording {
  animation: pulse 2s infinite;
  background: rgba(255, 44, 85, 0.1);
  border: 2px solid #ff2c55;
}

.icon {
  font-size: 24px;
}

.badge {
  position: absolute;
  top: -2px;
  right: -2px;
  background: #ff2c55;
  color: white;
  font-size: 10px;
  font-weight: bold;
  min-width: 16px;
  height: 16px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
  border: 1px solid white;
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(255, 44, 85, 0.4);
  }
  70% {
    box-shadow: 0 0 0 10px rgba(255, 44, 85, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(255, 44, 85, 0);
  }
}
</style>
