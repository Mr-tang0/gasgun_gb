<template>
  <div class="gasgun-container">
    <aside class="sidebar">
      <div class="sidebar-header">
        <div class="brand-box">
          <h2 class="brand-title">二级轻气炮控制系统</h2>
          <p class="subtitle">Two-Stage Gas Gun Control System</p>
        </div>
      </div>

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
          <button class="btn btn-connect" @click="handleConnect(true)" :disabled="System.connected">
            连接设备
          </button>
          <button class="btn btn-disconnect" @click="handleConnect(false)" :disabled="!System.connected">
            断开
          </button>
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

    </aside>

    <main class="main-content">
      <section class="metrics-grid">
        <div class="metric-card">
          <span class="label">输入气压 (MPa)</span>
          <span class="main-value">{{ System.InputPressure.toFixed(2) }}</span>
        </div>
        <div class="metric-card">
          <span class="label">尾部真空度 (Pa)</span>
          <span class="main-value">{{ System.TailVacuumDegree.toFixed(1) }}</span>
        </div>
        <div class="metric-card">
          <span class="label">一级气室气压 (MPa)</span>
          <span class="main-value">{{ System.CylinderPressure.toFixed(2) }}</span>
        </div>
        <div class="metric-card pump-tube-card">
          <span class="label">{{ pumpTubeLabel }}</span>
          <div class="value-area">
            <span class="main-value">{{ pumpTubeMainValue }}</span>
            <button
              class="toggle-btn"
              @mousedown="PumpTubePrecision('high')"
              @mouseup="PumpTubePrecision('low')"
              @mouseleave="PumpTubePrecision('low')"
            >切换</button>
          </div>
          <span class="sub-value">{{ pumpTubeSubValue }}</span>
        </div>
        <div class="metric-card">
          <span class="label">靶室真空度 (Pa)</span>
          <span class="main-value">{{ System.TargetVacuumDegree.toFixed(1) }}</span>
        </div>
      </section>

      <section class="device-visualization">
        <div class="schematic-view">
          <img src="../../assets/images/devices/gasgun2.png" alt="Gasgun2示意图" class="cannon-image" draggable="false"/>
        </div>

        <!-- 阀门状态悬浮窗 -->
        <div class="valve-float-panel">
          <button
            v-for="v in valveButtons"
            :key="v.key"
            class="valve-btn"
            :class="{ 'active': v.active }"true
            @click="toggleValveButton(v.key)"
          >{{ v.label }}</button>
        </div>
      </section>

      <footer class="action-bar">

        <!-- 1. 真空操作 -->
        <div class="action-zone">
          <div class="zone-title">真空操作</div>
          <div class="btn-column">
            <button class="ctrl-btn vacuum-target-btn" 
            :class="{ 'active': System.VacuumRunning }" 
            @click="toggleVacuum">
              {{ System.VacuumRunning ? '停止抽真空' : '开始抽真空' }}
            </button>
            <button class="ctrl-btn vacuum-pump-btn" 
            :class="{ 'active': System.PumpTubeVacuumRunning }" 
            @click="togglePumpTubeVacuum">
              {{ System.PumpTubeVacuumRunning ? '停止抽泵管' : '开始抽泵管' }}
            </button>
          </div>
        </div>

        <!-- 2. 气瓶压力控制 -->
        <div class="action-zone">
          <div class="zone-title">气瓶压力控制</div>
          <div class="mode-row">
            <span class="mode-label">模式：{{ Target.CylinderAutoMode ? '自动' : '手动' }}</span>
            <div class="toggle-switch" 
            :class="{ 'active': Target.CylinderAutoMode }" 
            @click="Target.CylinderAutoMode = !Target.CylinderAutoMode"></div>
          </div>

          <div v-if="Target.CylinderAutoMode" class="auto-group">
            <div class="auto-fields">
              <label class="auto-label">目标 (MPa)</label>
              <input v-model.number="Target.CylinderPressure" 
              type="number" step="0.1" class="pressure-input" />
            </div>
            <button class="ctrl-btn start-btn" @click="toggleCylinderAutoMode">开始</button>
          </div>
          <div v-else class="manual-controls">
            <button class="mini-btn plus-btn" 
              @mousedown="manualPressurize(true)" @mouseup="manualPressurize(false)">进气</button>
            <button class="mini-btn minus-btn"
              @mousedown="manualDecompress(true)" @mouseup="manualDecompress(false)">排气</button>
          </div>
        </div>

        <!-- 3. 泵管压力控制 -->
        <div class="action-zone">
          <div class="zone-title">泵管压力控制</div>
          <div class="mode-row">
            <span class="mode-label">模式：{{ Target.PumpTubeAutoMode ? '自动' : '手动' }}</span>
            <div class="toggle-switch" 
            :class="{ 'active': Target.PumpTubeAutoMode }" 
            @click="Target.PumpTubeAutoMode = !Target.PumpTubeAutoMode"></div>
          </div>
          <div v-if="Target.PumpTubeAutoMode" class="auto-group">
            <div class="auto-fields">
              <label class="auto-label">目标 (MPa)</label>
              <input v-model.number="pumpTubeTargetPressure" type="number" step="0.1" class="pressure-input" />
            </div>
            <button class="ctrl-btn start-btn" @click="togglePumpTubeAutoMode">开始</button>
          </div>
          <div v-else class="manual-controls">
            <button class="mini-btn plus-btn"
              @mousedown="manualPumpTubePressurize(true)" @mouseup="manualPumpTubePressurize(false)">进气</button>
            <button class="mini-btn minus-btn"
              @mousedown="manualPumpTubeDecompress(true)" @mouseup="manualPumpTubeDecompress(false)">排气</button>
          </div>
        </div>

        <!-- 4. 发射控制 -->
        <div class="action-zone">
          <div class="zone-title">发射控制</div>
          <div class="mode-row">
            <span class="mode-label">模式：{{ Target.ExternalTrigger ? '外触发' : '内触发' }}</span>
            <div class="toggle-switch" 
            :class="{ 'active': Target.ExternalTrigger }" 
            style="--off-text: '内触发'; --on-text: '外触发'" 
            @click="Target.ExternalTrigger = !Target.ExternalTrigger"></div>
          </div>
          <button class="ctrl-btn fire-btn" :disabled="Target.ExternalTrigger" @click="handleFire">立即发射</button>
        </div>

        <!-- 5. 系统恢复 -->
        <div class="action-zone">
          <div class="zone-title">系统恢复</div>
          <button class="ctrl-btn reset-btn" :class="{ 'active': isResetting }" @click="handleReset">{{ isResetting ? '恢复中...' : '系统恢复' }}</button>
        </div>
      </footer>
    </main>
  </div>

  <!-- 发射倒计时模态框 -->
  <div v-if="showCountdown" class="countdown-overlay">
    <div class="countdown-container">
      <div class="countdown-ring">
        <div class="countdown-number">{{ countdown }}</div>
      </div>
      <div class="countdown-text">发射倒计时</div>
      <div class="countdown-hint">按下任意键取消发射</div>
    </div>
  </div>

