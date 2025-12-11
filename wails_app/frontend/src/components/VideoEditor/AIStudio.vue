<script setup>
import { ref, watch } from 'vue'
import { GenerateAIContent } from '../../wailsjs/go/main/App'

const props = defineProps({
  capturedFrame: {
    type: String,
    default: ''
  }
})

const prompt = ref('')
const generatedContent = ref('')
const isLoading = ref(false)
const error = ref('')

// Pre-defined prompts
const promptTemplates = [
  { label: 'Describe Scene', text: 'Describe what is happening in this scene in detail.' },
  { label: 'Social Copy', text: 'Write a viral social media caption for this video frame.' },
  { label: 'Analyze', text: 'Analyze the visual elements, lighting, and composition of this frame.' },
]

const setPrompt = (text) => {
  prompt.value = text
}

const generate = async () => {
  if (!props.capturedFrame) return
  if (!prompt.value) return

  isLoading.value = true
  error.value = ''
  generatedContent.value = ''

  try {
    // Strip the data URL prefix to get raw base64
    // data:image/jpeg;base64,......
    const base64Data = props.capturedFrame.split(',')[1]
    
    // Call backend
    // Note: Wails handles base64 string -> []byte conversion automatically for Go handlers
    const result = await GenerateAIContent(prompt.value, base64Data)
    generatedContent.value = result
  } catch (e) {
    console.error(e)
    error.value = 'Failed to generate content: ' + (e.message || e)
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="ai-studio-container">
    <div class="preview-section">
      <div v-if="capturedFrame" class="image-preview">
        <img :src="capturedFrame" alt="Captured Frame" />
      </div>
      <div v-else class="placeholder">
        <div class="icon">🖼️</div>
        <div>Capture a frame from the video to start</div>
      </div>
    </div>

    <div class="interaction-section">
      <div class="prompt-templates">
        <button 
          v-for="tmpl in promptTemplates" 
          :key="tmpl.label"
          class="chip"
          @click="setPrompt(tmpl.text)"
        >
          {{ tmpl.label }}
        </button>
      </div>

      <div class="input-area">
        <textarea 
          v-model="prompt" 
          placeholder="Enter a prompt for AI..."
          :disabled="isLoading || !capturedFrame"
        ></textarea>
        <button 
          class="generate-btn" 
          @click="generate" 
          :disabled="isLoading || !prompt || !capturedFrame"
        >
          <span v-if="isLoading" class="spinner"></span>
          <span v-else>✨ Generate</span>
        </button>
      </div>

      <div class="result-area" v-if="generatedContent || error">
        <div v-if="error" class="error-message">
          {{ error }}
        </div>
        <div v-else class="content-display">
          <h3>AI Output:</h3>
          <div class="markdown-body" style="white-space: pre-wrap;">{{ generatedContent }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ai-studio-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  gap: 16px;
  overflow: hidden;
}

.preview-section {
  flex: 1;
  background: #f8f9fa;
  border-radius: 12px;
  border: 1px solid #ebeef5;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  max-height: 40%;
}

.image-preview {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #000;
}

.image-preview img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  color: #909399;
  gap: 8px;
}

.placeholder .icon {
  font-size: 32px;
}

.interaction-section {
  flex: 1.5;
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow-y: auto;
  padding-bottom: 12px;
}

.prompt-templates {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.chip {
  padding: 6px 12px;
  border-radius: 16px;
  background: #f0f2f5;
  border: 1px solid #e4e7ed;
  color: #606266;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.chip:hover {
  background: #fff;
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.input-area {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

textarea {
  flex: 1;
  min-height: 80px;
  resize: vertical;
  font-family: inherit;
}

.generate-btn {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  border: none;
  padding: 0 24px;
  height: 80px;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.1s, box-shadow 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.generate-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
}

.generate-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  background: #c0c4cc;
}

.result-area {
  background: white;
  border-radius: 8px;
  border: 1px solid #ebeef5;
  padding: 16px;
  flex: 1;
  overflow-y: auto;
  box-shadow: 0 2px 8px rgba(0,0,0,0.02);
}

.error-message {
  color: #f56c6c;
}

.content-display h3 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #909399;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 3px solid rgba(255,255,255,0.3);
  border-radius: 50%;
  border-top-color: #fff;
  animation: spin 1s ease-in-out infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
