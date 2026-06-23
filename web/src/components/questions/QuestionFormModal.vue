<template>
  <Transition name="modal">
    <div class="modal-overlay" v-if="visible" @click.self="$emit('close')">
      <div class="modal">
        <div class="modal-header">
          <h2>{{ editingQuestion ? '编辑题目' : '新增题目' }}</h2>
          <button class="modal-close" @click="$emit('close')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>模块</label>
            <select v-model="localForm.moduleId" class="form-input">
              <option value="">请选择模块</option>
              <option v-for="mod in modules" :key="mod.id" :value="mod.id">{{ mod.name }}</option>
            </select>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>题型</label>
              <select v-model="localForm.type" class="form-input">
                <option value="single">单选题</option>
                <option value="multi">多选题</option>
                <option value="judge">判断题</option>
                <option value="fill">填空题</option>
              </select>
            </div>
            <div class="form-group">
              <label>难度</label>
              <select v-model="localForm.difficulty" class="form-input">
                <option value="1">★</option>
                <option value="2">★★</option>
                <option value="3">★★★</option>
                <option value="4">★★★★</option>
                <option value="5">★★★★★</option>
              </select>
            </div>
          </div>
          <div class="form-group">
            <label>题干</label>
            <textarea v-model="localForm.content" class="form-input textarea" rows="3"></textarea>
          </div>
          <div class="form-group">
            <label>选项（JSON 数组格式）</label>
            <textarea v-model="localForm.options" class="form-input textarea mono" rows="3" placeholder='["A. 选项A", "B. 选项B", "C. 选项C", "D. 选项D"]'></textarea>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>正确答案</label>
              <input v-model="localForm.answer" class="form-input" placeholder="如: A 或 A,B,C" />
            </div>
            <div class="form-group">
              <label>来源</label>
              <input v-model="localForm.source" class="form-input" placeholder="如: 2024国考" />
            </div>
          </div>
          <div class="form-group">
            <label>解析</label>
            <textarea v-model="localForm.analysis" class="form-input textarea" rows="3"></textarea>
          </div>
          <div class="form-group">
            <label>标签（逗号分隔）</label>
            <input v-model="localForm.tags" class="form-input" placeholder="如: 言语理解,错别字" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-ghost" @click="$emit('close')">取消</button>
          <button class="btn btn-primary" @click="handleSave" :disabled="saving">
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { reactive, watch } from 'vue'
import { showToast } from '../../utils/toast'

const props = defineProps({
  visible: { type: Boolean, default: false },
  editingQuestion: { type: Object, default: null },
  modules: { type: Array, default: () => [] },
  form: { type: Object, default: () => ({ moduleId: '', type: 'single', content: '', options: '', answer: '', analysis: '', difficulty: 1, tags: '', source: '' }) },
  saving: { type: Boolean, default: false },
})

const emit = defineEmits(['save', 'close'])

const localForm = reactive({
  moduleId: '',
  type: 'single',
  content: '',
  options: '',
  answer: '',
  analysis: '',
  difficulty: 1,
  tags: '',
  source: '',
})

watch(() => props.form, (newForm) => {
  Object.assign(localForm, newForm)
}, { deep: true, immediate: true })

watch(() => props.visible, (val) => {
  if (val) {
    Object.assign(localForm, props.form)
  }
})

function handleSave() {
  if (!localForm.moduleId || !localForm.content || !localForm.answer) {
    showToast('请填写模块、题干和正确答案', 'error')
    return
  }
  emit('save', {
    module_id: parseInt(localForm.moduleId),
    type: localForm.type,
    content: localForm.content,
    options: localForm.options,
    answer: localForm.answer,
    analysis: localForm.analysis,
    difficulty: parseInt(localForm.difficulty),
    tags: localForm.tags,
    source: localForm.source,
  })
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.4);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: var(--bg-card);
  border-radius: var(--radius-xl);
  width: 92%;
  max-width: 560px;
  max-height: 85vh;
  overflow-y: auto;
  box-shadow: var(--shadow-lg);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border);
}

.modal-header h2 {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text);
  margin: 0;
}

.modal-close {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-muted);
  padding: 0.25rem;
  border-radius: var(--radius-sm);
  transition: var(--transition);
  display: flex;
}

.modal-close svg {
  width: 18px;
  height: 18px;
}

.modal-close:hover {
  background: var(--bg-hover);
  color: var(--text);
}

.modal-body {
  padding: 1.25rem 1.5rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 0.3rem;
}

.form-input {
  width: 100%;
  padding: 0.55rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  font-size: 0.875rem;
  color: var(--text);
  background: var(--bg-card);
  outline: none;
  transition: var(--transition);
}

.form-input:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-bg);
}

.form-input.textarea {
  resize: vertical;
  min-height: 72px;
}

.form-input.mono {
  font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
  font-size: 0.8rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.85rem;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border);
}

.btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: var(--radius);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: var(--transition);
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-primary {
  background: var(--primary);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--primary-dark);
}

.btn-ghost {
  background: transparent;
  color: var(--text-secondary);
  border: 1px solid var(--border);
}

.btn-ghost:hover:not(:disabled) {
  background: var(--bg-hover);
  border-color: var(--text-muted);
}

/* Modal transition */
.modal-enter-active, .modal-leave-active {
  transition: all 0.2s ease;
}

.modal-enter-from, .modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal, .modal-leave-to .modal {
  transform: scale(0.95) translateY(8px);
}

@media (max-width: 768px) {
  .modal {
    width: 96%;
    max-height: 90vh;
  }

  .modal-header {
    padding: 1rem 1.25rem;
  }

  .modal-body {
    padding: 1rem 1.25rem;
  }

  .form-row {
    grid-template-columns: 1fr;
    gap: 0;
  }

  .modal-footer {
    padding: 0.75rem 1.25rem;
  }
}
</style>
