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
        <!-- 输入气压 -->
        <div class="metric-card">
          <span class="label">输入气压 (MPa)</span>
          <span class="main-value">{{ inputPressure }}</span>
        </div>
        
        <!-- 尾部真空度 -->
        <div class="metric-card">
          <span class="label">尾部真空度 (Pa)</span>
          <span class="main-value">{{ tailVacuumDegree }}</span>
        </div>
        
        <!-- 一级气室气压 -->
        <div class="metric-card">
          <span class="label">一级气室气压 (MPa)</span>
          <span class="main-value">{{ cylinderPressure }}</span>
        </div>
        
        <!-- 二级泵管气压 -->
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
        
        <!-- 靶室真空度 -->
        <div class="metric-card">
          <span class="label">靶室真空度 (Pa)</span>
          <span class="main-value">{{ targetVacuumDegree }}</span>
        </div>
      </section>

      <section class="device-controls"> 
        <section class="device-visualization">
          <div class="schematic-view">
            <img src="../../assets/images/gasgun2.png" alt="Gasgun2示意图" class="cannon-image" draggable="false"/>
          </div>
          <div v-if=false class="footer-leds">
            <div class="led-row">
              <div v-for="i in 16" :key="'row1-' + i" class="footer-led" :class="{ 'active': ledStatus[0][i-1] }"></div>
            </div>
            <div class="led-row">
              <div v-for="i in 16" :key="'row2-' + i" class="footer-led" :class="{ 'active': ledStatus[1][i-1] }"></div>
            </div>
          </div>
        </section>

        <section class="control-panel">
          <div class="control-section">
            <h3 class="section-title">① 真空控制</h3>
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
            <h3 class="section-title">② 压力控制</h3>
            <div class="pressure-controls">
              <div class="pressure-input-group">
                <label>泵管压力 (MPa):</label>
                <input v-model.number="pumpTubeTargetPressure" type="number" step="0.1" class="pressure-input" />
                <button class="ctrl-btn" :class="{ 'active': pumpTubePressureRunning }" @click="togglePumpTubePressure">
                  {{ pumpTubePressureRunning ? '停止' : '自动控制' }}
                </button>
              </div>
              <div class="pressure-input-group">
                <label>气瓶压力 (MPa):</label>
                <input v-model.number="cylinderTargetPressure" type="number" step="0.1" class="pressure-input" />
                <button class="ctrl-btn" :class="{ 'active': cylinderPressureRunning }" @click="toggleCylinderPressure">
                  {{ cylinderPressureRunning ? '停止' : '自动控制' }}
                </button>
              </div>
            </div>
          </div>

          <div class="control-section">
            <h3 class="section-title">③ 发射控制</h3>
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
            <h3 class="section-title">④ 系统控制</h3>
            <div class="control-buttons">
              <button class="ctrl-btn reset-btn" @click="handleReset">系统恢复</button>
            </div>
          </div>
        </section>
      </section>

      

      
    </main>
  </div>

  <div v-if="showSettings" class="modal-overlay">
    <div class="modal-content config-modal">
      <div class="modal-header">
        <h3>系统配置</h3>
      </div>
      <div class="settings-scroll">
        <div class="settings-section">
          <h4 class="settings-section-title">网络配置</h4>
          <div class="setting-group">
            <label>默认IP:</label>
            <input v-model="config.ip" placeholder="192.168.6.6" />
          </div>
        </div>
        
        <div class="settings-section">
          <h4 class="settings-section-title">阀门点位配置</h4>
          <div class="settings-grid-2col">
            <div class="setting-group">
              <label>增压阀:</label>
              <input v-model.number="config.switches.Pressurize" type="number" />
            </div>
            <div class="setting-group">
              <label>减压阀:</label>
              <input v-model.number="config.switches.Decompress" type="number" />
            </div>
            <div class="setting-group">
              <label>泵管增压阀:</label>
              <input v-model.number="config.switches.PumpTubePressurize" type="number" />
            </div>
            <div class="setting-group">
              <label>泵管减压阀:</label>
              <input v-model.number="config.switches.PumpTubeDecompress" type="number" />
            </div>
            <div class="setting-group">
              <label>抽泵管真空阀:</label>
              <input v-model.number="config.switches.PumpTubeVacuum" type="number" />
            </div>
            <div class="setting-group">
              <label>靶室真空阀:</label>
              <input v-model.number="config.switches.TargetVacuum" type="number" />
            </div>
            <div class="setting-group">
              <label>尾真空保护阀:</label>
              <input v-model.number="config.switches.TailVacuumProtect" type="number" />
            </div>
            <div class="setting-group">
              <label>泵管保护阀:</label>
              <input v-model.number="config.switches.PumpTubeProtect" type="number" />
            </div>
            <div class="setting-group">
              <label>发射阀:</label>
              <input v-model.number="config.switches.FireSwitch" type="number" />
            </div>
            <div class="setting-group">
              <label>系统减压阀:</label>
              <input v-model.number="config.switches.SystemDecompress" type="number" />
            </div>
          </div>
        </div>
        
        <div class="settings-section">
          <h4 class="settings-section-title">真空泵点位配置</h4>
          <div class="settings-grid-2col">
            <div class="setting-group">
              <label>靶室真空泵:</label>
              <input v-model.number="config.switches.TargetVacuumPump" type="number" />
            </div>
            <div class="setting-group">
              <label>尾真空泵:</label>
              <input v-model.number="config.switches.TailVacuumPump" type="number" />
            </div>
          </div>
        </div>
        
        <div class="settings-section">
          <h4 class="settings-section-title">数据地址配置</h4>
          <div class="settings-grid-2col">
            <div class="setting-group">
              <label>输入压力:</label>
              <input v-model.number="config.dataAddresses.InputPressure" type="number" />
            </div>
            <div class="setting-group">
              <label>气瓶压力:</label>
              <input v-model.number="config.dataAddresses.CylinderPressure" type="number" />
            </div>
            <div class="setting-group">
              <label>泵管压力:</label>
              <input v-model.number="config.dataAddresses.PumpTubePressure" type="number" />
            </div>
            <div class="setting-group">
              <label>泵管压力(高精度):</label>
              <input v-model.number="config.dataAddresses.PumpTubePressureHi" type="number" />
            </div>
            <div class="setting-group">
              <label>靶室真空度:</label>
              <input v-model.number="config.dataAddresses.TargetVacuumDegree" type="number" />
            </div>
            <div class="setting-group">
              <label>尾部真空度:</label>
              <input v-model.number="config.dataAddresses.TailVacuumDegree" type="number" />
            </div>
          </div>
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