</template>


<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime'

import {
  ConnectDevice,
  DisconnectDevice,

  StartAutoVacuum,
  StopAutoVacuum,

  StartPumpTubeVacuum,
  StopPumpTubeVacuum,

  AutoPumpTubePressure,
  StopAutoPumpTubePressure,

  AutoCylinderPressure,
  StopAutoCylinderPressure,

  PrepareFire,
  Fire,
  ResetSystem,
  SetTriggerMode,
  ManualPumpTubePressurize,
  ManualPumpTubeDecompress,
  ManualPressurize,
  ManualDecompress,
  CloseSwitch,
  OpenSwitch,
} from '../../../wailsjs/go/xinjiegasgun2/XinjieGasGun2'


// ===== 系统状态 =====
const System = reactive({
  connected:             false,
  alarm:                 '',

  InputPressure:         0.00,
  CylinderPressure:      0.00,
  PumpTubePressure:      0.00,
  PumpTubePressureHi:    0.000,

  TargetVacuumDegree:    100000,
  TailVacuumDegree:      100000,

  time:                  '00:00:00',

  VacuumRunning:         false,
  PumpTubeVacuumRunning: false,
  PressureAutoRunning:   false,
  PumpTubeAutoRunning: false,
})

const Target = reactive({
  ip:                  '192.168.6.6',

  CylinderAutoMode:    false,
  PumpTubeAutoMode:    false,
  ExternalTrigger:      false,

  CylinderPressure:   1.00,
  PumpTubePressure:   1.00,
})

