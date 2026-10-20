<template>
  <div class="content">
      <!-- 左侧：状态 / 连接 / 日志 -->
      <div class="left-panel">
        <div class="side-card status-panel">
          <div class="status-grid">
            <div class="indicator-item">
              <div class="led led-connect" :class="{ 'active': Heartbeat.Running }"></div>
              <span class="indicator-label">PLC连接</span>
            </div>
            <div class="indicator-item">
              <div class="led led-alarm" :class="{ 'active': Heartbeat.Alarm != '' && Heartbeat.Running }"></div>
              <span class="indicator-label">异常警报</span>
            </div>
          </div>
        </div>

        <div class="side-card config-panel">
          <input v-model="Target.ip" type="text" class="ip-input-modern" placeholder="PLC IP Address" />
          <div class="conn-btns-group">
            <button class="btn btn-connect" @click="handleConnect">{{Heartbeat.Running ? '断开连接' : '连接设备'}}</button>
          </div>
        </div>

        <div class="side-card log-panel">
          <div class="panel-label">实时日志</div>

          <div class="info-log-box">
            <div class="log-scroll-area">
              <p v-for="(log, i) in logs" :key="i" class="log-line">
                <span class="log-time">{{ log.time }}</span> {{ log.msg }}
              </p>
            </div>
          </div>

          <div class="log-actions">
            <button class="log-btn" @click="handleClearLogs">清除日志</button>
            <button class="log-btn" @click="handleExportLogs">导出日志</button>
          </div>
        </div>
      </div>

      <!-- 中间：数值 / 示意图 / 操作 -->
      <div class="middle-panel">
        <section class="metrics-grid">
            <div class="metric-card">
                <span class="label">供气压力</span>
                <div class="value-row">
                <span class="value">{{ Heartbeat.InputPressure.toFixed(2) }}</span>
                <span class="unit">MPa</span>
                </div>
            </div>
            <div class="metric-card">
                <span class="label">气室压力</span>
                <div class="value-row">
                <span class="value">{{ Heartbeat.CylinderPressure.toFixed(2) }}</span>
                <span class="unit">MPa</span>
                </div>
            </div>
           
            <div class="metric-card">
                <span class="label">尾部真空度</span>
                <div class="value-row">
                <span class="value">{{ Heartbeat.TailVacuumDegree.toFixed(0) }}</span>
                <span class="unit">Pa</span>
                </div>
            </div>

             <div class="metric-card">
                <span class="label">靶室真空度</span>
                <div class="value-row">
                <span class="value">{{ Heartbeat.TargetVacuumDegree.toFixed(0) }}</span>
                <span class="unit">Pa</span>
                </div>
            </div>

        </section>

        <section class="device-visualization">
          <div class="schematic-view">
            <div class="led-list">
              <div v-for="led in ledItems" :key="led.key" class="led-item" :class="{ on: Output[led.key] && Heartbeat.Running }">
                <span class="led-dot"></span>
                <span class="led-text">{{ led.label }}</span>
              </div>
            </div>
            <img src="../../assets/images/devices/gasgun1.png" alt="气炮示意图" class="cannon-image" draggable="false" />
          </div>
        </section>

        <footer class="action-bar">
          <!-- 1. 尾部真空 -->
          <div class="action-zone">
            <div class="zone-title">尾部真空</div>
            <button class="ctrl-btn vacuum-tail-btn" :class="{ 'active': Output.TailVacuumPump }"
              @click="handleTailVacuumSwitch(!Output.TailVacuumPump)">
              {{ Output.TailVacuumPump ? '停抽尾部' : '抽尾部' }}
            </button>
          </div>

          <!-- 2. 靶室真空 -->
          <div class="action-zone">
            <div class="zone-title">靶室真空</div>
            <button class="ctrl-btn vacuum-target-btn" :class="{ 'active': Output.TarVacuumPump }"
              @click="handleVacuumSwitch(!Output.TarVacuumPump)">
              {{ Output.TarVacuumPump ? '停抽靶室' : '抽靶室' }}
            </button>
          </div>

          <!-- 3. 气室压力控制 -->
          <div class="action-zone">
            <div class="zone-title">气室压力控制</div>
            <div class="mode-row">
              <span class="mode-label">模式：{{ Target.AutoMode ? '自动' : '手动' }}</span>
              <div class="toggle-switch"
                :class="{ 'active': Target.AutoMode, 'disabled': Target.AutoRunning }"
                @click="!Target.AutoRunning && (Target.AutoMode = !Target.AutoMode)"></div>
            </div>
            <div v-if="Target.AutoMode" class="auto-group">
              <div class="auto-fields">
                <label class="auto-label">目标 (MPa)</label>
                <input v-model.number="Target.TargetPressure" type="number" step="0.1" class="pressure-input" />
              </div>
              <button :class="['ctrl-btn', Target.AutoRunning ? 'stop-btn' : 'start-btn']" @click="handleAutoPressure">
                {{ Target.AutoRunning ? '停止' : '开始' }}
              </button>
            </div>
            <div v-else class="manual-controls">
              <button class="mini-btn plus-btn"
                @mousedown="handleInletSwitch(true)" @mouseup="handleInletSwitch(false)" @mouseleave="handleInletSwitch(false)">进气</button>
              <button class="mini-btn minus-btn"
                @mousedown="handleOutletSwitch(true)" @mouseup="handleOutletSwitch(false)" @mouseleave="handleOutletSwitch(false)">排气</button>
            </div>
          </div>

          <!-- 4. 发射控制 -->
          <div class="action-zone">
            <div class="zone-title">发射控制</div>
            <div class="mode-row">
              <span class="mode-label">模式：{{ Target.ExternalTrigger ? '外触发' : '内触发' }}</span>
              <div class="toggle-switch"
                :class="{ 'active': Target.ExternalTrigger, 'disabled': Target.ExternalTriggerWaiting }"
                style="--off-text: '内触发'; --on-text: '外触发'"
                @click="!Target.ExternalTriggerWaiting && (Target.ExternalTrigger = !Target.ExternalTrigger)"></div>
            </div>
            <button class="ctrl-btn fire-btn" v-if="!Target.ExternalTrigger" @click="requestFire">立即发射</button>
            <button class="ctrl-btn prepare-btn" :class="{ 'waiting': Target.ExternalTriggerWaiting }" v-if="Target.ExternalTrigger" @click="requestPrepareFire">
              {{ Target.ExternalTriggerWaiting ? '等待触发...' : '发射准备' }}
            </button>
          </div>

          <!-- 5. 系统释放 -->
          <div class="action-zone">
            <div class="zone-title">系统释放</div>
            <button class="ctrl-btn release-btn" :class="{ 'active': Output.SystemDecompress }"
              @click="handleSystemRelease(!Output.SystemDecompress)">
              {{ Output.SystemDecompress ? '释放中...' : '系统释放' }}
            </button>
          </div>
        </footer>
      </div>
  </div>

  <!-- 发射确认模态框 -->
  <div v-if="confirmModal.show" class="confirm-overlay" @click.self="confirmModal.show = false">
    <div class="confirm-modal">
      <div class="confirm-header" :class="{ 'danger': confirmModal.type === 'fire' }">
        <h3>{{ confirmModal.title }}</h3>
      </div>
      <div class="confirm-body">
        <p>{{ confirmModal.message }}</p>
      </div>
      <div class="confirm-actions">
        <button class="confirm-btn cancel-btn" @click="confirmModal.show = false">取消</button>
        <button class="confirm-btn ok-btn" :class="{ 'danger': confirmModal.type === 'fire' }" @click="onConfirm">确认执行</button>
      </div>
    </div>
  </div>
