<template>
  <div class="gasgun-container">
    <aside class="sidebar">
      <div class="sidebar-header">
        <div class="brand-box">
          <h2 class="title">二级轻气炮控制系统</h2>
          <p class="subtitle">Two-Stage Gas Gun Control System</p>
        </div>
      </div>

      <div class="side-card status-panel">
        <div class="status-grid">
          <div class="indicator-item">
            <div class="led led-connect" :class="{ 'active': isConnected }"></div>
            <span class="indicator-label">PLC连接</span>
          </div>
          <div class="indicator-item">
            <div class="led led-alarm" :class="{ 'active': hasAlarm }"></div>
            <span class="indicator-label">异常警报</span>
          </div>
        </div>
      </div>

      <div class="side-card config-panel">
        <input v-model="deviceIp" type="text" class="ip-input-modern" placeholder="PLC IP Address" />
        <div class="conn-btns-group">
          <button class="btn btn-connect" @click="handleConnect(true)" :disabled="isConnected">
            连接设备
          </button>
          <button class="btn btn-disconnect" @click="handleConnect(false)" :disabled="!isConnected">
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
      </div>

      <div class="sidebar-footer">
        <button class="btn-settings-modern" @click="showSettings = true">
          <span class="icon">⚙</span> 系统配置
        </button>
        <div class="logo-area">
          <span class="logo-text">LAB-PLATFORM</span>
        </div>
      </div>
    </aside>

    <main class="main-content">
      <section class="metrics-grid">
        <div class="metric-card" v-for="m in metrics" :key="m.label">
          <div class="card-inner">
            <span class="label">{{ m.label }}</span>
            <div class="value-row">
              <span class="value">{{ m.value }}</span>
              <span class="unit">{{ m.unit }}</span>
            </div>
          </div>
        </div>
      </section>

      <section class="device-visualization">
        <div class="schematic-view">
          <img src="../../assets/images/gasgun2.png" alt="Gasgun2示意图" class="cannon-image" draggable="false"/>
        </div>
      </section>

      <section class="control-panel">
        <div class="control-section">
          <h3 class="section-title">真空控制</h3>
          <div class="control-buttons">
            <button class="ctrl-btn" :class="{ 'active': vacuumRunning }" @click="toggleVacuum">
              {{ vacuumRunning ? '停止抽真空' : '开始抽真空' }}
            </button>
            <button class="ctrl-btn" :class="{ 'active': pumpTubeVacuumRunning }" @click="togglePumpTubeVacuum">
              {{ pumpTubeVacuumRunning ? '停止抽泵管' : '开始抽泵管' }}
            </button>
          </div>
        </div>

        <div class="control-section">
          <h3 class="section-title">压力控制</h3>
          <div class="pressure-controls">
            <div class="pressure-input-group">
              <label>泵管目标压力 (MPa):</label>
              <input v-model.number="pumpTubeTargetPressure" type="number" step="0.1" class="pressure-input" />
              <button class="ctrl-btn" :class="{ 'active': pumpTubePressureRunning }" @click="togglePumpTubePressure">
                {{ pumpTubePressureRunning ? '停止' : '自动控制' }}
              </button>
            </div>
            <div class="pressure-input-group">
              <label>气瓶目标压力 (MPa):</label>
              <input v-model.number="cylinderTargetPressure" type="number" step="0.1" class="pressure-input" />
              <button class="ctrl-btn" :class="{ 'active': cylinderPressureRunning }" @click="toggleCylinderPressure">
                {{ cylinderPressureRunning ? '停止' : '自动控制' }}
              </button>
            </div>
          </div>
        </div>

        <div class="control-section">
          <h3 class="section-title">发射控制</h3>
          <div class="fire-controls">
            <div class="trigger-mode">
              <label>触发模式:</label>
              <button class="mode-btn" :class="{ 'active': !isExternalTrigger }" @click="setTriggerMode(false)">内触发</button>
              <button class="mode-btn" :class="{ 'active': isExternalTrigger }" @click="setTriggerMode(true)">外触发</button>
            </div>
            <button class="ctrl-btn fire-btn" @click="prepareFire">准备发射</button>
            <button class="ctrl-btn fire-btn" @click="handleFire">立即发射</button>
          </div>
        </div>

        <div class="control-section">
          <h3 class="section-title">系统控制</h3>
          <div class="control-buttons">
            <button class="ctrl-btn reset-btn" @click="handleReset">系统恢复</button>
          </div>
        </div>
      </section>
    </main>
  </div>

  <div v-if="showSettings" class="modal-overlay">
    <div class="modal-content">
      <h3>系统配置</h3>
      <div class="settings-grid">
        <div class="setting-group">
          <label>默认IP:</label>
          <input v-model="config.ip" placeholder="192.168.2.1" />
        </div>
      </div>
      <div class="modal-actions">
        <button class="btn-settings btn-cancel" @click="showSettings = false">取消</button>
        <button class="btn-settings btn-confirm" @click="saveSettings">确认保存</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, inject } from 'vue'
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime'
import { 
  ConnectPLC, 
  DisconnectPLC,
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
  SaveConfig,
  GetConfig,
} from '../../../wailsjs/go/backend/GasGun2Controller'