const valveButtons = reactive([
  { key: 'Pressurize',         label: '气瓶增压阀', active: false },
  { key: 'Decompress',         label: '气瓶减压阀', active: false },
  { key: 'PumpTubePressurize', label: '泵管增压阀', active: false },
  { key: 'PumpTubeDecompress', label: '泵管减压阀', active: false },
  { key: 'PumpTubeVacuum',     label: '泵管真空阀', active: false },
  { key: 'TargetVacuum',       label: '靶室真空阀', active: false },
  { key: 'TailVacuumProtect',  label: '尾部真空阀', active: false },
  { key: 'PumpTubeProtect',    label: '泵管保护阀', active: false },
  { key: 'FireSwitch',         label: '系统发射阀', active: false },
  { key: 'SystemDecompress',   label: '系统排气阀', active: false },
  { key: 'TailVacuumPump',     label: '尾部真空泵', active: false },
  { key: 'TargetVacuumPump',   label: '靶室真空泵', active: false },
])



onMounted(() => {
  EventsOn('xinjie_gasgun2_heartbeat', (info) => {
    if (!info) return
    try {
        System.connected = info.Running
        System.alarm = info.Alarm

        System.InputPressure = info.InputPressure
        System.CylinderPressure = info.CylinderPressure
        System.PumpTubePressure = info.PumpTubePressure
        System.PumpTubePressureHi = info.PumpTubePressureHi

        System.TargetVacuumDegree = info.TargetVacuumDegree
        System.TailVacuumDegree = info.TailVacuumDegree

        System.time = info.Time


        valveButtons.forEach(v => {
          v.active = info.out[v.key]
        })
    } catch (e) {
      addLog(`解析心跳数据失败: ${e}`)
    }
  })

  EventsOn('xinjie_gasgun2_message', (msg) => {
    addLog(msg)
  })
})


// ===== 日志 =====
const logs = ref([])

const addLog = (msg) => {
  const now = new Date()
  const time = `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`
  logs.value.push({ time, msg })
  if (logs.value.length > 200) logs.value.shift()
}

const handleClearLogs = () => {
  if (!logs.value.length) return
  logs.value = []
  addLog('日志已清除')
}

const handleExportLogs = () => {
  if (!logs.value.length) {
    addLog('暂无日志可导出')
    return
  }
  addLog('日志导出功能待实现')
}

async function handleConnect(open) {
  try {
    if (open) {
      await ConnectDevice(Target.ip)
    } else {
      await DisconnectDevice()
    }
    System.connected = open
    addLog(`连接已切换: ${open}`)
  } catch (e) {
    addLog(`操作失败: ${e}`)
  }
}


// ===== 泵管精度切换 =====
const pumpTubeIsHighPrecision = ref(false)
const pumpTubeLabel = ref('二级泵管气压 (MPa)')
const pumpTubeMainValue = ref('0.00')
const pumpTubeSubValue = ref('0.000')
const PumpTubePrecision = async (key) => {
  if (key === 'high') {
    try { await OpenSwitch('PumpTubeProtect') } catch (e) { addLog(`操作失败: ${e}`) }
    pumpTubeIsHighPrecision.value = true
    pumpTubeLabel.value = '二级泵管气压 (高精度)'
    pumpTubeMainValue.value = System.PumpTubePressureHi.toFixed(3)
    pumpTubeSubValue.value = System.PumpTubePressure.toFixed(2)
  } else {
    try { await CloseSwitch('PumpTubeProtect') } catch (e) { addLog(`操作失败: ${e}`) }
    pumpTubeIsHighPrecision.value = false
    pumpTubeLabel.value = '二级泵管气压 (MPa)'
    pumpTubeMainValue.value = System.PumpTubePressure.toFixed(2)
    pumpTubeSubValue.value = System.PumpTubePressureHi.toFixed(3)
  }
}

// ===== 阀门切换 =====
async function toggleValveButton(v) {
  try {
    if (v.active) {
      await CloseSwitch(v.key)
    } else {
      await OpenSwitch(v.key)
    }
    v.active = !v.active
    addLog(`${v.label} ${v.active ? '开启' : '关闭'}`)
  } catch (e) {
    addLog(`操作失败: ${e}`)
  }
}