const pumpTubeTargetPressure = ref(2.0)
const cylinderTargetPressure = ref(1.0)

// 指标数据
const inputPressure = ref('00.00')
const tailVacuumDegree = ref('10000.0')
const cylinderPressure = ref('00.00')
const pumpTubePressure = ref('00.00')
const pumpTubePressureHi = ref('00.000')
const targetVacuumDegree = ref('10000.0')

// LED状态 (两行，每行16个)
const ledStatus = reactive([
  [false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false],
  [false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false]
])

// 二级泵管气压精度切换
const pumpTubeIsHighPrecision = ref(false)
const pumpTubeLabel = ref('二级泵管气压 (MPa)')
const pumpTubeMainValue = ref('0.00')
const pumpTubeSubValue = ref('0.000')

const PumpTubePrecision = (key) => {
  if (key === 'high') {
    pumpTubeIsHighPrecision.value = true
  } else if (key === 'low') {
    pumpTubeIsHighPrecision.value = false
  }
  if (pumpTubeIsHighPrecision.value) {
    pumpTubeLabel.value = '二级泵管气压 (高精度)'
    // 交换数值：主值显示高精度，副值显示普通精度
    pumpTubeMainValue.value = pumpTubePressureHi.value
    pumpTubeSubValue.value = pumpTubePressure.value
  } else {
    pumpTubeLabel.value = '二级泵管气压 (MPa)'
    // 恢复：主值显示普通精度，副值显示高精度
    pumpTubeMainValue.value = pumpTubePressure.value
    pumpTubeSubValue.value = pumpTubePressureHi.value
  }
}

