<template>
  <div class="modal-mask">
    <div class="modal-container">
      <h2 class="modal-title">设备选择系统</h2>
      
      <div class="options-group">
        <div 
          v-for="opt in options" 
          :key="opt.id"
          :class="['option-item', { active: selected === opt.id }]"
          @click="selected = opt.id"
        >
          <div class="radio-dot"></div>
          <span class="option-text">{{ opt.name }}</span>
        </div>
      </div>

      <div class="button-group">
        <button class="btn btn-exit" @click="$emit('exit')">退出软件</button>
        <button class="btn btn-confirm" :disabled="!selected" @click="handleConfirm">确认进入</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const selected = ref(null)
const options = [
  { id: 'gasgun1', name: '一级气炮' },
  { id: 'gasgun2', name: '二级气炮' },
  { id: 'hopkinson', name: '常温Hopkinson杆' },
  { id: '15mmbirdgun', name: '15mm鸟枪' },

]

const emit = defineEmits(['confirm', 'exit'])

const handleConfirm = () => {
  if (selected.value) {
    emit('confirm', selected.value)
  }
}
</script>

<style scoped>
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