async function toggleVacuum() {
  try {
    if (System.VacuumRunning) {
      await StopAutoVacuum()
    } else {
      await StartAutoVacuum()
    }
    System.VacuumRunning = !System.VacuumRunning
    addLog(`真空运行已切换: ${System.VacuumRunning}`)
  } catch (e) {
    addLog(`操作失败: ${e}`)
  }
}

async function togglePumpTubeVacuum() {
  try {
    if (System.PumpTubeVacuumRunning) {
      await StopPumpTubeVacuum()
    } else {
      await StartPumpTubeVacuum()
    }
    System.PumpTubeVacuumRunning = !System.PumpTubeVacuumRunning
    addLog(`泵管真空运行已切换: ${System.PumpTubeVacuumRunning}`)
  } catch (e) {
    addLog(`操作失败: ${e}`)
  }
}

async function toggleCylinderAutoMode() {
  try {
    if (System.PressureAutoRunning) {
      await StopAutoCylinderPressure()
    } else {
      await AutoCylinderPressure(Target.CylinderPressure.value)
    }
  } catch (e) {
    addLog(`操作失败: ${e}`)
  }
}

async function togglePumpTubeAutoMode() {
  try {
    if (System.PumpTubeAutoRunning) {
      await StopAutoPumpTubePressure()
    } else {
      await AutoPumpTubePressure(Target.PumpTubePressure.value)
    }
  } catch (e) {
    addLog(`操作失败: ${e}`)
  }
}

async function handleFire() {
  try {
    await PrepareFire()
    await Fire()
    addLog('发射已开启')
  } catch (e) {
    addLog(`操作失败: ${e}`)
  }
}


// // ===== 控制状态 =====
// const vacuumRunning = ref(false)
// const pumpTubeVacuumRunning = ref(false)
// const isResetting = ref(false)
// const isExternalTrigger = ref(false)
// const pumpTubeAutoMode = ref(false)
// const cylinderAutoMode = ref(false)
// const pumpTubeTargetPressure = ref(2.0)
// const cylinderTargetPressure = ref(1.0)
// const showCountdown = ref(false)
// const countdown = ref(3)













// // ===== 事件处理 =====
// const handleConnect = async (open) => {
//   if (open) {
//     try {
//       await ConnectDevice(Target.ip)
//       addLog(`尝试连接 ${Target.ip}`)
//     } catch (e) {
//       addLog(`连接失败: ${e}`)
//     }
//   } else {
//     try {
//       await DisconnectDevice()
//       addLog('断开连接')
//     } catch (e) {
//       addLog(`断开失败: ${e}`)
//     }
//   }
// }

// const toggleVacuum = async () => {
//   if (vacuumRunning.value) {
//     try { await StopAutoVacuum(); vacuumRunning.value = false; addLog('停止抽真空') } catch (e) { addLog(`操作失败: ${e}`) }
//   } else {
//     try { await StartAutoVacuum(); vacuumRunning.value = true; addLog('开始抽真空') } catch (e) { addLog(`操作失败: ${e}`) }
//   }
// }

// const togglePumpTubeVacuum = async () => {
//   if (pumpTubeVacuumRunning.value) {
//     try { await StopPumpTubeVacuum(); pumpTubeVacuumRunning.value = false; addLog('停止抽泵管') } catch (e) { addLog(`操作失败: ${e}`) }
//   } else {
//     try { await StartPumpTubeVacuum(); pumpTubeVacuumRunning.value = true; addLog('开始抽泵管') } catch (e) { addLog(`操作失败: ${e}`) }
//   }
// }

// const toggleCylinderAutoMode = async () => {
//   if (cylinderAutoMode.value) {
//     try { await StopAutoCylinderPressure(); addLog('气瓶自动控制已停止') } catch (e) { addLog(`操作失败: ${e}`) }
//   } else {
//     try { await AutoCylinderPressure(cylinderTargetPressure.value); addLog(`气瓶自动控制, 目标 ${cylinderTargetPressure.value} MPa`) } catch (e) { addLog(`操作失败: ${e}`) }
//   }
// }

// const setCylinderMode = async (auto) => {
//   if (auto) {
//     cylinderAutoMode.value = true
//   } else {
//     if (cylinderAutoMode.value) {
//       try { await StopAutoCylinderPressure(); addLog('气瓶自动控制已停止') } catch (e) { addLog(`操作失败: ${e}`) }
//     }
//     cylinderAutoMode.value = false
//   }
// }