</template>


<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime'
import {
  ConnectDevice,
  DisconnectDevice,
  InletSwitch,
  ExhaustSwitch,
  TargetVacuumSwitch,
  TailVacuumSwitch,
  SystemReleaseSwitch,
  AutoCylinderPressure,
  StopAutoCylinderPressure,
  FireSwitch,
  PrepareFire,
  ExternalFire,
  ExportAlarmHistory,
} from '../../../wailsjs/go/hepsgasgun1/HEPSGasGun1'



const Heartbeat = reactive({
  Running:            false, // 是否连接
  Alarm:              '',

  InputPressure:      0.00,   // 输入压力
  CylinderPressure:   0.00,   // 气室压力
  TargetVacuumDegree: 100000, // 靶室真空度
  TailVacuumDegree:   100000, // 尾部真空度
})

const Output = ref({
    Inlet:               false,
    Outlet:              false,
    Fire:                false,
    TargetVacuumRelease: false,
    TailVacuumPump:      false,
    TarVacuumPump:       false,
    SystemDecompress:    false,
})

// 示意图左上角 LED 输出指示
const ledItems = [
    { key: 'Inlet',               label: '进气阀' },
    { key: 'Outlet',              label: '出气阀' },
    { key: 'Fire',                label: '发射阀' },
    { key: 'TargetVacuumRelease', label: '靶室真空阀' },
    { key: 'TailVacuumPump',      label: '尾部真空泵' },
    { key: 'TarVacuumPump',       label: '靶室真空泵' },
    { key: 'SystemDecompress',    label: '系统释放阀' },
]