const logs = ref([
  { time: '00:00:00', msg: '初始化系统内核...' },
  { time: '00:00:00', msg: '等待PLC连接指令' }
])

const config = reactive({
  ip: '192.168.6.6',
  switches: {
    Pressurize: 1,
    Decompress: 2,
    PumpTubePressurize: 3,
    PumpTubeDecompress: 4,
    PumpTubeVacuum: 5,
    TargetVacuum: 6,
    TailVacuumProtect: 7,
    PumpTubeProtect: 8,
    FireSwitch: 9,
    SystemDecompress: 10,
    TargetVacuumPump: 11,
    TailVacuumPump: 12,
  },
  dataAddresses: {
    InputPressure: 0,
    CylinderPressure: 2,
    PumpTubePressure: 4,
    PumpTubePressureHi: 6,
    TargetVacuumDegree: 8,
    TailVacuumDegree: 10,
  }
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
    if (result.switches) {
      config.switches = { ...result.switches }
    }
    if (result.dataAddresses) {
      config.dataAddresses = { ...result.dataAddresses }
    }
  }

  EventsOn("update_gasgun2_metrics", (data) => {
    if (!data) return
    inputPressure.value = data.inputPressure.toFixed(2)
    tailVacuumDegree.value = data.tailVacuumDegree.toFixed(1)
    cylinderPressure.value = data.cylinderPressure.toFixed(2)
    pumpTubePressure.value = data.pumpTubePressure.toFixed(2)
    pumpTubePressureHi.value = data.pumpTubePressureHi.toFixed(3)
    targetVacuumDegree.value = data.targetVacuumDegree.toFixed(1)
    
    // 更新泵管气压显示值
    if (pumpTubeIsHighPrecision.value) {
      pumpTubeMainValue.value = pumpTubePressureHi.value
      pumpTubeSubValue.value = pumpTubePressure.value
    } else {
      pumpTubeMainValue.value = pumpTubePressure.value
      pumpTubeSubValue.value = pumpTubePressureHi.value
    }
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

.log-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  margin-bottom: 16px;
}

.log-panel .panel-label {
  flex-shrink: 0;
}

.log-panel .info-log-box {
  flex: 1;
  min-height: 0;
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
  grid-template-columns: repeat(5, 1fr);
  gap: 16px;
  flex-shrink: 0;
}

.metric-card {
  background: #fff;
  padding: 16px;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.05);
  border-top: 4px solid #4facfe;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.metric-card .label { 
  font-size: 13px; 
  color: #666; 
  text-align: center;
}

.metric-card .main-value {
  font-size: 28px;
  font-weight: bold;
  color: #1a1c24;
}

.metric-card .sub-value {
  display: none;
}

/* 二级泵管气压卡片特殊样式 */
.pump-tube-card {
  border-top-color: #f97316;
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
  border: 1px solid #4facfe;
  border-radius: 4px;
  background: #fff;
  color: #4facfe;
  cursor: pointer;
  transition: all 0.2s;
}

.pump-tube-card .toggle-btn:hover {
  background: #4facfe;
  color: #fff;
}

.device-controls {
  display: flex;
  flex-direction: row;
  gap: 20px;
  flex: 1;
  min-height: 0;
}


.device-visualization {
  flex: 1;
  background: #fff;
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px;
  min-height: 300px;
}

.schematic-view {
  flex: 1;
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

.cannon-image { 
  width: 100%; 
  height: 100%;
  object-fit: contain;
}

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
  background: #4facfe;
  border-color: #4facfe;
  box-shadow: 0 0 8px rgba(79, 172, 254, 0.6);
}

.control-panel {
  background: #fff;
  border-radius: 16px;
  padding: 16px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.05);
  flex-shrink: 0;
  width: 320px;
  max-width: 40%;
  box-sizing: border-box;
  overflow-y: auto;
  max-height: 100%;
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
  gap: 8px;
  flex-wrap: wrap;
}

