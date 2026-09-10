<template>
  <div class="content">
      <!-- 左侧：状态 / 连接 / 日志 -->
      <div class="left-panel">
        <div class="side-card status-panel">
          <div class="status-grid">
            <div class="indicator-item">
              <div class="led led-connect" :class="{ 'active': System.connected }"></div>
              <span class="indicator-label">PLC连接</span>
            </div>
            <div class="indicator-item">
              <div class="led led-alarm" :class="{ 'active': System.alarm != '' && System.connected }"></div>
              <span class="indicator-label">异常警报</span>
            </div>
          </div>
        </div>

        <div class="side-card config-panel">
          <input v-model="Target.ip" type="text" class="ip-input-modern" placeholder="PLC IP Address" />
          <div class="conn-btns-group">
            <button class="btn btn-connect" @click="handleConnect">{{System.connected ? '断开连接' : '连接设备'}}</button>
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
                <span class="value">{{ System.tube.toFixed(2) }}</span>
                <span class="unit">MPa</span>
                </div>
            </div>
            <div class="metric-card">
                <span class="label">气室压力</span>
                <div class="value-row">
                <span class="value">{{ System.pressure.toFixed(2) }}</span>
                <span class="unit">MPa</span>
                </div>
            </div>
            <div class="metric-card">
                <span class="label">靶室真空度</span>
                <div class="value-row">
                <span class="value">{{ System.vacuum.toFixed(0) }}</span>
                <span class="unit">Pa</span>
                </div>
            </div>
            <div class="metric-card">
                <span class="label">运行时长</span>
                <div class="value-row">
                <span class="value">{{ System.time }}</span>
                </div>
            </div>

        </section>

        <section class="device-visualization">
          <div class="schematic-view">
            <div class="led-list">
              <div v-for="led in ledItems" :key="led.key" class="led-item" :class="{ on: Y[led.key] && System.connected }">
                <span class="led-dot"></span>
                <span class="led-text">{{ led.label }}</span>
              </div>
            </div>
            <img src="../../assets/images/devices/gasgun1.png" alt="气炮示意图" class="cannon-image" draggable="false" />
          </div>
        </section>

        <footer class="action-bar">
          <!-- 增压泵 -->
          <button class="act-btn pump" :class="{ 'active-pressing': Y.PressureOpen }"
            @click="handlePressureSwitch(!Y.PressureOpen)">
            {{ Y.PressureOpen ? '停增压泵' : '增压泵' }}
          </button>

          <!-- 目标压力输入 + 自动增压 -->
          <div class="btn-column">
            <span class="target-label">目标气压（MPa）</span>
            <input v-model="targetPressureView" @blur="formatTargetPressure" class="target-input" inputmode="decimal" placeholder="0.00" />
            <button class="act-btn auto" :class="{ 'active-pressing': Target.autoRunning }" @click="handleAutoPressurize">
              {{ Target.autoRunning ? '自动增压中' : '自动增压' }}
            </button>
          </div>

          <!-- 进气 / 出气（按住生效） -->
          <div class="btn-column">
            <button class="act-btn" :class="{ 'active-pressing': Y.Inlet }"
              @mousedown="handleInletSwitch(true)" @mouseup="handleInletSwitch(false)">
              {{ Y.Inlet ? '进气中...' : '进气' }}
            </button>
            <button class="act-btn valve" :class="{ 'active-pressing': Y.Outlet }"
              @mousedown="handleOutletSwitch(true)" @mouseup="handleOutletSwitch(false)">
              {{ Y.Outlet ? '出气中...' : '出气' }}
            </button>
          </div>

          <button class="act-btn vacuum" :class="{ 'active-pressing': Y.TailVacuumPump }" @click="handleTailVacuumSwitch(!Y.TailVacuumPump)">
            {{ Y.TailVacuumPump ? '停抽尾部' : '抽尾部' }}
          </button>
          <button class="act-btn vacuum" :class="{ 'active-pressing': Y.TarVacuumPump }" @click="handleVacuumSwitch(!Y.TarVacuumPump)">
            {{ Y.TarVacuumPump ? '停抽靶室' : '抽靶室' }}
          </button>

          <button class="act-btn fire" @click="handleFire">发射</button>
        </footer>
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
  VacuumSwitch,
  TailVacuumSwitch,
  FireSwitch,
  PressureSwitch,
  AutoPressurize,
  StopAutoPressurize,
  ExportAlarmHistory,
} from '../../../wailsjs/go/xinjiegasgun1/XinjieGasGun1'



const System = reactive({
  connected:    false,
  alarm:        '',

  tube:         0.00,
  pressure:     0.00,
  vacuum:       100000,

  time:         '00:00:00',
})

const Y = ref({
    Inlet:          false,
    Outlet:         false,
    Fire:           false,
    VacuumRealse:   false,
    PressureOpen:   false,
    PressureClose:  false,
    TailVacuumPump: false,
    TarVacuumPump:  false,
})

// 示意图左上角 LED 输出指示
const ledItems = [
    { key: 'Inlet',          label: '进气阀' },
    { key: 'Outlet',         label: '出气阀' },
    { key: 'Fire',           label: '发射阀' },
    { key: 'VacuumRealse',   label: '尾部真空阀' },
    { key: 'PressureOpen',   label: '增压泵开结点' },
    { key: 'PressureClose',  label: '增压泵关结点' },
    { key: 'TailVacuumPump', label: '尾部真空泵' },
    { key: 'TarVacuumPump',  label: '靶室真空泵' },
]

const Target = reactive({
    ip:         '192.168.6.6',
    pressure:    1.00,
    autoRunning:  false,
})

