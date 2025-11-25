import { createApp, defineComponent, ref } from 'vue'

// IPC示例：在未生成 wailsjs 时，提供占位方法。
function startRecordStub() {
  console.log('调用后端 StartRecordRPC（占位）')
  // 真实调用示例（生成 wailsjs 后）：
  // import { StartRecordRPC } from '../wailsjs/go/internal/adapter/wails/RecorderService'
  // await StartRecordRPC({...})
}

async function renderConfigStub() {
  console.log('调用后端 RenderRPC（占位）')
  // 真实调用示例：
  // const val = await RenderRPC('录制设置.输出目录', { key_name: 'ROOT' })
  return 'ROOT/downloads'
}

async function fetchRoomStub() {
  console.log('调用后端 FetchRoomInfoRPC（占位）')
  return { platform: 'douyin', is_live: true, m3u8_url: 'http://m3u8', quality: '原画' }
}

async function selectSourceStub(room: any) {
  console.log('调用后端 SelectSourceRPC（占位）')
  return { url: room.m3u8_url || 'http://flv', quality: room.quality, valid: true }
}

async function planRecordStub(url: string) {
  console.log('调用后端 PlanRecordRPC（占位）')
  // 真实调用示例：
  // const [args, ns] = await PlanRecordRPC(url, '/tmp/out/file', 'mp4', 60)
  return { args: ['-i', 'http://m3u8', '-c', 'copy', '/tmp/out/file.mp4'], ns: 10_000_000 }
}

const App = defineComponent({
  setup() {
    const status = ref('就绪')
    const cfg = ref('')
    const roomInfo = ref<any>(null)
    const source = ref<any>(null)
    const authVal = ref('')
    const plan = ref<any>(null)
    const start = async () => {
      status.value = '调用中…'
      await startRecordStub()
      status.value = '已调用'
    }
    const showCfg = async () => {
      cfg.value = await renderConfigStub()
    }
    const parseRoom = async () => {
      roomInfo.value = await fetchRoomStub()
    }
    const chooseSource = async () => {
      if (!roomInfo.value) await parseRoom()
      source.value = await selectSourceStub(roomInfo.value)
    }
    const getAuth = async () => { authVal.value = await getKeyStub() }
    const setAuth = async () => { await setKeyStub() }
    const planRecord = async () => { plan.value = await planRecordStub('https://v.douyin.com/xxx') }
    return { status, start, cfg, showCfg, roomInfo, source, parseRoom, chooseSource, authVal, getAuth, setAuth, plan, planRecord }
  },
  template: `
    <main style="padding:16px;font-family:system-ui">
      <h1>DouyinLiveRecorder</h1>
      <p>状态：{{ status }}</p>
      <button @click="start">开始录制（占位IPC）</button>
      <hr />
      <p>渲染配置示例：{{ cfg }}</p>
      <button @click="showCfg">渲染录制输出目录（占位IPC）</button>
      <hr />
      <p>房间信息：{{ roomInfo }}</p>
      <button @click="parseRoom">解析抖音房间（占位IPC）</button>
      <p>选择源：{{ source }}</p>
      <button @click="chooseSource">选择播放源（占位IPC）</button>
      <hr />
      <p>管线规划结果：{{ plan }}</p>
      <button @click="planRecord">解析→选择→参数构造（占位IPC）</button>
      <hr />
      <p>登录配置（遮蔽展示）：{{ authVal }}</p>
      <button @click="getAuth">读取键（占位IPC）</button>
      <button @click="setAuth">写入键（占位IPC）</button>
    </main>
  `
})

createApp(App).mount('#app')
async function getKeyStub() {
  console.log('调用后端 GetKeyRPC（占位）')
  return 'TOKEN-****'
}

async function setKeyStub() {
  console.log('调用后端 SetKeyRPC（占位）')
}
