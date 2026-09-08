<template>
  <div class="main-window">
    <div class="title-bar" @dblclick="WindowToggleMaximise">

      <div class="logo"> 
        <img v-if="User=='NIMTE'" src="./assets/images/logo/nimte.png" width="100" height="25">
        <img v-if="User=='PIMS'"  src="./assets/images/logo/pims.png" width="100" height="25">
        <!-- <img v-if="User=='ADMIN'" src="./assets/images/admin/logo.png" width="100" height="25"> -->
      </div>

      <div class="title">力学实验室控制系统</div>
      <div class="sub-title" v-if="currentDevice">({{ currentDeviceName }})</div>

      <div class='window-actions'>
        <button class="window-btn switch-device-btn" type="button" title="切换设备" @click="showDeviceModal = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 7v6h-6"/>
                <path d="M3 17v-6h6"/>
                <path d="M21 13a9 9 0 0 0-15-6.7L3 13"/>
                <path d="M3 11a9 9 0 0 0 15 6.7L21 11"/>
            </svg>
        </button>

        <button class="window-btn alarm-history-btn" type="button" title="历史报警消息" @click="alarmHistoryVisible = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/>
                <path d="M13.73 21a2 2 0 0 1-3.46 0"/>
            </svg>
        </button>

        <button class="window-btn settings-btn" type="button" title="系统设置" @click="ShowSetModal = true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="3"/>
                <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.6 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.6a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
            </svg>
        </button>

        <span class="title-divider"></span>

        <button class="window-btn minimize-btn" type="button" title="最小化" @click="WindowMinimise">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="4" y1="12" x2="22" y2="12" />
          </svg>
        </button>
        <button class="window-btn maximize-btn" type="button" title="最大化" @click="WindowToggleMaximise">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="5" y="5" width="17" height="17" rx="2" />
          </svg>
        </button>
        <button class="window-btn close-btn" type="button" title="关闭" @click="Quit">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="20" y1="4" x2="4" y2="20" />
            <line x1="4" y1="4" x2="20" y2="20" />
          </svg>
        </button>
      </div>
    </div>

    <div class="content">
      <XINJIE_Gasgun1 v-if="currentDevice === 'xinjie-gasgun1'" />
      <XINJIE_Gasgun2 v-if="currentDevice === 'xinjie-gasgun2'" />
    </div>

    <XINJIE_Gasgun2_SetModal v-if="currentDevice==='xinjie-gasgun2'" v-model:show="ShowSetModal" @save="ShowSetModal = false" />
    <XINJIE_Gasgun1_SetModal v-if="currentDevice==='xinjie-gasgun1'" v-model:show="ShowSetModal" @save="ShowSetModal = false" />
    

    <!-- 设备选择模态框 -->
    <div class="modal-mask" v-if="showDeviceModal" @click.self="showDeviceModal = false">
      <div class="modal-container">
        <h2 class="modal-title">设备选择</h2>

        <div class="options-group">
          <div
            v-for="opt in Devices[User]"
            :key="opt.id"
            :class="['option-item', { active: tempSelected === opt.id }]"
            @click="tempSelected = opt.id"
          >
            <div class="radio-dot"></div>
            <span class="option-text">{{ opt.name }}</span>
          </div>
        </div>

        <div class="button-group">
          <button class="btn btn-exit" @click="Quit">退出软件</button>
          <button class="btn btn-confirm" :disabled="!tempSelected" @click="handleDeviceConfirm">确认进入</button>
        </div>
      </div>
    </div>
  </div>
</template>


<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { WindowMinimise, WindowToggleMaximise, WindowMaximise, Quit } from '../wailsjs/runtime/runtime'



import XINJIE_Gasgun1 from './components/xinjiegasgun1/GasGunWindow.vue'
import XINJIE_Gasgun2 from './components/xinjiegasgun2/GasGunWindow.vue'
import XINJIE_Gasgun1_SetModal from './components/xinjiegasgun1/SystemSetModal.vue'
import XINJIE_Gasgun2_SetModal from './components/xinjiegasgun2/SystemSetModal.vue'



const ShowSetModal = ref(false)
const showDeviceModal = ref(true)
const tempSelected = ref('xinjie-gasgun1')
const currentDevice = ref('xinjie-gasgun1')

const User = ref('PIMS')

const Devices = ref({
    PIMS:[
      { id: 'xinjie-gasgun1', name: '一级气炮' },
      { id: 'xinjie-gasgun2', name: '二级气炮' },
      { id: 'gongbei-hopkinson', name: '常温Hopkinson杆' },
    ],
    NIMTE:[
      { id: 'xinjie-gasgun1', name: '一级气炮' },
    ],
    SWJTU:[
      { id: 'swjtu-gasgun1', name: '一级气炮' },
    ],
    HEPS:[
      { id: 'heps-gasgun1', name: '一级气炮' },
    ]
  }
)


const currentDeviceName = computed(() => {
  const list = Devices.value[User.value] || []
  const opt = list.find(o => o.id === currentDevice.value)
  return opt ? opt.name : ''
})

const handleDeviceConfirm = () => {
  if (tempSelected.value) {
    currentDevice.value = tempSelected.value
    showDeviceModal.value = false
  }
}

onMounted(() => {
  WindowMaximise()
})

onUnmounted(() => {
})



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
.sub-title {
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

/* ===== 设备选择模态框 ===== */
.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  justify-content: center;
  align-items: center;
  backdrop-filter: blur(4px);
  z-index: 9999;
}

.modal-container {
  background: #ffffff;
  width: 350px;
  padding: 30px;
  border-radius: 16px;
  box-shadow: 0 10px 25px rgba(0,0,0,0.3);
  text-align: center;
}

.modal-title {
  margin-bottom: 25px;
  color: #333;
  font-weight: 600;
  letter-spacing: 1px;
}

.options-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 30px;
}

.option-item {
  padding: 15px;
  border: 2px solid #eee;
  border-radius: 10px;
  display: flex;
  align-items: center;
  cursor: pointer;
  transition: all 0.2s ease;
}

.option-item:hover {
  background: #f8f9fa;
  border-color: #007aff;
}

.option-item.active {
  background: #eef6ff;
  border-color: #007aff;
}

.radio-dot {
  width: 12px;
  height: 12px;
  border: 2px solid #ddd;
  border-radius: 50%;
  margin-right: 15px;
  position: relative;
}

.active .radio-dot {
  border-color: #007aff;
  background: #007aff;
  box-shadow: inset 0 0 0 2px #fff;
}

.option-text {
  font-size: 16px;
  color: #444;
  font-weight: 500;
}

.button-group {
  display: flex;
  gap: 15px;
}

.btn {
  flex: 1;
  padding: 12px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: opacity 0.2s;
}

.btn-confirm {
  background: #007aff;
  color: white;
}

.btn-confirm:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.btn-exit {
  background: #f2f2f7;
  color: #ff3b30;
}

.btn:hover:not(:disabled) {
  opacity: 0.8;
}

</style>
