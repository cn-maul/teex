<template>
  <div class="question-manage">
    <div class="page-header">
      <h1>题目管理</h1>
      <div class="header-actions">
        <button class="btn btn-primary" @click="showAddModal = true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
          新增题目
        </button>
        <label class="btn btn-ghost import-btn">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
          导入 JSON
          <input type="file" accept=".json" @change="handleImport" hidden />
        </label>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <select v-model="filter.moduleId" class="filter-select" @change="loadQuestions">
        <option value="">全部模块</option>
        <option v-for="mod in modules" :key="mod.id" :value="mod.id">{{ mod.name }}</option>
      </select>
      <select v-model="filter.type" class="filter-select" @change="loadQuestions">
        <option value="">全部题型</option>
        <option value="single">单选题</option>
        <option value="multi">多选题</option>
        <option value="judge">判断题</option>
        <option value="fill">填空题</option>
      </select>
      <select v-model="filter.difficulty" class="filter-select" @change="loadQuestions">
        <option value="">全部难度</option>
        <option value="1">★☆☆☆☆</option>
        <option value="2">★★☆☆☆</option>
        <option value="3">★★★☆☆</option>
        <option value="4">★★★★☆</option>
        <option value="5">★★★★★</option>
      </select>
    </div>

    <!-- 题目列表 -->
    <div class="table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th style="width: 56px">ID</th>
            <th style="width: 72px">题型</th>
            <th>题干</th>
            <th style="width: 72px">答案</th>
            <th style="width: 80px">难度</th>
            <th style="width: 96px">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(q, idx) in questions" :key="q.id" :class="{ 'row-alt': idx % 2 === 1 }">
            <td class="cell-id">{{ q.id }}</td>
            <td>
              <span class="type-badge">{{ getTypeLabel(q.type) }}</span>
            </td>
            <td class="cell-content">{{ truncate(q.content, 60) }}</td>
            <td class="cell-answer"><strong>{{ q.answer }}</strong></td>
            <td>
              <span class="difficulty-stars">
                <span v-for="i in 5" :key="i" class="star" :class="{ filled: i <= q.difficulty }">★</span>
              </span>
            </td>
            <td class="cell-actions">
              <button class="btn-icon" @click="editQuestion(q)" title="编辑">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
              </button>
              <button class="btn-icon btn-danger" @click="confirmDelete(q)" title="删除">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
              </button>
            </td>
          </tr>
          <tr v-if="questions.length === 0 && !loading">
            <td colspan="6" class="empty-row">暂无题目</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页 -->
    <div class="pagination" v-if="total > pageSize">
      <button class="btn-page" :disabled="currentPage <= 1" @click="changePage(currentPage - 1)">
        ← 上一页
      </button>
      <span class="page-info">第 {{ currentPage }} / {{ totalPages }} 页（共 {{ total }} 题）</span>
      <button class="btn-page" :disabled="currentPage >= totalPages" @click="changePage(currentPage + 1)">
        下一页 →
      </button>
    </div>

    <!-- 新增/编辑弹窗 -->
    <Transition name="modal">
      <div class="modal-overlay" v-if="showAddModal || editingQuestion" @click.self="closeModal">
        <div class="modal">
          <div class="modal-header">
            <h2>{{ editingQuestion ? '编辑题目' : '新增题目' }}</h2>
            <button class="modal-close" @click="closeModal">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
            </button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label>模块</label>
              <select v-model="form.moduleId" class="form-input">
                <option value="">请选择模块</option>
                <option v-for="mod in modules" :key="mod.id" :value="mod.id">{{ mod.name }}</option>
              </select>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label>题型</label>
                <select v-model="form.type" class="form-input">
                  <option value="single">单选题</option>
                  <option value="multi">多选题</option>
                  <option value="judge">判断题</option>
                  <option value="fill">填空题</option>
                </select>
              </div>
              <div class="form-group">
                <label>难度</label>
                <select v-model="form.difficulty" class="form-input">
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
              <textarea v-model="form.content" class="form-input textarea" rows="3"></textarea>
            </div>
            <div class="form-group">
              <label>选项（JSON 数组格式）</label>
              <textarea v-model="form.options" class="form-input textarea mono" rows="3" placeholder='["A. 选项A", "B. 选项B", "C. 选项C", "D. 选项D"]'></textarea>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label>正确答案</label>
                <input v-model="form.answer" class="form-input" placeholder="如: A 或 A,B,C" />
              </div>
              <div class="form-group">
                <label>来源</label>
                <input v-model="form.source" class="form-input" placeholder="如: 2024国考" />
              </div>
            </div>
            <div class="form-group">
              <label>解析</label>
              <textarea v-model="form.analysis" class="form-input textarea" rows="3"></textarea>
            </div>
            <div class="form-group">
              <label>标签（逗号分隔）</label>
              <input v-model="form.tags" class="form-input" placeholder="如: 言语理解,错别字" />
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="closeModal">取消</button>
            <button class="btn btn-primary" @click="saveQuestion" :disabled="saving">
              {{ saving ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import {
  getQuestions, createQuestion, updateQuestion, deleteQuestion,
  getExamModules, importQuestions
} from '../api'
import { useExamStore } from '../stores/exam'

const examStore = useExamStore()

const questions = ref([])
const modules = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = 20
const loading = ref(false)
const saving = ref(false)

const filter = ref({ moduleId: '', type: '', difficulty: '' })
const showAddModal = ref(false)
const editingQuestion = ref(null)

const form = ref({
  moduleId: '', type: 'single', content: '', options: '', answer: '',
  analysis: '', difficulty: 1, tags: '', source: '',
})

const totalPages = computed(() => Math.ceil(total.value / pageSize))

watch(() => examStore.state.currentExamId, async () => {
  filter.value.moduleId = ''
  await loadModules()
  currentPage.value = 1
  await loadQuestions()
}, { immediate: true })

async function loadModules() {
  if (!examStore.state.currentExamId) { modules.value = []; return }
  try {
    const res = await getExamModules(examStore.state.currentExamId)
    modules.value = res.data.data || []
  } catch (err) {
    console.error('Failed to load modules:', err)
  }
}

async function loadQuestions() {
  loading.value = true
  try {
    const params = { page: currentPage.value, size: pageSize }
    if (filter.value.moduleId) {
      params.module_id = filter.value.moduleId
    } else if (examStore.state.currentExamId) {
      params.exam_type_id = examStore.state.currentExamId
    }
    if (filter.value.type) params.type = filter.value.type
    if (filter.value.difficulty) params.difficulty = filter.value.difficulty
    const res = await getQuestions(params)
    questions.value = res.data.data || []
    total.value = res.data.total || 0
  } catch (err) {
    console.error('Failed to load questions:', err)
  } finally {
    loading.value = false
  }
}

function changePage(page) { currentPage.value = page; loadQuestions() }

function editQuestion(q) {
  editingQuestion.value = q
  form.value = {
    moduleId: q.module_id, type: q.type || 'single', content: q.content,
    options: q.options, answer: q.answer, analysis: q.analysis || '',
    difficulty: q.difficulty || 1, tags: q.tags || '', source: q.source || '',
  }
}

function closeModal() {
  showAddModal.value = false
  editingQuestion.value = null
  form.value = { moduleId: '', type: 'single', content: '', options: '', answer: '', analysis: '', difficulty: 1, tags: '', source: '' }
}

async function saveQuestion() {
  if (!form.value.moduleId || !form.value.content || !form.value.answer) {
    alert('请填写模块、题干和正确答案'); return
  }
  saving.value = true
  try {
    const data = {
      module_id: parseInt(form.value.moduleId), type: form.value.type,
      content: form.value.content, options: form.value.options,
      answer: form.value.answer, analysis: form.value.analysis,
      difficulty: parseInt(form.value.difficulty), tags: form.value.tags, source: form.value.source,
    }
    if (editingQuestion.value) await updateQuestion(editingQuestion.value.id, data)
    else await createQuestion(data)
    closeModal(); await loadQuestions()
  } catch (err) {
    alert('保存失败：' + (err.response?.data?.error || err.message))
  } finally { saving.value = false }
}

async function confirmDelete(q) {
  if (!confirm(`确定要删除第 ${q.id} 题吗？`)) return
  try { await deleteQuestion(q.id); await loadQuestions() }
  catch (err) { alert('删除失败') }
}

async function handleImport(event) {
  const file = event.target.files[0]
  if (!file) return
  try {
    const text = await file.text()
    const data = JSON.parse(text)
    if (!Array.isArray(data)) { alert('JSON 格式错误：需要是一个数组'); return }
    const count = data.length
    if (!confirm(`即将导入 ${count} 道题目，确定吗？`)) return
    try {
      const res = await importQuestions(data)
      const imported = res.data?.imported ?? count
      alert(`导入完成：成功导入 ${imported} 道题目`)
    } catch (err) {
      alert('导入失败：' + (err.response?.data?.error || err.message))
    }
    await loadQuestions()
  } catch (err) { alert('导入失败：文件格式错误') }
  event.target.value = ''
}

function getTypeLabel(type) {
  return { single: '单选', multi: '多选', judge: '判断', fill: '填空' }[type] || type
}

function truncate(str, len) {
  if (!str) return ''
  return str.length > len ? str.substring(0, len) + '...' : str
}
</script>

<style scoped>
.question-manage {
  max-width: 1000px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.25rem;
}

h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text);
}

