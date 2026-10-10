<template>
  <div class="update-mask" @click.self="$emit('close')">
    <div class="update-card">
      <div class="update-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="7 10 12 15 17 10"/>
          <line x1="12" y1="15" x2="12" y2="3"/>
        </svg>
      </div>
      <div class="update-content">
        <div class="update-text">发现更新版本：{{ version }}</div>
        <div class="update-actions">
          <button class="update-btn" @click="handleUpdate">更新</button>
          <button class="cancel-btn" @click="$emit('close')">取消</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'

const props = defineProps({
  version: { type: String, default: '' },
  url: { type: String, default: '' }
})
const emit = defineEmits(['close'])

const handleUpdate = () => {
  BrowserOpenURL(props.url)
  emit('close')
}
</script>

<style scoped>
.update-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 10000;
  backdrop-filter: blur(4px);
}

.update-card {
  background: #ffffff;
  border-radius: 12px;
  padding: 30px 36px;
  display: flex;
  align-items: center;
  gap: 24px;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.25);
  min-width: 380px;
}

.update-icon {
  flex-shrink: 0;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: #eef6ff;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #007aff;
}

.update-icon svg {
  width: 28px;
  height: 28px;
}

.update-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.update-text {
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.update-actions {
  display: flex;
  gap: 12px;
}

.update-btn,
.cancel-btn {
  padding: 8px 24px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: opacity 0.2s;
}

.update-btn {
  background: #007aff;
  color: #fff;
}

.cancel-btn {
  background: #f2f2f7;
  color: #555;
}

.update-btn:hover,
.cancel-btn:hover {
  opacity: 0.85;
}
</style>
