<template>
  <div v-if="show" class="modal-overlay" @click.self="close">
    <div class="modal-content">
      <div class="modal-header">
        <h3>系统配置</h3>
      </div>
      <div class="settings-grid">
        <div class="setting-group">
          <label>默认IP:</label>
          <input :value="ip" placeholder="192.168.2.1" @input="$emit('update:ip', $event.target.value)" />
        </div>
      </div>
      <div class="modal-actions">
        <button class="btn-settings btn-cancel" @click="close">取消</button>
        <button class="btn-settings btn-confirm" @click="$emit('save')">确认保存</button>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  show: { type: Boolean, default: false },
  ip: { type: String, default: '' },
})

const emit = defineEmits(['update:show', 'update:ip', 'save'])

const close = () => emit('update:show', false)
</script>

<style scoped>
/* ===== 配置弹窗 ===== */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  justify-content: center;
  align-items: center;
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
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3), 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.modal-header {
  background: linear-gradient(135deg, #4e69b5 0%, #6d8ad6 100%);
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
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
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
  border-color: #4e69b5;
  background: #ffffff;
  box-shadow: 0 0 0 3px rgba(78, 105, 181, 0.1);
}

.modal-actions {
  padding: 16px 24px;
  background: #f8f9fa;
  border-top: 1px solid #e8eaed;
  display: flex;
  gap: 12px;
  justify-content: flex-end;
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
  background: linear-gradient(135deg, #4e69b5 0%, #6d8ad6 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(78, 105, 181, 0.3);
}

.btn-confirm:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(78, 105, 181, 0.4);
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