const Target = reactive({
    ip: '192.168.6.6',

    AutoMode:               false, // 压力控制：手动/自动模式
    AutoRunning:            false, // 自动压力运行中
    TargetPressure:         1.00,  // 自动压力目标值 (MPa)

    ExternalTrigger:        false, // 发射：内触发/外触发模式
    ExternalTriggerWaiting: false, // 外触发等待中
})

let lastFireTrigger = false


const logs = ref([])

onMounted(() => {
  EventsOn('heps_gasgun1_heartbeat', (info) => {
    if (!info) return
    try {
        Heartbeat.Running = info.Running
        Heartbeat.Alarm = info.Alarm

        Heartbeat.InputPressure = info.InputPressure || 0.00
        Heartbeat.CylinderPressure = info.CylinderPressure || 0.00
        Heartbeat.TargetVacuumDegree = info.TargetVacuumDegree || 100000
        Heartbeat.TailVacuumDegree = info.TailVacuumDegree || 100000

        // 后端 JSON tag 为 Output；逐字段赋值避免整体替换丢失键
        const out = info.Output || {}
        Output.value = {
            Inlet:               !!out.Inlet,
            Outlet:              !!out.Outlet,
            Fire:                !!out.Fire,
            TargetVacuumRelease: !!out.TargetVacuumRelease,
            TailVacuumPump:      !!out.TailVacuumPump,
            TarVacuumPump:       !!out.TarVacuumPump,
            SystemDecompress:    !!out.SystemDecompress,
        }

        // 外部触发信号上升沿：等待外触发时收到信号则执行发射
        const fireTrigger = !!(info.Input && info.Input['FireTrigger'])
        if (fireTrigger && !lastFireTrigger && Target.ExternalTriggerWaiting) {
            handleExternalTriggered()
        }
        lastFireTrigger = fireTrigger
    } catch (e) {
        addLog(`解析心跳数据失败: ${e}`)
    }

  })
  EventsOn('heps_gasgun1_message', (msg) => {
    addLog(msg)
  })
})

onUnmounted(() => {
  EventsOff('heps_gasgun1_heartbeat')
  EventsOff('heps_gasgun1_message')
})


const addLog = (msg) => {
  const now = new Date()
  const time = `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`
  logs.value.push({ time, msg })
  if (logs.value.length > 200) logs.value.shift()
}

// 清除 / 导出日志
const handleClearLogs = () => {
  if (!logs.value.length) return
  logs.value = []
  addLog('日志已清除')
}

const handleExportLogs = async () => {
  if (!logs.value.length) {
    addLog('暂无日志可导出')
    return
  }
  const content = logs.value.map((l) => `[${l.time}] ${l.msg}`).join('\n')
  try {
    await ExportAlarmHistory(content)
    addLog('日志已导出')
  } catch (e) {
    const msg = String(e)
    if (msg.includes('取消')) {
      addLog('已取消导出')
    } else {
      addLog(`导出失败: ${msg}`)
    }
  }
}