.header-actions {
  display: flex;
  gap: 0.5rem;
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

.btn svg {
  width: 16px;
  height: 16px;
}

.btn-primary {
  background: var(--primary);
  color: white;
}

.btn-primary:hover { background: var(--primary-dark); }

.btn-ghost {
  background: var(--bg-card);
  color: var(--text-secondary);
  border: 1px solid var(--border);
}

.btn-ghost:hover { background: var(--bg-hover); border-color: var(--text-muted); }

.import-btn { cursor: pointer; }

.filter-bar {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1.25rem;
  flex-wrap: wrap;
}

.filter-select {
  padding: 0.45rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--bg-card);
  color: var(--text);
  font-size: 0.85rem;
  cursor: pointer;
  transition: var(--transition);
}

.filter-select:focus {
  border-color: var(--primary);
  outline: none;
}

.table-container {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  background: var(--bg-hover);
  padding: 0.65rem 0.85rem;
  text-align: left;
  font-weight: 600;
  font-size: 0.78rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  border-bottom: 1px solid var(--border);
}

.data-table td {
  padding: 0.65rem 0.85rem;
  border-bottom: 1px solid var(--border-light);
  font-size: 0.85rem;
  color: var(--text);
}

.data-table tbody tr:last-child td {
  border-bottom: none;
}

