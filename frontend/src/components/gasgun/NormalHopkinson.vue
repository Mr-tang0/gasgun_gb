<template>
  <div class="gasgun-container">
    <aside class="sidebar">
      <div class="sidebar-header">
        <div class="brand-box">
          <h2 class="title">常温Hopkinson</h2>
          <p class="subtitle">Hopkinson SYSTEM</p>
        </div>
      </div>

      <div class="side-card status-panel">
        <div class="panel-label">设备状态</div>
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
        <div class="panel-label">网络通信</div>
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
              <span class="value" :class="{ 'alarm': parseFloat(m.value) >= 0.3 }">
                {{ m.value }}
              </span>
              <span class="unit">{{ m.unit }}</span>
            </div>
          </div>
        </div>
      </section>

      <section class="device-visualization">
        <div class="schematic-view">
          <img src="../../assets/images/normalHopkinson.png" alt="Hopkinson示意图" class="cannon-image" draggable="false"/>
        </div>
      </section>

      <footer class="action-bar">
        <button class="act-btn" :class="{ 'active-pressing': inletSwitchFlag }"
            @mousedown="handleInletSwitch(true)" 
            @mouseup="handleInletSwitch(false)"
            @mouseleave="handleInletSwitch(false)"> {{ inletSwitchFlag ? '进气中...' : '手动进气' }}
        </button>

        <button class="act-btn" :class="{ 'active-pressing': exhaustSwitchFlag }"
          @mousedown="handleOutletSwitch(true)" 
          @mouseup="handleOutletSwitch(false)" 
          @mouseleave="handleOutletSwitch(false)"> {{ exhaustSwitchFlag ? '排气中...' : '手动排气' }}
        </button>
        <button class="act-btn fire" @click="handleFire">系统发射</button>
        <button class="act-btn reset" @click="handleReset">复位/归零</button>
      </footer>
    </main>

    <div v-if="showSettings" class="modal-overlay">

      <div class="modal-content">
        <div class="modal-header">
          <h3>工程地址与量程配置</h3>
        </div>

        <div class="settings-grid">

          <div class="setting-group">
            <label>进气地址 (Q):</label>
            <input v-model="config.inletAddr" placeholder="e.g. Q0.0" />
          </div>

          <div class="setting-group">
            <label>出气地址 (Q):</label>
            <input v-model="config.outletAddr" placeholder="e.g. Q0.1" />
          </div>

          <div class="setting-group">
            <label>发射地址 (Q):</label>
            <input v-model="config.fireAddr" placeholder="e.g. Q0.2" />
          </div>

          <hr class="full-width" />

          <div class="setting-group">
            <label>气瓶压力地址 (VD):</label>
            <input v-model="config.tankPressureAddr" placeholder="e.g. VD0" />
          </div>

          <div class="setting-group">
            <label>供气压力地址 (VD):</label>
            <input v-model="config.supplyPressureAddr" placeholder="e.g. VD4" />
          </div>

          <div class="setting-group">
            <label>气瓶量程 (MPa):</label>
            <input v-model.number="config.tankRange" type="number" />
          </div>

          <div class="setting-group">
            <label>供气量程 (MPa):</label>
            <input v-model.number="config.supplyRange" type="number" />
          </div>

          <hr class="full-width" />

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
  </div>
</template>

<script setup>
/* ... script 逻辑保持不变，确保导入和变量已定义 ... */
import { ref, reactive, onMounted, onUnmounted, inject} from 'vue'
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime'
import { 
    ConnectPLC, 
    DisconnectPLC, 
    InletSwitch, 
    ExhaustSwitch, 
    FireSwitch,
    SaveConfig,
    GetConfig,
} from '../../../wailsjs/go/backend/NormalHopkinsonContoller'

const metrics = reactive([
  { label: '供气压力', value: '0.00', unit: 'MPa' },
  { label: '气瓶压力', value: '0.00', unit: 'MPa' },
])

const logs = ref([
  { time: '14:20:01', msg: '初始化系统内核...' },
  { time: '14:20:05', msg: '等待PLC连接指令' }
])

const config = reactive({
    inletAddr: 'Q0.0', outletAddr: 'Q0.1', fireAddr: 'Q0.2',
    tankPressureAddr: 'VD0', supplyPressureAddr: 'VD4',
    tankRange: 16.0, supplyRange: 16.0,
    ip: '192.168.2.1',
})


const notify = inject('globalNotify')
const deviceIp = ref('192.168.2.1')
const hasAlarm = ref(false)
const isConnected = ref(false)
const inletSwitchFlag = ref(false)
const exhaustSwitchFlag = ref(false)
const showSettings = ref(false)


const addLog = (msg) => {
  const now = new Date();
  const timeStr = now.toTimeString().split(' ')[0]; // 获取 HH:mm:ss 格式
  logs.value.push({
    time: timeStr,
    msg: msg
  });
  if (logs.value.length > 50) {
    logs.value.shift();
  }

  scrollToBottom();
};


const handleConnect = async (open) => {
  if (open){
    const response = await ConnectPLC(deviceIp.value)
    if(response.Status) isConnected.value = true

    notify(response.Message, response.Status? "success": "error", 2000)
    addLog(response.Message)
  } else {

    const response = await DisconnectPLC()
    if (response.Status) isConnected.value = false

    notify(response.Message, "info", 2000)
    addLog(response.Message)
  }
}

const handleInletSwitch = async (open) => {
  if (inletSwitchFlag.value === open) return;
  const response = await InletSwitch(open);
  if (response.Status) {
    inletSwitchFlag.value = open;
    
  }
  notify(response.Message, response.Status? "success": "error", 2000)
  addLog(response.Message)
};

