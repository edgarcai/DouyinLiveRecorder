<script setup>
// 登录模态框组件
import { ref } from 'vue'
import { Login } from '../wailsjs/go/main/App.js'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  visible: Boolean
})

const emit = defineEmits(['update:visible', 'login-success'])
const { t } = useI18n()

const username = ref('')
const password = ref('')
const loading = ref(false)
const errorMsg = ref('')

const handleClose = () => {
  emit('update:visible', false)
  resetForm()
}

const resetForm = () => {
  username.value = ''
  password.value = ''
  errorMsg.value = ''
}

const handleLogin = async () => {
  if (!username.value || !password.value) {
    return
  }
  
  loading.value = true
  errorMsg.value = ''
  
  try {
    const result = await Login(username.value, password.value)
    if (result.success) {
      emit('login-success', result)
      handleClose()
    } else {
      errorMsg.value = result.message || t('login.failed')
    }
  } catch (err) {
    errorMsg.value = err.message || t('login.failed')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <transition name="modal">
    <div v-if="visible" class="modal-mask" @click.self="handleClose">
      <div class="modal-container">
        <div class="modal-header">
          <h3>{{ t('login.title') }}</h3>
          <button class="close-btn" @click="handleClose">✕</button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label>{{ t('login.username') }}</label>
            <input 
              type="text" 
              v-model="username" 
              :placeholder="t('login.placeholder.username')"
              @keyup.enter="handleLogin"
            >
          </div>
          
          <div class="form-group">
            <label>{{ t('login.password') }}</label>
            <input 
              type="password" 
              v-model="password" 
              :placeholder="t('login.placeholder.password')"
              @keyup.enter="handleLogin"
            >
          </div>

          <div v-if="errorMsg" class="error-msg">
            {{ errorMsg }}
          </div>
        </div>

        <div class="modal-footer">
          <button class="cancel-btn" @click="handleClose">
            {{ t('login.cancel') }}
          </button>
          <button 
            class="submit-btn" 
            :disabled="loading" 
            @click="handleLogin"
          >
            {{ loading ? '...' : t('login.submit') }}
          </button>
        </div>
      </div>
    </div>
  </transition>
</template>

<style scoped>
.modal-mask {
  position: fixed;
  z-index: 9998;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: opacity 0.3s ease;
  backdrop-filter: blur(4px);
}

.modal-container {
  width: 360px;
  background-color: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.33);
  transition: all 0.3s ease;
  padding: 24px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  color: var(--text-primary);
}

.close-btn {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  color: var(--text-secondary);
  padding: 4px;
}

.close-btn:hover {
  color: var(--text-primary);
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  color: var(--text-secondary);
}

.form-group input {
  width: 100%;
  box-sizing: border-box;
}

.error-msg {
  color: #ff2c55;
  font-size: 13px;
  margin-top: 8px;
  text-align: center;
}

.modal-footer {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.cancel-btn {
  padding: 8px 16px;
  border: 1px solid var(--border-color);
  background: white;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  color: var(--text-secondary);
  transition: all 0.2s;
}

.cancel-btn:hover {
  border-color: var(--text-secondary);
  color: var(--text-primary);
}

.submit-btn {
  padding: 8px 24px;
  background: var(--primary-color);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: background 0.2s;
}

.submit-btn:hover {
  background: var(--primary-hover);
}

.submit-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

/* Modal Transition */
.modal-enter-from {
  opacity: 0;
}

.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-container,
.modal-leave-to .modal-container {
  transform: scale(0.95);
  opacity: 0;
}
</style>