// 连接 / 断开
async function handleConnect() {
  if (!Heartbeat.Running) {
    try {
      await ConnectDevice(Target.ip)
      addLog(`尝试连接 ${Target.ip}`)
    } catch (e) {
      addLog(`连接失败: ${e}`)
    }
  } else {
    try {
      await DisconnectDevice()
      addLog('断开连接')
    } catch (e) {
      addLog(`断开失败: ${e}`)
    }
  }
}

// 进气开关（按住生效）
async function handleInletSwitch(open) {
  try {
    await InletSwitch(open)
    addLog(`${open ? '开启' : '关闭'}进气阀`)
  } catch (e) {
    addLog(`进气操作失败: ${e}`)
  }
}

// 出气开关（按住生效）
async function handleOutletSwitch(open) {
  try {
    await ExhaustSwitch(open)
    addLog(`${open ? '开启' : '关闭'}出气阀`)
  } catch (e) {
    addLog(`排气操作失败: ${e}`)
  }
}

async function handleVacuumSwitch(open) {
  try {
    await TargetVacuumSwitch(open)
    addLog(`${open ? '开启' : '停止'}抽靶室真空`)
  } catch (e) {
    addLog(`抽靶室失败: ${e}`)
  }
}

async function handleTailVacuumSwitch(open) {
  try {
    await TailVacuumSwitch(open)
    addLog(`${open ? '开启' : '停止'}抽尾部操作`)
  } catch (e) {
    addLog(`抽尾部失败: ${e}`)
  }
}

// 发射
async function handleFire() {
  try {
    await FireSwitch(500)
    addLog('执行发射 (500ms)')
  } catch (e) {
    addLog(`发射失败: ${e}`)
  }
}

// ===== 发射确认弹窗 =====
const confirmModal = reactive({
  show:    false,
  type:    '', // 'fire' | 'prepare'
  title:   '',
  message: '',
})

// 立即发射前弹出确认框
function requestFire() {
  confirmModal.type = 'fire'
  confirmModal.title = '确认发射'
  confirmModal.message = '即将执行发射操作，请确认设备状态正常。是否继续？'
  confirmModal.show = true
}

// 发射准备前弹出确认框；等待触发状态下点击为取消，不弹窗
function requestPrepareFire() {
  if (Target.ExternalTriggerWaiting) {
    handlePrepareFire()
    return
  }
  confirmModal.type = 'prepare'
  confirmModal.title = '确认发射准备'
  confirmModal.message = '即将执行发射准备，随后进入等待外触发状态。是否继续？'
  confirmModal.show = true
}

async function onConfirm() {
  const type = confirmModal.type
  confirmModal.show = false
  if (type === 'fire') {
    await handleFire()
  } else if (type === 'prepare') {
    await handlePrepareFire()
  }
}

// 自动压力控制：开始 / 停止
async function handleAutoPressure() {
  try {
    if (Target.AutoRunning) {
      await StopAutoCylinderPressure()
      Target.AutoRunning = false
      addLog('已停止自动压力控制')
    } else {
      await AutoCylinderPressure(Target.TargetPressure)
      Target.AutoRunning = true
      addLog(`开始自动压力控制，目标 ${Target.TargetPressure} MPa`)
    }
  } catch (e) {
    addLog(`自动压力操作失败: ${e}`)
  }
}

// 系统释放阀
async function handleSystemRelease(open) {
  try {
    await SystemReleaseSwitch(open)
    addLog(`${open ? '开启' : '关闭'}系统释放阀`)
  } catch (e) {
    addLog(`系统释放操作失败: ${e}`)
  }
}