.ctrl-btn {
  flex: 1;
  min-width: calc(50% - 4px);
  padding: 10px 12px;
  border-radius: 6px;
  border: 1px solid #dee2e6;
  background: #f8f9fa;
  font-weight: 500;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  box-sizing: border-box;
  word-break: keep-all;
  text-align: center;
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
  gap: 8px;
}

.pressure-input-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
  box-sizing: border-box;
}

.pressure-input-group label {
  font-size: 12px;
  color: #666;
  font-weight: 500;
}

.pressure-input {
  padding: 8px;
  border: 1px solid #dee2e6;
  border-radius: 6px;
  font-size: 13px;
  width: 100%;
  box-sizing: border-box;
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
  background: rgba(0, 0, 0, 0.6);
  display: flex; justify-content: center; align-items: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal-content {
  background: linear-gradient(145deg, #ffffff, #f1f3f4);
  padding: 0;
  border-radius: 16px;
  width: 500px;
  color: #333;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3),
              0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from { 
    opacity: 0;
    transform: translateY(20px);
  }
  to { 
    opacity: 1;
    transform: translateY(0);
  }
}

.config-modal {
  width: 650px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  padding: 20px 24px;
  color: white;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 10px;
}

.modal-header h3::before {
  content: '⚙';
  font-size: 20px;
}

.settings-scroll {
  overflow-y: auto;
  max-height: 60vh;
  padding: 24px;
}

.settings-scroll::-webkit-scrollbar {
  width: 6px;
}

.settings-scroll::-webkit-scrollbar-track {
  background: #f1f3f4;
  border-radius: 3px;
}

.settings-scroll::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.settings-scroll::-webkit-scrollbar-thumb:hover {
  background: #a1a1a1;
}

.settings-section {
  margin-bottom: 28px;
  padding: 16px;
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  border: 1px solid #e8eaed;
}

.settings-section:last-child {
  margin-bottom: 0;
}

.settings-section-title {
  font-size: 14px;
  font-weight: 600;
  color: #1a1c24;
  margin-bottom: 16px;
  padding-left: 12px;
  border-left: 4px solid #4facfe;
  position: relative;
}

.settings-section-title::after {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 4px;
  height: 16px;
  background: linear-gradient(180deg, #4facfe 0%, #00f2fe 100%);
  border-radius: 0 2px 2px 0;
}

.settings-grid {
  display: grid; grid-template-columns: 1fr; gap: 16px;
}

.settings-grid-2col {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14px;
}

.setting-group { 
  display: flex; 
  flex-direction: column; 
  gap: 8px; 
}

.setting-group label { 
  font-size: 13px; 
  color: #5f6368; 
  font-weight: 500; 
  padding-left: 8px;
}

.setting-group input { 
  padding: 10px 12px; 
  border: 2px solid #e8eaed; 
  border-radius: 8px;
  font-size: 14px;
  color: #3c4043;
  background: #fafafa;
  transition: all 0.2s ease;
  outline: none;
}

.setting-group input:focus { 
  border-color: #4facfe; 
  background: #ffffff;
  box-shadow: 0 0 0 3px rgba(79, 172, 254, 0.1);
}

.setting-group input:hover {
  border-color: #dadce0;
}

.setting-group input[type="number"] {
  -moz-appearance: textfield;
}

.setting-group input[type="number"]::-webkit-outer-spin-button,
.setting-group input[type="number"]::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.modal-actions {
  padding: 16px 24px;
  background: #f8f9fa;
  border-top: 1px solid #e8eaed;
  display: flex; gap: 12px; justify-content: flex-end;
}

.btn-settings {
  padding: 10px 24px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
  outline: none;
}

.btn-confirm { 
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  color: white; 
  box-shadow: 0 2px 8px rgba(79, 172, 254, 0.3);
}

.btn-confirm:hover { 
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(79, 172, 254, 0.4);
}

.btn-confirm:active {
  transform: translateY(0);
}

.btn-cancel { 
  background: #e8eaed; 
  color: #5f6368; 
}

.btn-cancel:hover { 
  background: #dadce0;
  color: #3c4043;
}
</style>