<script setup>
import { ref, onMounted } from 'vue'
import { CheckAppUpdate, StartAppUpdate, GetConfig } from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'
import logo from '../assets/images/logo-universal.png'

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
    }, 1000)
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
  <div class="splash-container">
    <div class="bg-decoration">
        <!-- 抽象形状或淡出的应用截图可以放在这里 -->
    </div>
    <div class="content">
      <img :src="logo" class="logo" alt="Logo" />
      <h1 class="title">{{ appTitle }}</h1>
      
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
  background: var(--bg-color);
  display: flex;
  align-items: center;
  justify-content: center;
  user-select: none;
  position: relative;
  overflow: hidden;
}

.content {
  text-align: center;
  width: 80%;
  max-width: 400px;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.logo {
  width: 80px;
  height: 80px;
  margin-bottom: 16px;
  border-radius: 16px;
  box-shadow: 0 8px 24px rgba(255, 44, 85, 0.15);
}

.title {
  font-size: 24px;
  color: var(--text-primary);
  margin-bottom: 60px;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.status-area {
  width: 100%;
  padding: 0 20px;
}

.progress-bar-container {
  width: 100%;
  height: 6px;
  background: rgba(0,0,0,0.05);
  border-radius: 3px;
  overflow: hidden;
  margin-bottom: 12px;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, var(--primary-color), var(--primary-hover));
  transition: width 0.1s linear;
  border-radius: 3px;
}

.status-text {
  font-size: 13px;
  color: var(--text-secondary);
  text-align: center;
  font-weight: 500;
}

.status-text.loading {
  color: var(--primary-color);
}
</style>
