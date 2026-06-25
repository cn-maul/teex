<template>
  <Transition name="modal">
    <div class="modal-overlay" v-if="visible" @click.self="$emit('close')">
      <div class="modal-container">
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
.form-input.textarea {
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

.modal-container {
  max-width: 560px;
}

@media (max-width: 768px) {
  .form-row {
    grid-template-columns: 1fr;
    gap: 0;
  }
}
</style>