const notify = inject('globalNotify')

const deviceIp = ref('192.168.6.6')
const hasAlarm = ref(false)
const isConnected = ref(false)
const showSettings = ref(false)

const vacuumRunning = ref(false)
const pumpTubeVacuumRunning = ref(false)
const pumpTubePressureRunning = ref(false)
const cylinderPressureRunning = ref(false)
const isExternalTrigger = ref(false)

const pumpTubeTargetPressure = ref(25.0)
const cylinderTargetPressure = ref(12.0)

const metrics = reactive([
  { label: '输入压力', value: '0.00', unit: 'MPa' },
  { label: '气瓶压力', value: '0.00', unit: 'MPa' },
  { label: '泵管压力', value: '0.00', unit: 'MPa' },
  { label: '泵管压力(高精)', value: '0.000', unit: 'MPa' },
  { label: '靶室真空度', value: '0.0', unit: 'Pa' },
  { label: '尾部真空度', value: '0.0', unit: 'Pa' }
])

const logs = ref([
  { time: '00:00:00', msg: '初始化系统内核...' },
  { time: '00:00:00', msg: '等待PLC连接指令' }
])

const config = reactive({
  ip: '192.168.6.6',
})

const addLog = (msg) => {
  const now = new Date();
  const timeStr = now.toTimeString().split(' ')[0];
  logs.value.push({
    time: timeStr,
    msg: msg
  });
  if (logs.value.length > 50) {
    logs.value.shift();
  }
};

const handleConnect = async (open) => {
  if (open) {
    const response = await ConnectPLC(deviceIp.value)
    if (response.Status) {
      isConnected.value = true
      addLog('PLC连接成功')
    }
    notify(response.Message, response.Status ? "success" : "error", 2000)
    addLog(response.Message)
  } else {
    const response = await DisconnectPLC()
    if (response.Status) {
      isConnected.value = false
      vacuumRunning.value = false
      pumpTubeVacuumRunning.value = false
      pumpTubePressureRunning.value = false
      cylinderPressureRunning.value = false
      addLog('PLC连接已断开')
    }
    notify(response.Message, "info", 2000)
    addLog(response.Message)
  }
}

const toggleVacuum = async () => {
  if (vacuumRunning.value) {
    const response = await StopAutoVacuum()
    if (response.Status) {
      vacuumRunning.value = false
      addLog('抽真空已停止')
    }
    notify(response.Message, response.Status ? "success" : "error", 2000)
  } else {
    const response = await StartAutoVacuum()
    if (response.Status) {
      vacuumRunning.value = true
      addLog('抽真空已启动')
    }
    notify(response.Message, response.Status ? "success" : "error", 2000)
  }
}

const togglePumpTubeVacuum = async () => {
  if (pumpTubeVacuumRunning.value) {
    const response = await StopPumpTubeVacuum()
    if (response.Status) {
      pumpTubeVacuumRunning.value = false
      addLog('抽泵管已停止')
    }
    notify(response.Message, response.Status ? "success" : "error", 2000)
  } else {
    const response = await StartPumpTubeVacuum()
    if (response.Status) {
      pumpTubeVacuumRunning.value = true
      addLog('抽泵管已启动')
    }
    notify(response.Message, response.Status ? "success" : "error", 2000)
  }
}

const togglePumpTubePressure = async () => {
  if (pumpTubePressureRunning.value) {
    const response = await StopAutoPumpTubePressure()
    if (response.Status) {
      pumpTubePressureRunning.value = false
      addLog('泵管自动压力控制已停止')
    }
    notify(response.Message, response.Status ? "success" : "error", 2000)
  } else {
    const response = await AutoPumpTubePressure(pumpTubeTargetPressure.value)
    if (response.Status) {
      pumpTubePressureRunning.value = true
      addLog(`泵管自动压力控制已启动，目标: ${pumpTubeTargetPressure.value} MPa`)
    }
    notify(response.Message, response.Status ? "success" : "error", 2000)
  }
}