// 发射准备（外触发）：准备后进入等待外触发状态；再次点击取消
async function handlePrepareFire() {
  try {
    if (!Target.ExternalTriggerWaiting) {
      await PrepareFire()
      Target.ExternalTriggerWaiting = true
      addLog('已准备发射，等待外触发信号中...')
    } else {
      Target.ExternalTriggerWaiting = false
      addLog('已取消准备发射')
    }
  } catch (e) {
    addLog(`发射准备失败: ${e}`)
  }
}

// 外部触发信号到来时执行发射并退出等待状态
async function handleExternalTriggered() {
  try {
    Target.ExternalTriggerWaiting = false
    await ExternalFire(500)
    addLog('外部信号触发发射 (500ms)')
  } catch (e) {
    addLog(`外触发发射失败: ${e}`)
  }
}
</script>

<style scoped>
.main-window {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* ===== 标题栏 ===== */
.title-bar {
  display: flex;
  width: 100%;
  height: 50px;
  align-items: center;
  gap: 24px;
  padding: 0 20px;
  box-sizing: border-box;
  background-color: #4e69b5;
  --wails-draggable: drag;
}

.logo {
  display: flex;
  align-items: center;
  gap: 8px;
}

.title {
  font-size: 18px;
  font-weight: 500;
  color: #ffffff;
}

.window-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  justify-content: flex-end;
  margin-left: auto;
  --wails-draggable: no-drag;
}

.title-divider {
    width: 1px;
    height: 18px;
    background: #94a3b8;
    margin: 0 6px;
    flex-shrink: 0;
}

.window-btn {
  width: 36px;
  height: 36px;
  border: none;
  background: transparent;
  color: #ffffff;
  cursor: pointer;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s;
  padding: 0;
}

.window-btn svg {
  width: 18px;
  height: 18px;
}

.window-btn:hover {
  background: rgba(255, 255, 255, 0.18);
  color: #ffffff;
  transform: scale(1.1);
}

.window-btn.close-btn:hover {
  background: rgba(80, 63, 63, 0.2);
  color: #ef4444;
}

/* ===== 内容区 ===== */
.content {
  flex: 1;
  display: flex;
  flex-direction: row;
  background-color: #edf1f8;
  min-height: 0;
}

/* 左侧栏 */
.left-panel {
  width: 270px;
  background: #232c49;
  display: flex;
  flex-direction: column;
  padding: 16px;
  box-shadow: 4px 0 15px rgba(0, 0, 0, 0.2);
  box-sizing: border-box;
  z-index: 10;
}

.side-card {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 16px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.status-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.indicator-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.led {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: #39426b;
  box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.5);
  transition: 0.3s;
}

.led-connect.active {
  background: #41b883;
  box-shadow: 0 0 10px #41b883;
}

.led-alarm.active {
  background: #fa5252;
  box-shadow: 0 0 10px #fa5252;
  animation: blink 1s infinite;
}

@keyframes blink {
  0% { opacity: 1; }
  50% { opacity: 0.3; }
  100% { opacity: 1; }
}

.indicator-label {
  font-size: 11px;
  color: #ced4da;
}

.ip-input-modern {
  width: 100%;
  background: #2b3557;
  border: 1px solid #46507a;
  padding: 10px;
  border-radius: 4px;
  color: #fff;
  font-size: 1.1rem;
  text-align: center;
  box-sizing: border-box;
  margin-bottom: 10px;
}

.conn-btns-group {
  display: flex;
  gap: 8px;
}

.btn {
  flex: 1;
  padding: 8px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: opacity 0.2s;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-connect {
  background: #4e69b5;
  color: #fff;
}

.btn-disconnect {
  background: #39426b;
  color: #cdd5ea;
}

/* 日志区 */
.log-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  margin-bottom: 16px;
}

.panel-label {
  font-size: 12px;
  color: #8a97b5;
  margin-bottom: 12px;
  font-weight: bold;
  text-transform: uppercase;
  flex-shrink: 0;
}

.info-log-box {
  flex: 1;
  min-height: 0;
  background: #171e36;
  border-radius: 6px;
  padding: 8px;
  overflow: hidden;
}

.log-scroll-area {
  height: 100%;
  overflow-y: auto;
  font-size: 12px;
}

