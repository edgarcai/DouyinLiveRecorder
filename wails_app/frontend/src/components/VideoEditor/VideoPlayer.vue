<script setup>
import { ref, onUnmounted } from 'vue'

const emit = defineEmits(['capture'])

const videoUrl = ref('')
const videoRef = ref(null)
const isPlaying = ref(false)

const handleFileSelect = (event) => {
  const file = event.target.files[0]
  if (file) {
    if (videoUrl.value) {
      URL.revokeObjectURL(videoUrl.value)
    }
    videoUrl.value = URL.createObjectURL(file)
    isPlaying.value = false
  }
}

const togglePlay = () => {
  if (!videoRef.value) return
  if (isPlaying.value) {
    videoRef.value.pause()
  } else {
    videoRef.value.play()
  }
  isPlaying.value = !isPlaying.value
}

const captureFrame = () => {
  if (!videoRef.value) return
  
  const canvas = document.createElement('canvas')
  canvas.width = videoRef.value.videoWidth
  canvas.height = videoRef.value.videoHeight
  const ctx = canvas.getContext('2d')
  ctx.drawImage(videoRef.value, 0, 0, canvas.width, canvas.height)
  
  // Convert to data URL (base64)
  const dataUrl = canvas.toDataURL('image/jpeg', 0.8)
  emit('capture', dataUrl)
}

onUnmounted(() => {
  if (videoUrl.value) {
    URL.revokeObjectURL(videoUrl.value)
  }
})
</script>

<template>
  <div class="video-player-container">
    <div class="video-wrapper">
      <video 
        ref="videoRef" 
        :src="videoUrl" 
        class="video-element"
        @click="togglePlay"
        @ended="isPlaying = false"
      ></video>
      <div v-if="!videoUrl" class="placeholder">
        <label class="upload-btn">
          <input type="file" accept="video/*" @change="handleFileSelect" hidden />
          <span class="icon">📂</span>
          <span>Open Video File</span>
        </label>
      </div>
    </div>
    
    <div class="controls" v-if="videoUrl">
      <div class="left-controls">
        <label class="icon-btn" title="Open New Video">
          <input type="file" accept="video/*" @change="handleFileSelect" hidden />
          📁
        </label>
        <button class="icon-btn" @click="togglePlay">
          {{ isPlaying ? '⏸️' : '▶️' }}
        </button>
      </div>
      
      <button class="capture-btn" @click="captureFrame">
        📸 Capture Frame
      </button>
    </div>
  </div>
</template>

<style scoped>
.video-player-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: black;
  border-radius: 12px;
  overflow: hidden;
  position: relative;
}

.video-wrapper {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  background: #000;
  min-height: 200px;
}

.video-element {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #666;
  width: 100%;
  height: 100%;
}

.upload-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  padding: 32px;
  border: 2px dashed #333;
  border-radius: 16px;
  transition: all 0.3s;
}

.upload-btn:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
  background: rgba(255, 44, 85, 0.05);
}

.upload-btn .icon {
  font-size: 48px;
}

.controls {
  height: 60px;
  background: #1a1a1a;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  border-top: 1px solid #333;
}

.left-controls {
  display: flex;
  gap: 16px;
}

.icon-btn {
  background: transparent;
  border: none;
  color: white;
  font-size: 20px;
  cursor: pointer;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: background 0.2s;
}

.icon-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

.capture-btn {
  background: var(--primary-color);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 20px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: transform 0.1s;
}

.capture-btn:hover {
  filter: brightness(1.1);
  transform: scale(1.05);
}

.capture-btn:active {
  transform: scale(0.95);
}
</style>