.row-alt {
  background: #fafbfc;
}

.cell-id {
  color: var(--text-muted);
  font-size: 0.8rem;
}

.type-badge {
  background: var(--primary-bg);
  color: var(--primary);
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
}

.cell-content {
  max-width: 280px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.cell-answer {
  color: var(--primary);
  font-weight: 600;
}

.difficulty-stars {
  font-size: 0.7rem;
}

.star {
  color: var(--border);
}

.star.filled {
  color: var(--warning);
}

.cell-actions {
  white-space: nowrap;
}

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.3rem;
  border-radius: var(--radius-sm);
  transition: var(--transition);
  color: var(--text-muted);
  display: inline-flex;
  align-items: center;
}

.btn-icon svg {
  width: 16px;
  height: 16px;
}

.btn-icon:hover {
  background: var(--bg-hover);
  color: var(--primary);
}

.btn-icon.btn-danger:hover {
  background: var(--error-bg);
  color: var(--error);
}

.empty-row {
  text-align: center;
  color: var(--text-muted);
  padding: 2.5rem !important;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 0.75rem;
  margin-top: 1.25rem;
}

.btn-page {
  padding: 0.45rem 0.85rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--bg-card);
  color: var(--text-secondary);
  font-size: 0.85rem;
  cursor: pointer;
  transition: var(--transition);
}

.btn-page:hover:not(:disabled) {
  border-color: var(--primary);
  color: var(--primary);
}

.btn-page:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-info {
  font-size: 0.85rem;
  color: var(--text-muted);
}

/* Modal */
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

/* Mobile responsive */
@media (max-width: 768px) {
  .question-manage {
    max-width: 100%;
  }

  h1 {
    font-size: 1.25rem;
  }

  .page-header {
    flex-direction: column;
    align-items: stretch;
    gap: 0.75rem;
  }

  .header-actions {
    width: 100%;
  }

  .header-actions .btn {
    flex: 1;
    justify-content: center;
  }

  .filter-bar {
    flex-direction: column;
    gap: 0.5rem;
  }

  .filter-select {
    width: 100%;
  }

  /* Make table horizontally scrollable */
  .table-container {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .data-table {
    min-width: 560px;
  }

  .cell-content {
    max-width: 180px;
  }

  .pagination {
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .btn-page {
    padding: 0.4rem 0.7rem;
    font-size: 0.8rem;
  }

  .page-info {
    font-size: 0.8rem;
  }

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
