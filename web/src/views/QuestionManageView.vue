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
        <button class="btn btn-ghost" @click="downloadTemplate">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
          下载模板
        </button>
        <button v-if="selectedIds.size > 0" class="btn btn-danger" @click="batchDelete">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
          删除选中（{{ selectedIds.size }}）
        </button>
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
    <QuestionTable
      :questions="questions"
      :loading="loading"
      :total="total"
      :current-page="currentPage"
      :page-size="pageSize"
      :selected-ids="selectedIds"
      :all-selected="allSelected"
      @change-page="changePage"
      @toggle-select="toggleSelect"
      @toggle-select-all="toggleSelectAll"
      @edit="editQuestion"
      @delete="confirmDelete"
    />

    <!-- 新增/编辑弹窗 -->
    <QuestionFormModal
      :visible="showAddModal || !!editingQuestion"
      :editing-question="editingQuestion"
      :modules="modules"
      :form="form"
      :saving="saving"
      @save="handleSaveQuestion"
      @close="closeModal"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import {
  getQuestions, createQuestion, updateQuestion, deleteQuestion,
  getExamModules, importQuestions, batchDeleteQuestions,
} from '../api'
import { showToast } from '../utils/toast'
import { useExamStore } from '../stores/exam'
import { useConfirm } from '../utils/confirm'
import QuestionTable from '../components/questions/QuestionTable.vue'
import QuestionFormModal from '../components/questions/QuestionFormModal.vue'

const { showConfirm } = useConfirm()

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
const selectedIds = ref(new Set())

const form = ref({
  moduleId: '', type: 'single', content: '', options: '', answer: '',
  analysis: '', difficulty: 1, tags: '', source: '',
})

const totalPages = computed(() => Math.ceil(total.value / pageSize))

const allSelected = computed(() => {
  return questions.value.length > 0 && questions.value.every(q => selectedIds.value.has(q.id))
})

function toggleSelect(id) {
  const set = new Set(selectedIds.value)
  if (set.has(id)) { set.delete(id) } else { set.add(id) }
  selectedIds.value = set
}

function toggleSelectAll() {
  if (allSelected.value) {
    selectedIds.value = new Set()
  } else {
    selectedIds.value = new Set(questions.value.map(q => q.id))
  }
}

function clearSelection() {
  selectedIds.value = new Set()
}

async function batchDelete() {
  const ids = [...selectedIds.value]
  if (ids.length === 0) return
  if (!await showConfirm({ message: `确定要删除选中的 ${ids.length} 道题目吗？`, dangerMode: true })) return
  try {
    const res = await batchDeleteQuestions(ids)
    const deleted = res.data?.data?.deleted ?? ids.length
    showToast(`成功删除 ${deleted} 道题目`, 'success')
  } catch (err) {
    showToast('删除失败：' + (err.response?.data?.error || err.message), 'error')
  }
  clearSelection()
  await loadQuestions()
}

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
  clearSelection()
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

function changePage(page) {
  const p = typeof page === 'number' ? page : parseInt(page, 10)
  if (isNaN(p) || p < 1 || p > totalPages.value) return
  currentPage.value = p
  loadQuestions()
}

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

async function handleSaveQuestion(data) {
  saving.value = true
  try {
    if (editingQuestion.value) await updateQuestion(editingQuestion.value.id, data)
    else await createQuestion(data)
    closeModal()
    await loadQuestions()
  } catch (err) {
    showToast('保存失败：' + (err.response?.data?.error || err.message), 'error')
  } finally { saving.value = false }
}

async function confirmDelete(q) {
  if (!await showConfirm({ message: `确定要删除第 ${q.id} 题吗？`, dangerMode: true })) return
  try { await deleteQuestion(q.id); await loadQuestions() }
  catch (err) { showToast('删除失败', 'error') }
}

const TEMPLATE_DATA = [
  {
    "module_id": "请查看考试管理页面的模块ID，填入数字",
    "type": "single",
    "content": "我国的根本政治制度是？",
    "options": "[\"A. 人民代表大会制度\",\"B. 中国共产党领导的多党合作制\",\"C. 民族区域自治制度\",\"D. 基层群众自治制度\"]",
    "answer": "A",
    "analysis": "人民代表大会制度是我国的根本政治制度。",
    "difficulty": 1,
    "tags": "政治,宪法",
    "source": "示例"
  },
  {
    "module_id": "请查看考试管理页面的模块ID，填入数字",
    "type": "multi",
    "content": "下列属于我国国家机构的有哪些？（多选）",
    "options": "[\"A. 全国人民代表大会\",\"B. 国务院\",\"C. 人民法院\",\"D. 人民检察院\"]",
    "answer": "A,B,C,D",
    "analysis": "以上四项均属于我国国家机构。",
    "difficulty": 2,
    "tags": "政治,宪法",
    "source": "示例"
  },
  {
    "module_id": "请查看考试管理页面的模块ID，填入数字",
    "type": "judge",
    "content": "我国一切权力属于人民。",
    "options": "[\"A. 正确\",\"B. 错误\"]",
    "answer": "A",
    "analysis": "《宪法》规定中华人民共和国的一切权力属于人民。",
    "difficulty": 1,
    "tags": "政治,宪法",
    "source": "示例"
  }
]

function downloadTemplate() {
  const blob = new Blob([JSON.stringify(TEMPLATE_DATA, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = '题目导入模板.json'
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

async function handleImport(event) {
  const file = event.target.files[0]
  if (!file) return
  try {
    const text = await file.text()
    const data = JSON.parse(text)
    if (!Array.isArray(data)) { showToast('JSON 格式错误：需要是一个数组', 'error'); return }

    // 前端校验字段
    const requiredFields = ['module_id', 'content', 'type', 'answer']
    const validTypes = new Set(['single', 'multiple', 'judge', 'fill', 'essay'])
    let invalidCount = 0
    for (const item of data) {
      if (!requiredFields.every(f => item[f] !== undefined && item[f] !== '')) { invalidCount++; continue }
      if (!validTypes.has(item.type)) { invalidCount++; continue }
    }
    if (invalidCount > 0 && invalidCount === data.length) {
      showToast(`全部 ${data.length} 道题目校验失败，请检查字段（module_id/content/type/answer）`, 'error')
      event.target.value = ''
      return
    }

    const count = data.length
    const warn = invalidCount > 0 ? `（${invalidCount} 道格式不正确将被跳过）` : ''
    if (!await showConfirm({ message: `即将导入 ${count} 道题目${warn}，确定吗？` })) {
      event.target.value = ''
      return
    }
    try {
      const res = await importQuestions(data)
      const imported = res.data?.data?.count ?? count
      showToast(`导入完成：成功导入 ${imported} 道题目`, 'success')
    } catch (err) {
      showToast('导入失败：' + (err.response?.data?.error || err.message), 'error')
    }
    await loadQuestions()
  } catch (err) { showToast('导入失败：文件格式错误', 'error') }
  event.target.value = ''
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

.btn-danger {
  background: var(--error);
  color: white;
  border: none;
}

.btn-danger:hover { background: #dc2626; }

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
}
</style>
