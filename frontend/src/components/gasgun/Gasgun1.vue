<template>
  <div class="gasgun-container">
    <aside class="sidebar">
      <div class="sidebar-header">
        <h2 class="title">一级气炮系统</h2>
        <p class="subtitle">Single-Stage Gas Gun Control System</p>
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
          <span class="label">{{ m.label }}</span>
          <div class="value-row">
            <span class="value">{{ m.value }}</span>
            <span class="unit">{{ m.unit }}</span>
          </div>
        </div>
      </section>

      <section class="device-visualization">
        <div class="schematic-view">
            <div class="cannon-svg">
                <img src="../../assets/images/gasgun1.png" alt="气炮示意图" class="cannon-image" draggable="false"/>
            </div>
        </div>
      </section>

      <footer class="action-bar">
        <button class="act-btn" :class="{ 'active-pressing': inletSwitchFlag }"
            @mousedown="handleInletSwitch(true)" 
            @mouseup="handleInletSwitch(false)"
            @mouseleave="handleInletSwitch(false)"> {{ inletSwitchFlag ? '进气中...' : '进气' }}
        </button>
        <button class="act-btn" :class="{ 'active-pressing': exhaustSwitchFlag }"
            @mousedown="handleOutletSwitch(true)" 
            @mouseup="handleOutletSwitch(false)" 
            @mouseleave="handleOutletSwitch(false)"> {{ exhaustSwitchFlag ? '排气中...' : '排气' }}
        </button>

        <button class="act-btn" @click="handleVacuumSwitch(!vacuumSwitchFlag)">{{vacuumSwitchFlag?'停止':'抽真空'}}</button>
        <button class="act-btn" @click="handleTailVacuumSwitch(!tailVacuumSwitchFlag)">{{tailVacuumSwitchFlag?'停止':'抽尾部'}}</button>

        <button class="act-btn fire" @click="handleFire">发射</button>

        <button class="act-btn reset" @click="handleReset">恢复</button>
      </footer>
    </main>

    <div v-if="showSettings" class="modal-overlay">
      <div class="modal-content">
        <div class="modal-header">
          <h3>系统配置</h3>
        </div>
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
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, inject} from 'vue'
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime'
import { 
    ConnectPLC, 
    DisconnectPLC,
    GetRealTimeData,
    InletSwitch,
    ExhaustSwitch,
    VacuumSwitch,
    TailVacuumSwitch,
    FireSwitch,
    AutoPressurize,
 } from '../../../wailsjs/go/backend/GasGun1Controller'


const notify = inject('globalNotify')

const deviceIp = ref('192.168.2.1')
const hasAlarm = ref(false)
const isConnected = ref(false)
const inletSwitchFlag = ref(false)//进气开关状态
const exhaustSwitchFlag = ref(false)
const vacuumSwitchFlag = ref(false)
const tailVacuumSwitchFlag = ref(false)
const fireSwitchFlag = ref(false)
const showSettings = ref(false)

const config = reactive({
  ip: '192.168.2.1',
})

const saveSettings = async () => {
  showSettings.value = false
  notify('配置已保存', "success", 2000)
}



const QOStatus = ref([])//输出口状态
const QIStatus = ref([])//输入口状态



const metrics = reactive([
  { label: '供气压力', value: '0.0', unit: 'MPa' },
  { label: '气瓶压力', value: '0.0', unit: 'MPa' },
  { label: '靶室真空度', value: '100000', unit: 'Pa' },
  { label: '尾部真空度', value: '100000', unit: 'Pa' }
])

const logs = ref([
  { time: '14:20:01', msg: '系统自检完成' },
  { time: '14:21:30', msg: 'PLC 连接成功' }
])


onMounted(() => {
    EventsOn("update_metrics", (data) => {
        if (!data) return;

        metrics.forEach(item => {
            if (data[item.label] !== undefined) {
                item.value = data[item.label].toFixed(2);
            }
        });
    })

    EventsOn("WarringStatus", (data) => {
        notify(data.Message, data.Status, 2000)
    })
})

onUnmounted(() => {
  EventsOff("update_metrics")
  EventsOff("WarringStatus")
})


// 连接逻辑
const handleConnect = async (open) => {
  // 连接逻辑
  if (open){
    const response = await ConnectPLC(deviceIp.value)
    isConnected.value = response.Status? true : isConnected.value
    notify(response.Message, response.Status? "success": "error", 2000)
  }else{
    const response = await DisconnectPLC()
    isConnected.value = response.Status? false:isConnected.value
    notify(response.Message, response.Status? "success": "error", 2000)
  }
  
}