const toggleCylinderPressure = async () => {
  if (cylinderPressureRunning.value) {
    const response = await StopAutoCylinderPressure()
    if (response.Status) {
      cylinderPressureRunning.value = false
      addLog('气瓶自动压力控制已停止')
    }
    notify(response.Message, response.Status ? "success" : "error", 2000)
  } else {
    const response = await AutoCylinderPressure(cylinderTargetPressure.value)
    if (response.Status) {
      cylinderPressureRunning.value = true
      addLog(`气瓶自动压力控制已启动，目标: ${cylinderTargetPressure.value} MPa`)
    }
    notify(response.Message, response.Status ? "success" : "error", 2000)
  }
}

const setTriggerMode = async (isExternal) => {
  isExternalTrigger.value = isExternal
  SetTriggerMode(isExternal)
  addLog(`触发模式已切换为: ${isExternal ? '外触发' : '内触发'}`)
  notify(`触发模式已切换为: ${isExternal ? '外触发' : '内触发'}`, "success", 2000)
}

const prepareFire = async () => {
  const response = await PrepareFire()
  if (response.Status) {
    addLog('准备发射完成')
  }
  notify(response.Message, response.Status ? "success" : "error", 2000)
}

const handleFire = async () => {
  const response = await Fire()
  if (response.Status) {
    addLog('发射指令已执行')
  }
  notify(response.Message, response.Status ? "success" : "error", 2000)
}

const handleReset = async () => {
  const response = await ResetSystem()
  if (response.Status) {
    addLog('系统恢复完成')
  }
  notify(response.Message, response.Status ? "success" : "error", 2000)
}

const saveSettings = async () => {
  showSettings.value = false
  const response = await SaveConfig(config)
  if (!response.Status) {
    notify(response.Message, "error", 2000)
  } else {
    notify(response.Message, "success", 2000)
    addLog('配置已保存')
  }
}

onMounted(async () => {
  const result = await GetConfig()
  if (result) {
    config.ip = result.ip
    deviceIp.value = result.ip
  }

  EventsOn("update_gasgun2_metrics", (data) => {
    if (!data) return
    metrics[0].value = data.inputPressure.toFixed(2)
    metrics[1].value = data.cylinderPressure.toFixed(2)
    metrics[2].value = data.pumpTubePressure.toFixed(2)
    metrics[3].value = data.pumpTubePressureHi.toFixed(3)
    metrics[4].value = data.targetVacuumDegree.toFixed(1)
    metrics[5].value = data.tailVacuumDegree.toFixed(1)
  })
})

onUnmounted(() => {
  EventsOff("update_gasgun2_metrics")
})
</script>

<style scoped>
.gasgun-container {
  display: flex;
  height: 100vh;
  background: #f4f7f9;
  overflow: hidden;
}

.sidebar {
  width: 280px;
  background: #1e222d;
  display: flex;
  flex-direction: column;
  padding: 24px 16px;
  box-shadow: 4px 0 15px rgba(0,0,0,0.2);
  z-index: 10;
}

.sidebar-header {
  margin-bottom: 24px;
  padding: 0 8px;
}

.title { 
  font-size: 20px; 
  color: #fff; 
  font-weight: 600; 
  letter-spacing: 0.5px;
}

.subtitle { 
  font-size: 10px; 
  color: #4facfe; 
  text-transform: uppercase; 
  margin-top: 4px;
}

.side-card {
  background: rgba(255,255,255,0.05);
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 16px;
  border: 1px solid rgba(255,255,255,0.1);
}

.panel-label {
  font-size: 12px;
  color: #868e96;
  margin-bottom: 12px;
  font-weight: bold;
  text-transform: uppercase;
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
  background: #343a40;
  box-shadow: inset 0 1px 3px rgba(0,0,0,0.5);
  transition: 0.3s;
}