// const togglePumpTubeAutoMode = async () => {
//   if (pumpTubeAutoMode.value) {
//     try { await StopAutoPumpTubePressure(); addLog('泵管自动控制已停止') } catch (e) { addLog(`操作失败: ${e}`) }
//   } else {
//     try { await AutoPumpTubePressure(pumpTubeTargetPressure.value); addLog(`泵管自动控制, 目标 ${pumpTubeTargetPressure.value} MPa`) } catch (e) { addLog(`操作失败: ${e}`) }
//   }
// }

// const setPumpTubeMode = async (auto) => {
//   if (auto) {
//     pumpTubeAutoMode.value = true
//   } else {
//     if (pumpTubeAutoMode.value) {
//       try { await StopAutoPumpTubePressure(); addLog('泵管自动控制已停止') } catch (e) { addLog(`操作失败: ${e}`) }
//     }
//     pumpTubeAutoMode.value = false
//   }
// }

// const manualPressurize = async (enable) => {
//   try { await ManualPressurize(enable); addLog(enable ? '气瓶加气' : '气瓶停止加气') } catch (e) { addLog(`操作失败: ${e}`) }
// }

// const manualDecompress = async (enable) => {
//   try { await ManualDecompress(enable); addLog(enable ? '气瓶泄气' : '气瓶停止泄气') } catch (e) { addLog(`操作失败: ${e}`) }
// }

// const manualPumpTubePressurize = async (enable) => {
//   try { await ManualPumpTubePressurize(enable); addLog(enable ? '泵管加气' : '泵管停止加气') } catch (e) { addLog(`操作失败: ${e}`) }
// }

// const manualPumpTubeDecompress = async (enable) => {
//   try { await ManualPumpTubeDecompress(enable); addLog(enable ? '泵管泄气' : '泵管停止泄气') } catch (e) { addLog(`操作失败: ${e}`) }
// }

// const setTriggerMode = (isExternal) => {
//   isExternalTrigger.value = isExternal
//   try { SetTriggerMode(isExternal) } catch (e) { /* 后端未就绪 */ }
//   addLog(`触发模式: ${isExternal ? '外触发' : '内触发'}`)
// }

// const handleFire = async () => {
//   try {
//     await Fire()
//     addLog('执行发射')
//   } catch (e) {
//     addLog(`发射失败: ${e}`)
//   }
// }

// const handleReset = async () => {
//   if (isResetting.value) {
//     try { await ResetSystem(false); isResetting.value = false; addLog('系统重置完成') } catch (e) { addLog(`操作失败: ${e}`) }
//   } else {
//     try { await ResetSystem(true); isResetting.value = true; addLog('系统恢复中') } catch (e) { addLog(`操作失败: ${e}`) }
//   }
// }


// // ===== 生命周期 =====
// onMounted(() => {
//   EventsOn('heartbeat', (info) => {
//     if (!info) return
//     try {
//       System.connected = info.running
//       System.alarm = info.alarm || ''
//       if (info.inputPressure !== undefined) System.InputPressure = info.inputPressure
//       if (info.cylinderPressure !== undefined) System.CylinderPressure = info.cylinderPressure
//       if (info.pumpTubePressure !== undefined) System.PumpTubePressure = info.pumpTubePressure
//       if (info.pumpTubePressureHi !== undefined) System.PumpTubePressureHi = info.pumpTubePressureHi
//       if (info.targetVacuumDegree !== undefined) System.TargetVacuumDegree = info.targetVacuumDegree
//       if (info.tailVacuumDegree !== undefined) System.TailVacuumDegree = info.tailVacuumDegree
//       System.time = info.time || '00:00:00'

//       if (pumpTubeIsHighPrecision.value) {
//         pumpTubeMainValue.value = System.PumpTubePressureHi.toFixed(3)
//         pumpTubeSubValue.value = System.PumpTubePressure.toFixed(2)
//       } else {
//         pumpTubeMainValue.value = System.PumpTubePressure.toFixed(2)
//         pumpTubeSubValue.value = System.PumpTubePressureHi.toFixed(3)
//       }
//     } catch (e) {
//       addLog(`解析心跳数据失败: ${e}`)
//     }
//   })

//   EventsOn('message', (msg) => {
//     addLog(msg)
//   })
// })