.log-line {
  margin: 4px 0;
  color: #7fb3ff;
  line-height: 1.4;
}

.log-time {
  color: #5f6c94;
  margin-right: 6px;
}

/* 日志操作按钮 */
.log-actions {
  display: flex;
  gap: 10px;
  margin-top: 10px;
  flex-shrink: 0;
}

.log-btn {
  flex: 1;
  padding: 7px 0;
  border: 2px solid #2b3245;
  border-radius: 8px;
  background: #39426b;
  color: #dfe6f5;
  font-weight: bold;
  font-size: 0.95rem;
  cursor: pointer;
  transition: 0.2s;
  box-shadow: 0 2px 5px rgba(15, 23, 42, 0.15);
}

.log-btn:hover {
  background: #47529c;
  transform: translateY(-1px);
  box-shadow: 0 4px 9px rgba(15, 23, 42, 0.22);
}

.log-btn:active {
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.2);
}

.sidebar-footer {
  margin-top: auto;
}

.btn-settings-modern {
  width: 100%;
  padding: 12px;
  background: #39426b;
  border: none;
  border-radius: 8px;
  color: #a9b4d0;
  cursor: pointer;
  font-size: 13px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-settings-modern:hover {
  color: #fff;
  background: #4a5580;
}

/* 中间区 */
.middle-panel {
  flex: 1;
  padding: 20px;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
}

.metrics-grid {
  flex-shrink: 0;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 15px;
  margin-bottom: 20px;
}

.metric-card {
  background: #fff;
  padding: 20px;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  border-top: 4px solid #4e69b5;
}

.metric-card .label {
  font-size: 14px;
  color: #666;
}

.metric-card .value {
  font-size: 28px;
  font-weight: bold;
  color: #1a1c24;
}

.metric-card .unit {
  margin-left: 5px;
  color: #888;
}

.device-visualization {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-radius: 12px;
  margin-bottom: 20px;
}

.schematic-view {
  position: relative;
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: hidden;
  padding: 10px;
}

/* 左上角悬浮 LED 输出指示 */
.led-list {
  position: absolute;
  top: 10px;
  left: 10px;
  z-index: 10;
  display: flex;
  flex-direction: column;
  gap: 6px;
  background: rgba(23, 30, 54, 0.58);
  border: 1px solid #2b3245;
  border-radius: 8px;
  padding: 8px 12px;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.25);
}

.led-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.led-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #4a5578;
  border: 1px solid #2b3245;
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.4);
  transition: 0.2s;
  flex-shrink: 0;
}

.led-item.on .led-dot {
  background: #35d07f;
  border-color: #1d7a4a;
  box-shadow: 0 0 6px rgba(53, 208, 127, 0.8), inset 0 1px 2px rgba(255, 255, 255, 0.4);
}

.led-text {
  font-size: 12px;
  font-weight: bold;
  color: #aebadf;
  white-space: nowrap;
}

.led-item.on .led-text {
  color: #7bedb1;
}



.cannon-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.action-bar {
  flex-shrink: 0;
  height: 200px;
  display: flex;
  gap: 12px;
  align-items: stretch;
  background: #dfe6f3;
  padding: 12px;
  border-radius: 12px;
  box-sizing: border-box;
}

/* 操作子区域 */
.action-zone {
  flex: 1;
  min-width: 0;
  background: #fff;
  border-radius: 10px;
  padding: 10px 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  border: 1px solid #f0f0f0;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 10px;
  overflow-y: auto;
}

