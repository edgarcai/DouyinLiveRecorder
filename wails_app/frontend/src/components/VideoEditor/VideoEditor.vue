<script setup>
import { ref } from 'vue'
import VideoPlayer from './VideoPlayer.vue'
import AIStudio from './AIStudio.vue'

const capturedFrame = ref('')

const handleCapture = (dataUrl) => {
  capturedFrame.value = dataUrl
}
</script>

<template>
  <div class="video-editor-layout">
    <div class="header">
      <h2>✨ AI Video Studio</h2>
    </div>
    <div class="editor-content">
      <div class="left-panel">
        <VideoPlayer @capture="handleCapture" />
      </div>
      <div class="right-panel">
        <AIStudio :captured-frame="capturedFrame" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.video-editor-layout {
  display: flex;
  flex-direction: column;
  height: 100%;
  gap: 16px;
}

.header h2 {
  margin: 0;
  font-size: 20px;
  background: linear-gradient(45deg, #ff2c55, #6366f1);
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.editor-content {
  display: flex;
  gap: 20px;
  flex: 1;
  overflow: hidden;
}

.left-panel {
  flex: 3;
  min-width: 0;
}

.right-panel {
  flex: 2;
  min-width: 300px;
  background: white;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.05);
}

@media (max-width: 1024px) {
  .editor-content {
    flex-direction: column;
    overflow-y: auto;
  }
  
  .left-panel, .right-panel {
    flex: none;
    height: 50vh;
  }
}
</style>