// onUnmounted(() => {
//   EventsOff('heartbeat')
//   EventsOff('message')
// })
</script>

<style scoped>
/* ===== 根容器（来自 gasgun1 .content） ===== */
.gasgun-container {
  flex: 1;
  display: flex;
  flex-direction: row;
  background-color: #edf1f8;
  min-height: 0;
  height: 100%;
  overflow: hidden;
}

/* ===== 左侧栏（来自 gasgun1 .left-panel） ===== */
.sidebar {
  width: 270px;
  background: #232c49;
  display: flex;
  flex-direction: column;
  padding: 16px;
  box-shadow: 4px 0 15px rgba(0, 0, 0, 0.2);
  box-sizing: border-box;
  z-index: 10;
}

.sidebar-header {
  margin-bottom: 16px;
  padding: 0 8px;
}

.brand-title {
  font-size: 18px;
  font-weight: 500;
  color: #ffffff;
}

.subtitle {
  font-size: 10px;
  color: #7fb3ff;
  text-transform: uppercase;
  margin-top: 4px;
}

/* ===== 卡片通用（来自 gasgun1） ===== */
.side-card {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 16px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

/* ===== 状态指示（来自 gasgun1） ===== */
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

/* ===== IP输入 / 连接按钮（来自 gasgun1） ===== */
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
  font-size: 15px;
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

/* ===== 日志区（来自 gasgun1） ===== */
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
  font-size: 1.05rem;
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

/* ===== 侧边栏底部（来自 gasgun1） ===== */
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
  font-size: 15px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-settings-modern:hover {
  color: #fff;
  background: #4a5580;
}

/* ===== 主内容区（来自 gasgun1 .middle-panel） ===== */
.main-content {
  flex: 1;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
}

/* ===== 数值卡片（来自 gasgun1） ===== */
.metrics-grid {
  flex-shrink: 0;
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 15px;
}

.metric-card {
  background: #fff;
  padding: 20px;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  border-top: 4px solid #4e69b5;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.metric-card .label {
  font-size: 14px;
  color: #666;
}

.metric-card .main-value {
  font-size: 28px;
  font-weight: bold;
  color: #1a1c24;
}

/* 泵管卡片 */
.pump-tube-card {
  border-top-color: #4e69b5;
}

.pump-tube-card .sub-value {
  display: block;
  font-size: 14px;
  color: #888;
}

.pump-tube-card .value-area {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pump-tube-card .toggle-btn {
  padding: 4px 8px;
  font-size: 12px;
  border: 1px solid #4e69b5;
  border-radius: 4px;
  background: #fff;
  color: #4e69b5;
  cursor: pointer;
  transition: all 0.2s;
}

.pump-tube-card .toggle-btn:hover {
  background: #4e69b5;
  color: #fff;
}

/* ===== 设备示意图（来自 gasgun1） ===== */
.device-visualization {
  position: relative;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-radius: 12px;
  padding: 10px;
  margin-bottom: 0;
  overflow: hidden;
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

.cannon-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

/* 底部LED */
.footer-leds {
  width: 100%;
  padding-top: 16px;
  border-top: 1px solid #e9ecef;
}

.led-row {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-bottom: 8px;
}

.led-row:last-child {
  margin-bottom: 0;
}

.footer-led {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #e9ecef;
  border: 1px solid #dee2e6;
  transition: all 0.2s;
}

.footer-led.active {
  background: #4e69b5;
  border-color: #4e69b5;
  box-shadow: 0 0 8px rgba(78, 105, 181, 0.6);
}

/* ===== 阀门状态悬浮窗 ===== */
.valve-float-panel {
  position: absolute;
  left: 10px;
  right: 10px;
  bottom: 14px;
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-start;
  gap: 8px 6px;
  padding: 10px;
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(6px);
  border: 1px solid #e0e5ef;
  border-radius: 12px;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.15);
  z-index: 5;
  box-sizing: border-box;
}

.valve-btn {
  flex: 0 0 96px;
  width: 96px;
  padding: 9px 0;
  background: #ffffff;
  border: 2px solid #dee2e6;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  color: #444;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  white-space: nowrap;
  text-align: center;
  box-sizing: border-box;
}

.valve-btn:hover {
  border-color: #41b883;
  color: #1d7a4a;
}

.valve-btn.active {
  background: #41b883;
  border-color: #1d7a4a;
  color: #ffffff;
  box-shadow: 0 0 8px rgba(65, 184, 131, 0.5);
}

/* ===== 底部操作栏（来自 gasgun1 .action-bar） ===== */
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

/* 按钮纵向排列铺满 */
.btn-column {
  display: flex;
  flex-direction: column;
  gap: 10px;
  flex: 1;
}

.btn-column .ctrl-btn {
  flex: 1;
  min-height: 48px;
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

/* 自动模式：左（说明文字+输入框）右（开始按钮，等高铺满） */
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

/* 通用 flex:1 按钮（发射/恢复等） */
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

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: #1a1c24;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-title::before {
  content: '';
  width: 4px;
  height: 18px;
  background: #4e69b5;
  border-radius: 2px;
}

.step-number {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  background: #4e69b5;
  color: white;
  font-size: 12px;
  font-weight: 600;
  border-radius: 50%;
  flex-shrink: 0;
}

/* ===== 控制按钮（来自 gasgun1 .act-btn 风格） ===== */
.control-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

/* 双行按钮列 */
.btn-column {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

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

/* 抽真空（靶室）：绿色系 */
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

/* 抽泵管：蓝色系 */
.ctrl-btn.vacuum-pump-btn {
  background: #eef2fb;
  border-color: #35497e;
  color: #35497e;
}

.ctrl-btn.vacuum-pump-btn:hover {
  background: #dbe6fa;
}

.ctrl-btn.vacuum-pump-btn.active {
  background: #35497e;
  border-color: #24345c;
  color: #fff;
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.25);
}

.ctrl-btn.reset-btn {
  background: #39426b;
  color: #fff;
  border-color: #2b3245;
}

.ctrl-btn.reset-btn.active {
  background: #41b883;
  border-color: #1d7a4a;
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(65, 184, 131, 0.4); }
  50% { box-shadow: 0 0 0 10px rgba(65, 184, 131, 0); }
}

/* ===== 压力控制 ===== */
.pressure-controls {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.pressure-input-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
  box-sizing: border-box;
}

.pressure-input-group label {
  font-size: 14px;
  color: #666;
  font-weight: 500;
}

.pressure-input-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.pressure-input-row .pressure-input {
  flex: 1;
}

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

/* 手动加气/泄气按钮（来自 gasgun1 分组配色） */
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

/* ===== 发射控制 ===== */
.fire-controls {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* ===== 滚动条 ===== */
.log-scroll-area::-webkit-scrollbar {
  width: 4px;
}

.log-scroll-area::-webkit-scrollbar-thumb {
  background: #2b3245;
  border-radius: 10px;
}

/* ===== 发射倒计时模态框 ===== */
.countdown-overlay {
  position: fixed;
  top: 0; left: 0; width: 100%; height: 100%;
  background: rgba(0, 0, 0, 0.7);
  display: flex; justify-content: center; align-items: center;
  z-index: 2000;
  backdrop-filter: blur(6px);
  animation: fadeIn 0.3s ease;
}

.countdown-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
}

.countdown-ring {
  width: 180px;
  height: 180px;
  border-radius: 50%;
  background: linear-gradient(135deg, #ff6b6b 0%, #ee5a5a 50%, #d63031 100%);
  box-shadow: 0 0 30px rgba(238, 90, 90, 0.6), 0 0 60px rgba(238, 90, 90, 0.4),
              inset 0 2px 10px rgba(255, 255, 255, 0.2), inset 0 -2px 10px rgba(0, 0, 0, 0.3);
  display: flex;
  justify-content: center;
  align-items: center;
  animation: countdownPulse 1s ease-in-out infinite;
}

@keyframes countdownPulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.05); }
}

.countdown-number {
  font-size: 80px;
  font-weight: 700;
  color: white;
  text-shadow: 0 2px 10px rgba(0, 0, 0, 0.3), 0 0 20px rgba(255, 255, 255, 0.5);
}

.countdown-text {
  font-size: 24px;
  font-weight: 600;
  color: white;
  letter-spacing: 4px;
}

.countdown-hint {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.8);
  animation: blink 1s ease-in-out infinite;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* ===== 禁用状态 ===== */
button:disabled {
  opacity: 0.5;
  cursor: not-allowed !important;
  pointer-events: none;
}
</style>