.led-connect.active { background: #41b883; box-shadow: 0 0 10px #41b883; }
.led-alarm.active { background: #fa5252; box-shadow: 0 0 10px #fa5252; animation: blink 1s infinite; }

.indicator-label { font-size: 11px; color: #ced4da; }

.ip-input-modern {
  width: 80%;
  background: #2a2d3a;
  border: 1px solid #444;
  padding: 10px;
  border-radius: 4px;
  color: #fff;
  margin: 10px;
  font-size: 1.2rem;
}

.conn-btns-group {
  display: flex;
  gap: 8px;
}

.btn { flex: 1; padding: 8px; border: none; border-radius: 4px; cursor: pointer; }
.btn-connect { background: #4facfe; color: #fff; }
.btn-disconnect { background: #333; color: #ccc; }

.info-log-box {
  background: #14171e;
  border-radius: 6px;
  height: 180px;
  padding: 8px;
  overflow: hidden;
}

.log-scroll-area {
  height: 100%;
  overflow-y: auto;
  font-size: 12px;
}

.log-line { margin: 4px 0; color: #00ff00; line-height: 1.4; }
.log-time { color: #666; margin-right: 6px; }

.sidebar-footer {
  margin-top: auto;
}

.btn-settings-modern {
  width: 100%;
  padding: 12px;
  background: #343a40;
  border: none;
  border-radius: 8px;
  color: #adb5bd;
  cursor: pointer;
  font-size: 13px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-settings-modern:hover { color: #fff; background: #495057; }

.logo-area {
  text-align: center;
  margin-top: 16px;
  border-top: 1px solid #333;
  padding-top: 12px;
}

.logo-text { font-weight: 800; color: #444; letter-spacing: 2px; font-size: 14px; }

.main-content {
  flex: 1;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  overflow-y: auto;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  flex-shrink: 0;
}

.metric-card {
  background: #fff;
  padding: 20px;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.05);
  border-top: 4px solid #4facfe;
}

.metric-card .label { font-size: 14px; color: #666; }

.metric-card .value {
  font-size: 28px;
  font-weight: bold;
  color: #1a1c24;
}

.metric-card .unit { margin-left: 5px; color: #888; }

.device-visualization {
  flex: 1;
  background: #fff;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  min-height: 300px;
}

.schematic-view {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

.cannon-image { 
  width: 100%; 
  height: 100%;
  object-fit: contain;
}

.control-panel {
  background: #fff;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.05);
  flex-shrink: 0;
}

.control-section {
  margin-bottom: 24px;
  padding-bottom: 24px;
  border-bottom: 1px solid #e9ecef;
}

.control-section:last-child {
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: none;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #1a1c24;
  margin-bottom: 16px;
}

.control-buttons {
  display: flex;
  gap: 12px;
}

.ctrl-btn {
  flex: 1;
  padding: 12px 20px;
  border-radius: 8px;
  border: 1px solid #dee2e6;
  background: #f8f9fa;
  font-weight: 600;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.ctrl-btn:hover {
  background: #e9ecef;
  border-color: #adb5bd;
}

.ctrl-btn.active {
  background: #4facfe;
  color: white;
  border-color: #4facfe;
}

.ctrl-btn.fire-btn {
  background: #fa5252;
  color: white;
  border: none;
}

.ctrl-btn.fire-btn:hover {
  background: #e03131;
}

.ctrl-btn.reset-btn {
  background: #495057;
  color: white;
  border: none;
}

.pressure-controls {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.pressure-input-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.pressure-input-group label {
  font-size: 13px;
  color: #666;
  font-weight: 500;
}

.pressure-input {
  padding: 10px;
  border: 1px solid #dee2e6;
  border-radius: 6px;
  font-size: 14px;
}

.fire-controls {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.trigger-mode {
  display: flex;
  align-items: center;
  gap: 12px;
}

.trigger-mode label {
  font-size: 14px;
  color: #666;
  font-weight: 500;
}

.mode-btn {
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid #dee2e6;
  background: #f8f9fa;
  font-weight: 500;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.mode-btn.active {
  background: #4facfe;
  color: white;
  border-color: #4facfe;
}

@keyframes blink { 
  0%, 100% { opacity: 1; } 
  50% { opacity: 0.3; } 
}

.log-scroll-area::-webkit-scrollbar { width: 4px; }
.log-scroll-area::-webkit-scrollbar-thumb { background: #333; border-radius: 10px; }

.btn-settings {
  width: 100%;
  margin-top: 15px;
  padding: 10px;
  background: #2f3542;
  border: 1px solid #57606f;
  color: #ced4da;
  border-radius: 4px;
  cursor: pointer;
  transition: 0.3s;
}

.btn-settings:hover { background: #57606f; color: #fff; }

.modal-overlay {
  position: fixed;
  top: 0; left: 0; width: 100%; height: 100%;
  background: rgba(0,0,0,0.7);
  display: flex; justify-content: center; align-items: center;
  z-index: 1000;
}

.modal-content {
  background: #f8f9fa; padding: 25px; border-radius: 12px;
  width: 500px; color: #333;
}

.settings-grid {
  display: grid; grid-template-columns: 1fr; gap: 15px;
}

.setting-group { display: flex; flex-direction: column; gap: 5px; }
.setting-group label { font-size: 12px; color: #666; font-weight: bold; }
.setting-group input { padding: 8px; border: 1px solid #ccc; border-radius: 4px; }

.modal-actions {
  margin-top: 20px; display: flex; gap: 10px; justify-content: flex-end;
}

.btn-confirm { background: #4facfe; color: white; padding: 10px 25px; }
.btn-cancel { background: #ced4da; color: #333; padding: 10px 25px; }
</style>