.zone-title {
  font-size: 14px;
  font-weight: 600;
  color: #1a1c24;
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.zone-title::before {
  content: '';
  width: 4px;
  height: 14px;
  background: #4e69b5;
  border-radius: 2px;
}

/* 手动/自动、内/外触发切换行 */
.mode-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.mode-label {
  font-size: 14px;
  color: #666;
  font-weight: 500;
}

.mode-row .toggle-switch {
  margin-left: auto;
}

/* 自动模式：左（说明文字+输入框）右（开始按钮） */
.auto-group {
  display: flex;
  gap: 8px;
  align-items: stretch;
  flex: 1;
}

.auto-fields {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
  min-width: 0;
}

.auto-label {
  font-size: 13px;
  color: #666;
  font-weight: 500;
}

.auto-fields .pressure-input {
  flex: 1;
  width: auto;
  min-width: 0;
}

.auto-group .start-btn {
  flex: 0 0 auto;
  min-width: 70px;
  min-height: 0;
  background: #4e69b5;
  color: #fff;
  border-color: #35497e;
}

.auto-group .start-btn:hover {
  background: #3d549a;
}

.auto-group .stop-btn {
  flex: 0 0 auto;
  min-width: 70px;
  min-height: 0;
  background: #e74c3c;
  color: #fff;
  border-color: #c0392b;
}

.auto-group .stop-btn:hover {
  background: #c0392b;
}

/* 通用 flex:1 按钮（发射/系统释放等） */
.action-zone > .ctrl-btn {
  flex: 1;
  min-height: 48px;
}

.action-zone > .manual-controls {
  flex: 1;
  align-items: stretch;
}

.action-zone > .manual-controls .mini-btn {
  min-height: 48px;
}

/* ===== 控制按钮 ===== */
.ctrl-btn {
  flex: 1;
  min-width: 80px;
  padding: 12px 16px;
  border-radius: 8px;
  border: 2px solid #2b3245;
  background: #f8f9fa;
  font-weight: bold;
  font-size: 17px;
  cursor: pointer;
  transition: 0.2s;
  box-sizing: border-box;
  text-align: center;
  box-shadow: 0 2px 5px rgba(15, 23, 42, 0.15);
}

.ctrl-btn:hover {
  background: #eef2fb;
  transform: translateY(-1px);
  box-shadow: 0 4px 9px rgba(15, 23, 42, 0.22);
}

.ctrl-btn.active {
  background: #dbe6fa;
  border-color: #1f2a47;
  color: #24397a;
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.15);
}

.ctrl-btn.fire-btn {
  background: #e5484d;
  color: #fff;
  border-color: #144f68;
}

.ctrl-btn.fire-btn:hover {
  background: #c73e42;
}

.ctrl-btn.prepare-btn {
  background: #f3c42a;
  color: #fff;
  border-color: #144f68;
}

.ctrl-btn.prepare-btn:hover {
  background: #c29e27;
}

.ctrl-btn.prepare-btn.waiting {
  background: #17594e;
  border-color: #0f3f38;
}

.ctrl-btn.prepare-btn.waiting:hover {
  background: #0f3f38;
}

/* 抽尾部：海蓝 */
.ctrl-btn.vacuum-tail-btn {
  background: #e9f4f9;
  border-color: #144f68;
  color: #144f68;
}

.ctrl-btn.vacuum-tail-btn:hover {
  background: #ddeef7;
}

.ctrl-btn.vacuum-tail-btn.active {
  background: #144f68;
  border-color: #0b3a4e;
  color: #fff;
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.25);
}

/* 抽靶室：绿色系 */
.ctrl-btn.vacuum-target-btn {
  background: #ebf6f3;
  border-color: #17594e;
  color: #17594e;
}

.ctrl-btn.vacuum-target-btn:hover {
  background: #d3ece5;
}

.ctrl-btn.vacuum-target-btn.active {
  background: #17594e;
  border-color: #0f3f38;
  color: #fff;
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.25);
}

/* 系统释放 */
.ctrl-btn.release-btn {
  background: #39426b;
  color: #fff;
  border-color: #2b3245;
}

.ctrl-btn.release-btn.active {
  background: #41b883;
  border-color: #1d7a4a;
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(65, 184, 131, 0.4); }
  50% { box-shadow: 0 0 0 10px rgba(65, 184, 131, 0); }
}

/* 切换开关 */
.toggle-switch {
  width: 48px;
  height: 26px;
  background: #fff;
  border-radius: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  border: 2px solid #e8eaed;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  flex-shrink: 0;
}

.toggle-switch.disabled {
  opacity: 0.5;
  cursor: not-allowed;
  pointer-events: none;
}

