<template>
  <div v-if="show" class="modal-overlay" @click.self="close">
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
              <label>气瓶增压阀(Y):</label>
              <input v-model.number="config.switches.Pressurize" type="number" />
            </div>
            <div class="setting-group">
              <label>气瓶减压阀(Y):</label>
              <input v-model.number="config.switches.Decompress" type="number" />
            </div>
            <div class="setting-group">
              <label>泵管增压阀(Y):</label>
              <input v-model.number="config.switches.PumpTubePressurize" type="number" />
            </div>
            <div class="setting-group">
              <label>泵管减压阀(Y):</label>
              <input v-model.number="config.switches.PumpTubeDecompress" type="number" />
            </div>
            <div class="setting-group">
              <label>泵管真空阀(Y):</label>
              <input v-model.number="config.switches.PumpTubeVacuum" type="number" />
            </div>
            <div class="setting-group">
              <label>靶室真空阀(Y):</label>
              <input v-model.number="config.switches.TargetVacuum" type="number" />
            </div>
            <div class="setting-group">
              <label>尾部真空阀(Y):</label>
              <input v-model.number="config.switches.TailVacuumProtect" type="number" />
            </div>
            <div class="setting-group">
              <label>泵管保护阀(Y):</label>
              <input v-model.number="config.switches.PumpTubeProtect" type="number" />
            </div>
            <div class="setting-group">
              <label>系统发射阀(Y):</label>
              <input v-model.number="config.switches.FireSwitch" type="number" />
            </div>
            <div class="setting-group">
              <label>系统排气阀(Y):</label>
              <input v-model.number="config.switches.SystemDecompress" type="number" />
            </div>
          </div>
        </div>

        <div class="settings-section">
          <h4 class="settings-section-title">真空泵点位配置</h4>
          <div class="settings-grid-2col">
            <div class="setting-group">
              <label>靶室真空泵(Y):</label>
              <input v-model.number="config.switches.TargetVacuumPump" type="number" />
            </div>
            <div class="setting-group">
              <label>尾部真空泵(Y):</label>
              <input v-model.number="config.switches.TailVacuumPump" type="number" />
            </div>
          </div>
        </div>

        <div class="settings-section">
          <h4 class="settings-section-title">数据地址配置</h4>
          <div class="settings-grid-2col">
            <div class="setting-group">
              <label>气路输入压力(D):</label>
              <input v-model.number="config.dataAddresses.InputPressure" type="number" />
            </div>
            <div class="setting-group">
              <label>气瓶压力(D):</label>
              <input v-model.number="config.dataAddresses.CylinderPressure" type="number" />
            </div>
            <div class="setting-group">
              <label>泵管压力(D):</label>
              <input v-model.number="config.dataAddresses.PumpTubePressure" type="number" />
            </div>
            <div class="setting-group">
              <label>泵管高精度压力(D):</label>
              <input v-model.number="config.dataAddresses.PumpTubePressureHi" type="number" />
            </div>
            <div class="setting-group">
              <label>靶室真空度(D):</label>
              <input v-model.number="config.dataAddresses.TargetVacuumDegree" type="number" />
            </div>
            <div class="setting-group">
              <label>尾部真空度(D):</label>
              <input v-model.number="config.dataAddresses.TailVacuumDegree" type="number" />
            </div>
          </div>
        </div>
      </div>
      <div class="modal-actions">
        <button class="modal-btn btn-cancel" @click="close">取消</button>
        <button class="modal-btn btn-confirm" @click="$emit('save')">确认保存</button>
      </div>
    </div>
  </div>
</template>


<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
// ===== 配置 =====
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

defineProps({
  show: { type: Boolean, default: false },
})

const emit = defineEmits(['update:show', 'save'])

const close = () => emit('update:show', false)
</script>

<style scoped>
/* ===== 配置弹窗 ===== */
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
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3), 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.config-modal {
  width: 650px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  background: #4e69b5;
  padding: 20px 24px;
  color: white;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.settings-scroll {
  overflow-y: auto;
  max-height: 60vh;
  padding: 24px;
}

.settings-scroll::-webkit-scrollbar {
  width: 6px;
}

.settings-scroll::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
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
  border-left: 4px solid #4e69b5;
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
  border-color: #4e69b5;
  background: #ffffff;
  box-shadow: 0 0 0 3px rgba(78, 105, 181, 0.1);
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
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.modal-btn {
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
  background: #4e69b5;
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
