<template>
  <Transition name="fade">
    <div class="update-card">
        <div class="update-icon">🚀</div>
        <h2>发现新版本！</h2>
        <p class="version-text">最新版本: {{ latestInfo.version }}</p>
        <p class="current-text">当前版本: {{ currentVersion }}</p>
        <p class="update-text">更新内容: {{ latestInfo.description }}</p>
        
        <div class="btn-group">
          <button class="btn-cancel" @click="close">稍后再说</button>
          <button class="btn-confirm" @click="doDownload">立即下载更新</button>
        </div>
      </div>
  </Transition>
</template>

<script setup>
import { ref } from 'vue'
import { BrowserOpenURL } from '../../../wailsjs/runtime'

const props = defineProps({
})

const currentVersion = "V1.0.0" // 当前版本
const latestInfo = ref({ version: '', url: '' , description: ''}) // 最新版本信息


const emit = defineEmits(['close'])

const close = () => {
  emit('close') // 向上发送关闭事件
}

const doDownload = () => {
  BrowserOpenURL(latestInfo.value.url) 
  close()
}


</script>


<style scoped>
.update-card {
  background: #2c3e50;
  padding: 30px;
  border-radius: 16px;
  text-align: center;
  width: 320px;
  border: 1px solid #34495e;
  box-shadow: 0 10px 30px rgba(0,0,0,0.5);
}
.update-icon { font-size: 50px; margin-bottom: 10px; }
.version-text { color: #2ecc71; font-weight: bold; margin: 10px 0 5px; }
.current-text { color: #95a5a6; font-size: 13px; margin-bottom: 20px; }

.btn-group { display: flex; gap: 10px; }
.btn-cancel {
  flex: 1;
  padding: 10px;
  background: #34495e;
  border: none;
  color: #bdc3c7;
  border-radius: 8px;
  cursor: pointer;
}
.btn-confirm {
  flex: 2;
  padding: 10px;
  background: #3498db;
  border: none;
  color: white;
  border-radius: 8px;
  font-weight: bold;
  cursor: pointer;
}
.btn-confirm:hover { background: #2980b9; }

.fade-enter-active, .fade-leave-active { transition: opacity 0.3s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