.toggle-switch::before {
  content: var(--off-text, '手动');
  position: absolute;
  right: 3px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 10px;
  font-weight: 600;
  color: #4e69b5;
  transition: all 0.3s ease;
}

.toggle-switch::after {
  content: '';
  position: absolute;
  width: 20px;
  height: 20px;
  background: #4e69b5;
  border-radius: 50%;
  top: 1px;
  left: 1px;
  transition: all 0.3s ease;
  box-shadow: 0 2px 6px rgba(78, 105, 181, 0.4);
}

.toggle-switch.active {
  background: #4e69b5;
  border-color: #4e69b5;
}

.toggle-switch.active::before {
  content: var(--on-text, '自动');
  right: auto;
  left: 4px;
  color: white;
}

.toggle-switch.active::after {
  left: 25px;
  background: white;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2);
}

/* 手动进气/排气按钮 */
.manual-controls {
  display: flex;
  gap: 6px;
  margin-top: 6px;
}

.mini-btn {
  flex: 1;
  padding: 10px;
  border: 2px solid #2b3245;
  border-radius: 8px;
  font-size: 16px;
  font-weight: bold;
  cursor: pointer;
  transition: 0.2s;
  box-sizing: border-box;
}

.mini-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.mini-btn.plus-btn {
  background: #ebf6f3;
  border-color: #17594e;
  color: #17594e;
}

.mini-btn.plus-btn:hover:not(:disabled) {
  background: #d3ece5;
}

.mini-btn.minus-btn {
  background: #fdf2f2;
  border-color: #a53838;
  color: #a53838;
}

.mini-btn.minus-btn:hover:not(:disabled) {
  background: #fce8e8;
}

/* 压力输入框 */
.pressure-input {
  padding: 10px 12px;
  border: 2px solid #2b3245;
  border-radius: 8px;
  font-size: 15px;
  width: 100%;
  box-sizing: border-box;
  outline: none;
  transition: 0.2s;
}

.pressure-input:focus {
  border-color: #4e69b5;
  box-shadow: 0 0 0 3px rgba(78, 105, 181, 0.18);
}

/* ===== 发射确认弹窗 ===== */
.confirm-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1500;
  backdrop-filter: blur(4px);
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.confirm-modal {
  background: #fff;
  border-radius: 12px;
  width: 380px;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  animation: slideUp 0.25s ease;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}

.confirm-header {
  background: #4e69b5;
  padding: 16px 20px;
  color: #fff;
}

.confirm-header.danger {
  background: #e5484d;
}

.confirm-header h3 {
  margin: 0;
  font-size: 17px;
  font-weight: 600;
}

.confirm-body {
  padding: 20px;
  font-size: 14px;
  color: #3c4043;
  line-height: 1.6;
}

.confirm-body p {
  margin: 0;
}

.confirm-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  padding: 14px 20px;
  background: #f8f9fa;
  border-top: 1px solid #e8eaed;
}

.confirm-btn {
  padding: 9px 22px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s ease;
}

.confirm-btn.cancel-btn {
  background: #e8eaed;
  color: #5f6368;
}

.confirm-btn.cancel-btn:hover {
  background: #dadce0;
  color: #3c4043;
}

.confirm-btn.ok-btn {
  background: #4e69b5;
  color: #fff;
  box-shadow: 0 2px 8px rgba(78, 105, 181, 0.3);
}

.confirm-btn.ok-btn:hover {
  background: #3d549a;
}

.confirm-btn.ok-btn.danger {
  background: #e5484d;
  box-shadow: 0 2px 8px rgba(229, 72, 77, 0.3);
}

.confirm-btn.ok-btn.danger:hover {
  background: #c73e42;
}

/* ===== 底部状态栏 ===== */
.footer {
  height: 40px;
  line-height: 40px;
  text-align: center;
  font-size: 14px;
  font-weight: bold;
  background-color: #4e69b5;
  color: #fff;
}

.footer-alarm {
  color: #ffd2d2;
}
</style>