// 进气开关
const handleInletSwitch= async (open) => {
  // 进气开关逻辑
  const response = await InletSwitch(open)
  if (response.Status) {
    inletSwitchFlag.value = open
  }
}

// 排气开关
const handleOutletSwitch= async (open) => {
  // 排气开关逻辑
  const response = await ExhaustSwitch(open)
  if (response.Status) {
    exhaustSwitchFlag.value = open
  }
}


const handleVacuumSwitch= async (open) => {
  // 抽真空逻辑
  const response = await VacuumSwitch(open)
  if (response.Status) {
    vacuumSwitchFlag.value = open
  }
}

const handleTailVacuumSwitch= async (open) => {
  // 抽尾部逻辑
  const response = await TailVacuumSwitch(open)
  if (response.Status) {
    tailVacuumSwitchFlag.value = open
  }
}


//发射按钮
const handleFire= async () => {
  // 发射逻辑
  const response = await FireSwitch(500)
  notify(response.Message, response.Status? "success": "error", 2000)
}

//恢复按钮
function handleReset() {
  // 恢复逻辑
}










</script>

<style scoped>

    .gasgun-container {
        display: flex;
        height: 100vh;
        background: #f0f2f5;
        color: #333;
        font-family: 'Segoe UI', sans-serif;
        user-select: none;
        -webkit-user-select: none;
    }

    /* Sidebar Styles */
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
        margin: 0;
    }

    .subtitle { 
        font-size: 10px; 
        color: #4facfe; 
        text-transform: uppercase; 
        margin-top: 4px;
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
        background: #4facfe; 
        color: #fff; 
    }

    .btn-disconnect { 
        background: #333; 
        color: #ccc; 
    }


    /* 日志区 */

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

    .btn-settings-modern:hover { 
        color: #fff; 
        background: #495057; 
    }

    .logo-area {
      text-align: center;
      margin-top: 16px;
      border-top: 1px solid #333;
      padding-top: 12px;
    }

    .logo-text { font-weight: 800; color: #444; letter-spacing: 2px; font-size: 14px; }

    /* Main Content Styles */
    .main-content {
        flex: 1;
        padding: 20px;
        display: flex;
        flex-direction: column;
        height: 100vh;      /* 锁定为屏幕高度 */
        overflow: hidden;    /* 禁止主容器出现滚动条 */
        box-sizing: border-box;
    }

    /* 上方数值区：保持固定高度 */
    .metrics-grid {
    flex-shrink: 0;      /* 禁止缩小 */
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 15px;
    margin-bottom: 20px;
    }
    .metric-card {
        background: #fff;
        padding: 20px;
        border-radius: 12px;
        box-shadow: 0 4px 6px rgba(0,0,0,0.05);
        border-top: 4px solid #4facfe;
    }
    .metric-card .label { font-size: 14px; color: #666; }
    .metric-card .value { font-size: 28px; font-weight: bold; color: #1a1c24; }
    .metric-card .unit { margin-left: 5px; color: #888; }

    /* 中间图片区：核心修正点 */
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
        flex: 1;
        display: flex;
        justify-content: center;
        align-items: center;
        overflow: hidden;    /* 溢出隐藏 */
        padding: 10px;
    }

    .cannon-image {
        max-width: 100%;
        max-height: 100%;    /* 图片高度最大不会超过 .schematic-view 的剩余高度 */
        object-fit: contain; /* 保持比例 */
    }

    .action-bar {
        flex-shrink: 0; 
        height: 120px;
        display: flex;
        gap: 15px;
        align-items: center;
        background: #d3c7c7;
        padding: 10px;
        border-radius: 8px;
    }
    .act-btn {
        flex: 1;
        padding: 15px;
        border: 1px solid #ddd;
        background: #f8f9fa;
        border-radius: 8px;
        font-weight: bold;
        cursor: pointer;
        transition: 0.2s;
        font-size: 1.4rem;
        height: 80%;
    }
    .act-btn.active-pressing { background: #e7f5ff; border-color: #4facfe; color: #4facfe; box-shadow: inset 0 2px 4px rgba(0,0,0,0.1); }
    .act-btn:hover { background: #eef6ff; border-color: #4facfe; }
    .act-btn.fire { background: #ff3b30; color: #fff; border: none; }
    .act-btn.fire:hover { background: #ad8684; }
    .act-btn.reset { background: #8e8e93; color: #fff; border: none; }

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
        display: grid; grid-template-columns: 1fr; gap: 16px;
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

    .btn-cancel { 
        background: #e8eaed; 
        color: #5f6368; 
    }

    .btn-cancel:hover { 
        background: #dadce0;
        color: #3c4043;
    }
</style>