<script setup>
import ChooseUtil from './components/ChooseUtil.vue'
import Gasgun1 from './components/gasgun/Gasgun1.vue'
// import Gasgun2 from './components/gasgun/Gasgun2.vue'
import NormalHopkinson from './components/gasgun/NormalHopkinson.vue'
import { ref, provide, onMounted } from 'vue'
import { Quit } from '../wailsjs/runtime' // 导入 Wails 退出函数
import MessageContainer from './components/utils/MessageContainer.vue'

import Update from './components/utils/update.vue'

import { CallGasgun1 ,
    CallNormalHopkinson,
    GetLatestRelease,
} from '../wailsjs/go/main/APP'

const chooseUitls = ref(true)
const Gasgun1Enable = ref(false)
const Gasgun2Enable = ref(false)
const NormalHopkinsonEnable = ref(false)

const msgBoxRef = ref(null)

const notify = (content, type = 'info', duration = 3000) => {
  msgBoxRef.value?.addMessage(content, type, duration)
}

provide('globalNotify', notify)

onMounted(async() => {
  try {
    const res = await GetLatestRelease()
    if (res && res.version !== currentVersion && res.version!="") {
      latestInfo.value = {
        version: res.version,
        url: res.downloadUrl
      }
      showUpdateModal.value = true // 弹出更新模态框
    }
  } catch (e) {
    console.error("检查更新失败", e)
  }
})


const onSelected = async(type) => {
  chooseUitls.value = false
  
  if (type === 'hopkinson') {
    NormalHopkinsonEnable.value = true
    await CallNormalHopkinson()
    
  }
  else if (type === 'gasgun1') {
    Gasgun1Enable.value = true
    await CallGasgun1()
  }

  else if (type === 'gasgun2') Gasgun2Enable.value = true
  console.log('用户选择了:', type)
}

const onExit = () => {
  Quit() 
}

const showUpdateModal = ref(true)



</script>

<template>
  <ChooseUtil 
    v-if="chooseUitls" 
    @confirm="onSelected" 
    @exit="onExit"
  />

  <NormalHopkinson v-if="NormalHopkinsonEnable" />

  <Gasgun1 v-if="Gasgun1Enable" />

  <div v-if="Gasgun2Enable" class="main-content">
    <h1>二级气炮控制系统:还未实现，退出重新进入并选择一级气炮</h1>
  </div>

  <MessageContainer ref="msgBoxRef" />

  <div v-if="showUpdateModal" class="update-modal-overlay">
      <Update @close="showUpdateModal = false"/>
  </div>



</template>

<style>
.main-content {
  padding: 20px;
  animation: fadeIn 0.5s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.update-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.75);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 10000;
  backdrop-filter: blur(4px);
}
</style>