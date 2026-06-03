<script setup>
import ChooseUtil from './components/ChooseUtil.vue'
import Gasgun1 from './components/gasgun/Gasgun1.vue'
import Gasgun2 from './components/gasgun/Gasgun2.vue'
import NormalHopkinson from './components/gasgun/NormalHopkinson.vue'
import { ref, provide, onMounted } from 'vue'
import { Quit, BrowserOpenURL } from '../wailsjs/runtime' // 导入 Wails 退出函数
import MessageContainer from './components/utils/MessageContainer.vue'
import Update from './components/utils/update.vue'

import { CallGasgun1 ,
    CallNormalHopkinson,
    APIUpdate,
} from '../wailsjs/go/main/APP'

// 状态声明
const chooseUitls = ref(true)
const Gasgun1Enable = ref(false)
const Gasgun2Enable = ref(false)
const NormalHopkinsonEnable = ref(false)
const showUpdateModal = ref(false) // 默认不显示，检测到更新后再弹出

const msgBoxRef = ref(null)

const notify = (content, type = 'info', duration = 1000) => {
  msgBoxRef.value?.addMessage(content, type, duration)
}

provide('globalNotify', notify)

const updateInfo = ref({
  tagName: '无',
  htmlUrl: 'https://github.com'
})

onMounted(async() => {
  try {
    const release = await APIUpdate()
    if (release) {
        updateInfo.value.tagName = release.tag_name
        updateInfo.value.htmlUrl = release.html_url
        if (release.assets.length > 0) {
            updateInfo.value.htmlUrl = release.assets[0].browser_download_url
        }
        // 发现更新时，显示更新模态框
        showUpdateModal.value = true
    }
  } catch (e) {
    console.error("检查更新失败", e)
  }
})

// 更新处理函数
const handleUpdate = () => {
    if (updateInfo.value.htmlUrl) {
        BrowserOpenURL(updateInfo.value.htmlUrl) 
        showUpdateModal.value = false;
    }
};

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
  else if (type === 'gasgun2') {
    Gasgun2Enable.value = true
  }
  console.log('用户选择了:', type)
}

const onExit = () => {
  Quit() 
}
</script>

<template>
  <ChooseUtil 
    v-if="chooseUitls" 
    @confirm="onSelected" 
    @exit="onExit"
  />

  <NormalHopkinson v-if="NormalHopkinsonEnable" />

  <Gasgun1 v-if="Gasgun1Enable" />

  <Gasgun2 v-if="Gasgun2Enable" />

  <MessageContainer ref="msgBoxRef" />

  <!-- 传送更新模态框到 body -->
  <teleport to="body">
    <transition name="modal">
      <div v-if="showUpdateModal" class="modal-overlay" @click.self="showUpdateModal = false">
        <div class="modal-container update-modal">
          <div class="modal-header">
            <h3 class="modal-title">发现新版本</h3>
            <button class="modal-close" @click="showUpdateModal = false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"/>
                <line x1="6" y1="6" x2="18" y2="18"/>
              </svg>
            </button>
          </div>
          <div class="modal-body">
            <div class="update-content">
              <div class="update-icon">
                <svg viewBox="-2 -2 28 28" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/>
                  <path d="M3 3v6h6"/>
                  <path d="M3 16a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16"/>
                  <path d="M16 17h6v6"/>
                </svg>
              </div>
              <div class="update-info">
                <p class="update-version">新版本: <strong>{{ updateInfo.tagName }}</strong></p>
                <p class="update-desc">发现应用程序有新版本可用，建议及时升级以获得更好的体验和稳定性。</p>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showUpdateModal = false">
              稍后更新
            </button>
            <button class="btn btn-primary" @click="handleUpdate">
              立即更新
            </button>
          </div>
        </div>
      </div>
    </transition>
  </teleport>
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

/* 遮罩层 */
.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
    backdrop-filter: blur(4px);
}

/* 模态框容器 */
.modal-container {
    background: linear-gradient(180deg, rgba(30, 41, 59, 0.95) 0%, rgba(15, 23, 42, 0.98) 100%);
    border-radius: 16px;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
    border: 1px solid rgba(148, 163, 184, 0.1);
    overflow: hidden;
}

/* 更新模态框独有尺寸 */
.update-modal {
    width: 420px;
    max-width: 90%;
    border: 1px solid rgba(148, 163, 184, 0.15);
}

.modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
}

.modal-title {
    margin: 0;
    font-size: 15px;
    font-weight: 600;
    color: #cbd5e1;
}

.modal-close {
    background: transparent;
    border: none;
    color: #94a3b8;
    cursor: pointer;
    padding: 4px;
    border-radius: 8px;
    transition: all 0.2s;
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.modal-close:hover {
    background: rgba(255, 255, 255, 0.1);
    color: #f1f5f9;
}

.modal-close svg {
    width: 18px;
    height: 18px;
}

.modal-body {
    padding: 20px;
}

/* 新版本详情区域布局 */
.update-content {
    display: flex;
    align-items: flex-start;
    gap: 16px;
}

/* 图标圆圈与动画 */
.update-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 66px;
    height: 66px;
    background: rgba(59, 130, 246, 0.12);
    color: #3b82f6;
    border-radius: 12px;
    flex-shrink: 0;
}

.update-icon svg {
    width: 34px;
    height: 34px;
    animation: spin-slow 12s linear infinite; /* 缓慢旋转模拟同步更新质感 */
}

@keyframes spin-slow {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
}

.update-info {
    display: flex;
    flex-direction: column;
    gap: 6px;
}

.update-version {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: #f1f5f9;
}

.update-version strong {
    color: #3b82f6;
    background: rgba(59, 130, 246, 0.15);
    padding: 2px 8px;
    border-radius: 6px;
    font-size: 13px;
    margin-left: 6px;
}

.update-desc {
    margin: 0;
    font-size: 13px;
    line-height: 1.5;
    color: #94a3b8;
}

.modal-footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 12px;
    padding: 16px 20px;
}

/* 通用按钮系统定义 */
.btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 8px 16px;
    font-size: 13px;
    font-weight: 500;
    border-radius: 8px;
    border: none;
    cursor: pointer;
    transition: all 0.2s ease;
    outline: none;
}

.btn-secondary {
    background: rgba(148, 163, 184, 0.1);
    color: #cbd5e1;
}

.btn-secondary:hover {
    background: rgba(148, 163, 184, 0.18);
    color: #f1f5f9;
}

.btn-primary {
    background: #2563eb;
    color: #ffffff;
    box-shadow: 0 4px 12px rgba(37, 99, 235, 0.2);
}

.btn-primary:hover {
    background: #1d4ed8;
    box-shadow: 0 4px 16px rgba(37, 99, 235, 0.35);
    transform: translateY(-1px);
}

.btn-primary:active {
    transform: translateY(0);
}

/* 模态框 Transition 动画样式 */
.modal-enter-active,
.modal-leave-active {
    transition: opacity 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
    opacity: 0;
}

.modal-enter-active .modal-container,
.modal-leave-active .modal-container {
    transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1), opacity 0.3s ease;
}

.modal-enter-from .modal-container,
.modal-leave-to .modal-container {
    transform: scale(0.9) translateY(12px);
    opacity: 0;
}
</style>