const handleOutletSwitch = async (open) => {
  if (exhaustSwitchFlag.value === open) return;
  const response = await ExhaustSwitch(open);
  if (response.Status) exhaustSwitchFlag.value = open;

  notify(response.Message, response.Status? "success": "error", 2000)
  addLog(response.Message)
};

const handleFire = async () => {
  const response = await FireSwitch(500)
  notify(response.Message, response.Status? "success": "error", 2000)
  addLog(response.Message)
}

const handleReset = () => { /* 逻辑 */ }
const saveSettings = async () => { 
  showSettings.value = false; 
  const response = await SaveConfig(config)
  if (!response.Status) 
  {
    notify(response.Message, "error", 2000);
  }else
  {
    notify(response.Message, "success", 2000); 
  }
  
}

onMounted(async() => {
    // 初始化配置
    const result = await GetConfig()
    if (result) {
          config.inletAddr = result.inletAddr;
          config.outletAddr = result.outletAddr;
          config.fireAddr = result.fireAddr;
          config.tankPressureAddr = result.tankPressureAddr;
          config.supplyPressureAddr = result.supplyPressureAddr;
          config.tankRange = result.tankRange;
          config.supplyRange = result.supplyRange;
          config.ip = result.ip;
          deviceIp.value = result.ip;
      }

    EventsOn("normal_hopkinson_metrics", (data) => {
        if (!data) return;
        metrics.forEach(item => {
            if (data[item.label] !== undefined) item.value = data[item.label].toFixed(2);
        });
    })
})
onUnmounted(() => { EventsOff("normal_hopkinson_metrics") })
</script>

<style scoped>
/* 基础容器 */
.gasgun-container {
  display: flex;
  height: 100vh;
  background: #f4f7f9;
  overflow: hidden;
}

/* 边栏现代化样式 */
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

/* 侧边功能卡片 */
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

/* 状态灯网格 */
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

/* 输入框与按钮 */

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

.conn-btns { display: flex; gap: 10px; }
.btn { flex: 1; padding: 8px; border: none; border-radius: 4px; cursor: pointer; }
.btn-connect { background: #4facfe; color: #fff; }
.btn-disconnect { background: #333; color: #ccc; }


/* 日志区 */
.info-log-box {
  background: #14171e;
  border-radius: 6px;
  padding: 8px;
  overflow: hidden;
}

.log-scroll-area {
  height: 100%;
  overflow-y: auto;
  /* font-family: 'Consolas', monospace; */
  font-size: 12px;
}

.log-line { margin: 4px 0; color: #00ff00;; line-height: 1.4; }
.log-time { color: #666; margin-right: 6px; }

/* 侧边栏底部 */
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

/* 主区域网格 */
.main-content {
  flex: 1;
  padding: 24px;
  display: flex;
  flex-direction: column;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.metric-card {
      background: #fff;
      padding: 20px;
      border-radius: 12px;
      box-shadow: 0 4px 6px rgba(0,0,0,0.05);
      border-top: 4px solid #4facfe;
  }

.metric-card .label { font-size: 14px; color: #666; }



@keyframes alarm-blink {
    0% { background-color: #fff9f9; border-top-color: #ff3b30; }
    50% { background-color: #ffe5e5; border-top-color: #ff0000; box-shadow: 0 4px 15px rgba(255, 59, 48, 0.3); }
    100% { background-color: #fff9f9; border-top-color: #ff3b30; }
}

.alarm {
    color: #ff3b30 !important;
}

.metric-card:has(.alarm) {
    animation: alarm-blink 1s infinite ease-in-out;
    border-top-width: 4px;
    border-top-style: solid;
}

.metric-card .value {
    font-size: 28px;
    font-weight: bold;
    color: #1a1c24;
    transition: color 0.3s ease;
}

.metric-card .unit { margin-left: 5px; color: #888; }

/* 发射按钮条 */
.action-bar {
  background: #fff;
  padding: 16px;
  border-radius: 16px;
  display: flex;
  gap: 16px;
  box-shadow: 0 -5px 20px rgba(0,0,0,0.03);
}

.act-btn {
  flex: 1;
  height: 60px;
  border-radius: 12px;
  border: 1px solid #dee2e6;
  background: #f8f9fa;
  font-weight: 700;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.2s;
}

.act-btn:active { transform: scale(0.98); }
.act-btn.active-pressing { background: #e7f5ff; border-color: #4facfe; color: #4facfe; box-shadow: inset 0 2px 4px rgba(0,0,0,0.1); }

.act-btn.fire { background: #fa5252; color: white; border: none; }
.act-btn.fire:hover { background: #e03131; }

.act-btn.reset { background: #495057; color: white; border: none; }

/* 动画定义 */
@keyframes blink { 0%, 100% { opacity: 1; } 50% { opacity: 0.3; } }
@keyframes text-flicker { 0%, 100% { opacity: 1; } 50% { opacity: 0.7; } }

/* 滚动条美化 */
.log-scroll-area::-webkit-scrollbar { width: 4px; }
.log-scroll-area::-webkit-scrollbar-thumb { background: #333; border-radius: 10px; }

/* 响应式示意图 */
.device-visualization {
  flex: 1;
  background: #fff;
  border-radius: 16px;
  margin-bottom: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
}
.cannon-image { max-width: 100%; max-height: 100%; object-fit: contain; }


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

.settings-grid {
  padding: 24px;
  display: grid; grid-template-columns: 1fr 1fr; gap: 16px;
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

.full-width { 
  grid-column: span 2; 
  width: 100%; 
  border: 0; 
  border-top: 1px solid #e8eaed; 
  margin: 8px 0;
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