// 目标压力输入：失焦时格式化保留两位小数
const targetPressureView = ref(Target.pressure.toFixed(2))
function formatTargetPressure() {
  const n = Number(targetPressureView.value)
  if (!isFinite(n) || n < 0) {
    targetPressureView.value = '0.00'
    Target.pressure = 0
    return
  }
  Target.pressure = Math.round(n * 100) / 100
  targetPressureView.value = Target.pressure.toFixed(2)
}


const logs = ref([])

onMounted(() => {
  EventsOn('heartbeat', (info) => {
    if (!info) return
    try {
        System.connected = info.running
        System.alarm = info.alarm
        System.pressure = info.pressure
        System.vacuum = info.vacuum
        // 后端 JSON tag 为小写 output；逐字段赋值避免整体替换丢失键
        const out = info.output || {}
        Y.value = {
        Inlet:          !!out.Inlet,
        Outlet:         !!out.Outlet,
        Fire:           !!out.Fire,
        VacuumRealse:   !!out.VacuumRealse,
        PressureOpen:   !!out.PressureOpen,
        PressureClose:  !!out.PressureClose,
        TailVacuumPump: !!out.TailVacuumPump,
        TarVacuumPump:  !!out.TarVacuumPump,
        }
        System.time = info.time
    } catch (e) {
        addLog(`解析心跳数据失败: ${e}`)
    }
    
  })
  EventsOn('message', (msg) => {
    addLog(msg)
  })
})

onUnmounted(() => {
  EventsOff('heartbeat')
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
  if (!System.connected) {
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
    await VacuumSwitch(open)
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

// 增压泵开关
async function handlePressureSwitch(open) {
  try {
    await PressureSwitch(open)
    addLog(`${open ? '开启' : '关闭'}增压泵`)
  } catch (e) {
    addLog(`增压泵操作失败: ${e}`)
  }
}

// 自动增压（目标压力 Target.pressure）
function handleAutoPressurize() {
  if (Target.autoRunning) {
    Target.autoRunning = false
    StopAutoPressurize()
      .then(() => addLog('停止自动增压'))
      .catch((e) => addLog(`停止自动增压失败: ${e}`))
    return
  }
  const t = Math.round(Number(targetPressureView.value) * 100) / 100
  if (!isFinite(t) || t <= 0) {
    addLog('请输入有效的目标压力')
    return
  }
  Target.autoRunning = true
  addLog(`开始自动增压，目标 ${t.toFixed(2)} MPa`)
  AutoPressurize(t)
    .then(() => { Target.autoRunning = false })
    .catch((e) => {
      Target.autoRunning = false
      addLog(`自动增压失败: ${e}`)
    })
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
  height: 130px;
  display: flex;
  gap: 15px;
  align-items: stretch;
  background: #dfe6f3;
  padding: 10px;
  border-radius: 8px;
}

.act-btn {
  flex: 1;
  min-width: 0;
  padding: 8px;
  box-sizing: border-box;
  border: 2px solid #2b3245;
  background: #f8f9fa;
  border-radius: 8px;
  font-weight: bold;
  cursor: pointer;
  transition: 0.2s;
  font-size: 1.3rem;
  box-shadow: 0 2px 5px rgba(15, 23, 42, 0.15);
}

.act-btn.active-pressing {
  background: #dbe6fa;
  border-color: #1f2a47;
  color: #24397a;
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.15);
}

.act-btn:hover {
  background: #eef2fb;
  transform: translateY(-1px);
  box-shadow: 0 4px 9px rgba(15, 23, 42, 0.22);
}

.act-btn.fire {
  background: #e5484d;
  color: #fff;
  border-color: #144f68;
}

.act-btn.fire:hover {
  background: #c73e42;
}

/* ===== 分组配色：浅色底区分功能区 ===== */
/* 进气/出气：主蓝（白底） */

/* 增压泵：紫 */
.act-btn.pump {
  background: #f2f0fa;
  border-color: #4e4680;
  color: #403873;
}

.act-btn.pump:hover {
  background: #eae7f6;
}

.act-btn.pump.active-pressing {
  background: #e0dbf3;
  border-color: #322a63;
  color: #322a63;
}

/* 自动/目标压力：青绿 */
.act-btn.auto {
  background: #ebf6f3;
  border-color: #17594e;
  color: #17594e;
}

.act-btn.auto:hover {
  background: #e0f1ec;
}

.act-btn.auto.active-pressing {
  background: #d3ece5;
  border-color: #0e453d;
  color: #0e453d;
}

/* 抽尾部/抽靶室：海蓝 */
.act-btn.vacuum {
  background: #e9f4f9;
  border-color: #144f68;
  color: #144f68;
}

.act-btn.vacuum:hover {
  background: #ddeef7;
}

.act-btn.vacuum.active-pressing {
  background: #cfe7f2;
  border-color: #0b3a4e;
  color: #0b3a4e;
}

/* 双行按钮列 */
.btn-column {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.btn-column .act-btn {
  padding: 4px;
  font-size: 1.2rem;
}

/* 目标压力标签 */
.target-label {
  flex-shrink: 0;
  font-size: 0.85rem;
  font-weight: bold;
  color: #1d6e60;
  text-align: center;
  line-height: 1.2;
}

/* 目标压力输入框 */
.target-input {
  flex: 1;
  min-width: 0;
  padding: 4px 10px;
  border: 2px solid #17594e;
  border-radius: 8px;
  background: #ffffff;
  text-align: center;
  font-size: 1.3rem;
  font-weight: bold;
  color: #1a1c24;
  outline: none;
  transition: 0.2s;
  box-sizing: border-box;
}

.target-input:focus {
  border-color: #0e453d;
  box-shadow: 0 0 0 3px rgba(46, 158, 138, 0.18);
}

.target-input::-webkit-outer-spin-button,
.target-input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.target-input {
  -moz-appearance: textfield;
  appearance: textfield;
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
