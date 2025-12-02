<script setup>
import { ref, onMounted } from 'vue'
import { CheckAppUpdate, StartAppUpdate, GetConfig } from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'
import logo from '../assets/images/logo-universal.png'
import splashBg from '../assets/images/splash_bg.png'

const emit = defineEmits(['ready'])

const status = ref('正在检查更新...')
const progress = ref(0)
const showProgress = ref(false)
const hasUpdate = ref(false)
const appTitle = ref('Multi Platform Live Recorder') // 默认回退标题

onMounted(async () => {
  // 获取标题配置
  try {
    const config = await GetConfig()
    if (config && config.WindowSettings && config.WindowSettings.AppTitle) {
      appTitle.value = config.WindowSettings.AppTitle
    }
  } catch (e) {
    console.error("Failed to load config:", e)
  }

  // 监听进度
  EventsOn("update-progress", (p) => {
    progress.value = p
    status.value = `正在下载更新... ${p.toFixed(2)}%`
  })

  EventsOn("update-complete", () => {
    status.value = "更新完成，即将重启..."
    setTimeout(() => {
      // 在实际应用中，我们可能会在这里重启。
      // 目前，只需进入主应用以显示它已工作。
      emit('ready')
    }, 1000)
  })

  try {
    // 人为延迟以显示启动页
    setTimeout(async () => {
        const result = await CheckAppUpdate()
        if (result.error) {
        console.error(result.error)
        startApp()
        return
        }

        if (result.hasUpdate) {
        hasUpdate.value = true
        status.value = `发现新版本 ${result.newVersion}`
        showProgress.value = true
        // 自动开始更新
        StartAppUpdate()
        } else {
        startApp()
        }
    }, 1500) // 稍微增加延迟以展示精美画面
  } catch (e) {
    console.error(e)
    startApp()
  }
})

const startApp = () => {
  status.value = "准备启动..."
  setTimeout(() => {
    emit('ready')
  }, 500)
}
</script>

<template>
  <div class="splash-container" :style="{ backgroundImage: `url(${splashBg})` }">
    <div class="overlay"></div>
    <div class="content">
      <div class="header-group">
        <img :src="logo" class="logo" alt="Logo" />
        <h1 class="title">{{ appTitle }}</h1>
      </div>
      
      <div class="status-area">
        <div v-if="showProgress" class="progress-bar-container">
          <div class="progress-bar" :style="{ width: progress + '%' }"></div>
        </div>
        <div class="status-text" :class="{ 'loading': showProgress }">{{ status }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.splash-container {
  width: 100vw;
  height: 100vh;
  background-color: #000;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  display: flex;
  flex-direction: column;
  justify-content: flex-end; /* Align content to bottom */
  align-items: center;
  user-select: none;
  position: relative;
  overflow: hidden;
}

.overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  /* Linear gradient from bottom to allow text readability but keep top clear */
  background: linear-gradient(to top, rgba(0,0,0,0.9) 0%, rgba(0,0,0,0.6) 30%, rgba(0,0,0,0) 100%);
  z-index: 1;
}

.content {
  width: 100%;
  padding: 0 40px 60px 40px; /* Padding from bottom */
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  animation: slideUp 0.8s ease-out;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(40px); }
  to { opacity: 1; transform: translateY(0); }
}

.header-group {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 32px;
}

.logo {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.title {
  font-size: 24px;
  color: #ffffff;
  font-weight: 700;
  letter-spacing: 0.5px;
  text-shadow: 0 2px 4px rgba(0,0,0,0.5);
  margin: 0; /* Remove default margin */
  /* Remove gradient text for better readability against complex background */
}

.status-area {
  width: 100%;
  max-width: 400px;
}

.progress-bar-container {
  width: 100%;
  height: 4px; /* Thinner progress bar */
  background: rgba(255,255,255,0.15);
  border-radius: 2px;
  overflow: hidden;
  margin-bottom: 12px;
  backdrop-filter: blur(4px);
}

.progress-bar {
  height: 100%;
  background: #ff2c55; /* Solid color for better visibility */
  box-shadow: 0 0 10px rgba(255, 44, 85, 0.8);
  transition: width 0.1s linear;
  border-radius: 2px;
}

.status-text {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.7);
  text-align: center;
  font-weight: 400;
  letter-spacing: 0.5px;
  text-transform: uppercase;
}

.status-text.loading {
  color: #fff;
  text-shadow: 0 0 8px rgba(255, 255, 255, 0.5);
}
</style>
