<template>
  <div class="exam-manage">
    <div class="page-header">
      <h1>考试管理</h1>
      <button class="btn btn-primary" @click="openExamModal()">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
        新增考试
      </button>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>

    <!-- 空状态 -->
    <div v-else-if="exams.length === 0" class="empty">
      <p>暂无考试数据，请先添加考试类型</p>
    </div>

    <!-- 考试列表 -->
    <div v-else class="exam-list">
      <div v-for="exam in exams" :key="exam.id" class="exam-card">
        <!-- 考试头部 -->
        <div class="exam-header" @click="toggleExam(exam.id)">
          <div class="exam-header-left">
            <svg class="chevron" :class="{ open: expandedExams[exam.id] }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="6 9 12 15 18 9"></polyline></svg>
            <div>
              <h3>{{ exam.name }}</h3>
              <p class="exam-remark" v-if="exam.remark">{{ exam.remark }}</p>
            </div>
          </div>
          <div class="exam-header-right" @click.stop>
            <span class="module-count">{{ (exam.modules || []).length }} 个科目</span>
            <button class="btn-icon" @click="openExamModal(exam)" title="编辑">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
            </button>
            <button class="btn-icon btn-danger" @click="confirmDeleteExam(exam)" title="删除">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
            </button>
          </div>
        </div>

        <!-- 科目列表 -->
        <Transition name="slide">
          <div v-if="expandedExams[exam.id]" class="exam-modules">
            <div class="modules-header">
              <span class="modules-title">科目列表</span>
              <button class="btn btn-sm btn-primary" @click="openModuleModal(exam.id)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
                添加科目
              </button>
            </div>
            <div class="table-container">
              <table class="data-table">
                <thead>
                  <tr>
                    <th style="width: 56px">ID</th>
                    <th>科目名称</th>
                    <th style="width: 72px">排序</th>
                    <th style="width: 80px">题目数</th>
                    <th style="width: 96px">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(mod, idx) in examModules[exam.id] || []" :key="mod.id" :class="{ 'row-alt': idx % 2 === 1 }">
                    <td class="cell-id">{{ mod.id }}</td>
                    <td>{{ mod.name }}</td>
                    <td>{{ mod.sort }}</td>
                    <td><strong>{{ mod.question_count }}</strong></td>
                    <td class="cell-actions">
                      <button class="btn-icon" @click="openModuleModal(exam.id, mod)" title="编辑">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
                      </button>
                      <button class="btn-icon btn-danger" @click="confirmDeleteModule(mod)" title="删除">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                      </button>
                    </td>
                  </tr>
                  <tr v-if="(examModules[exam.id] || []).length === 0">
                    <td colspan="5" class="empty-row">暂无科目，请添加</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </Transition>
      </div>
    </div>

    <!-- 新增/编辑考试弹窗 -->
    <Transition name="modal">
      <div class="modal-overlay" v-if="showExamModal" @click.self="closeExamModal">
        <div class="modal">
          <div class="modal-header">
            <h2>{{ editingExam ? '编辑考试' : '新增考试' }}</h2>
            <button class="modal-close" @click="closeExamModal">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
            </button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label>考试名称</label>
              <input v-model="examForm.name" class="form-input" placeholder="如：国家公务员" />
            </div>
            <div class="form-group">
              <label>备注</label>
              <textarea v-model="examForm.remark" class="form-input textarea" rows="2" placeholder="可选"></textarea>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="closeExamModal">取消</button>
            <button class="btn btn-primary" @click="saveExam" :disabled="saving">
              {{ saving ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- 新增/编辑科目弹窗 -->
    <Transition name="modal">
      <div class="modal-overlay" v-if="showModuleModal" @click.self="closeModuleModal">
        <div class="modal">
          <div class="modal-header">
            <h2>{{ editingModule ? '编辑科目' : '新增科目' }}</h2>
            <button class="modal-close" @click="closeModuleModal">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
            </button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label>科目名称</label>
              <input v-model="moduleForm.name" class="form-input" placeholder="如：行测-言语理解" />
            </div>
            <div class="form-group">
              <label>排序（数字越小越靠前）</label>
              <input v-model.number="moduleForm.sort" type="number" class="form-input" placeholder="0" />
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="closeModuleModal">取消</button>
            <button class="btn btn-primary" @click="saveModule" :disabled="saving">
              {{ saving ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import {
  getExamTypes, getExamModules,
  createExamType, updateExamType, deleteExamType,
  createModule, updateModule, deleteModule,
  showToast,
} from '../api'
import { useExamStore } from '../stores/exam'

const examStore = useExamStore()

const exams = ref([])
const loading = ref(false)
const saving = ref(false)
const expandedExams = reactive({})
const examModules = reactive({})

// 考试弹窗
const showExamModal = ref(false)
const editingExam = ref(null)
const examForm = ref({ name: '', remark: '' })

// 科目弹窗
const showModuleModal = ref(false)
const editingModule = ref(null)
const moduleForm = ref({ name: '', sort: 0, examTypeId: null })

onMounted(() => loadExams())

async function loadExams() {
  loading.value = true
  try {
    const res = await getExamTypes()
    exams.value = res.data.data || []
  } catch (err) {
    console.error('Failed to load exams:', err)
  } finally {
    loading.value = false
  }
}

async function toggleExam(examId) {
  expandedExams[examId] = !expandedExams[examId]
  if (expandedExams[examId] && !examModules[examId]) {
    await loadModules(examId)
  }
}

async function loadModules(examId) {
  try {
    const res = await getExamModules(examId)
    examModules[examId] = res.data.data || []
  } catch (err) {
    console.error('Failed to load modules:', err)
  }
}

// --- 考试 CRUD ---

function openExamModal(exam = null) {
  editingExam.value = exam
  examForm.value = exam
    ? { name: exam.name, remark: exam.remark || '' }
    : { name: '', remark: '' }
  showExamModal.value = true
}

function closeExamModal() {
  showExamModal.value = false
  editingExam.value = null
}

async function saveExam() {
  if (!examForm.value.name.trim()) {
    showToast('请填写考试名称', 'error'); return
  }
  saving.value = true
  try {
    const data = { name: examForm.value.name.trim(), remark: examForm.value.remark.trim() }
    if (editingExam.value) {
      await updateExamType(editingExam.value.id, data)
    } else {
      await createExamType(data)
    }
    closeExamModal()
    await loadExams()
    await examStore.refreshExams()
  } catch (err) {
    showToast('保存失败：' + (err.response?.data?.error || err.message), 'error')
  } finally {
    saving.value = false
  }
}

async function confirmDeleteExam(exam) {
  if (!confirm(`确定要删除考试「${exam.name}」吗？其下所有科目也会被删除。`)) return
  try {
    await deleteExamType(exam.id)
    delete examModules[exam.id]
    delete expandedExams[exam.id]
    await loadExams()
    await examStore.refreshExams()
  } catch (err) {
    showToast('删除失败：' + (err.response?.data?.error || err.message), 'error')
  }
}

// --- 科目 CRUD ---

function openModuleModal(examTypeId, mod = null) {
  editingModule.value = mod
  moduleForm.value = mod
    ? { name: mod.name, sort: mod.sort || 0, examTypeId }
    : { name: '', sort: 0, examTypeId }
  showModuleModal.value = true
}

function closeModuleModal() {
  showModuleModal.value = false
  editingModule.value = null
}

async function saveModule() {
  if (!moduleForm.value.name.trim()) {
    showToast('请填写科目名称', 'error'); return
  }
  saving.value = true
  try {
    const data = {
      name: moduleForm.value.name.trim(),
      sort: moduleForm.value.sort || 0,
      exam_type_id: moduleForm.value.examTypeId,
    }
    if (editingModule.value) {
      await updateModule(editingModule.value.id, data)
    } else {
      await createModule(data)
    }
    closeModuleModal()
    await loadModules(moduleForm.value.examTypeId)
    await loadExams()
    await examStore.refreshExams()
  } catch (err) {
    showToast('保存失败：' + (err.response?.data?.error || err.message), 'error')
  } finally {
    saving.value = false
  }
}

async function confirmDeleteModule(mod) {
  if (!confirm(`确定要删除科目「${mod.name}」吗？`)) return
  try {
    await deleteModule(mod.id)
    await loadModules(mod.exam_type_id)
    await loadExams()
    await examStore.refreshExams()
  } catch (err) {
    showToast('删除失败：' + (err.response?.data?.error || err.message), 'error')
  }
}
</script>

<style scoped>
.exam-manage {
  max-width: 900px;
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

/* Buttons */
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

.btn svg { width: 16px; height: 16px; }

.btn-primary { background: var(--primary); color: white; }
.btn-primary:hover { background: var(--primary-dark); }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-ghost { background: var(--bg-card); color: var(--text-secondary); border: 1px solid var(--border); }
.btn-ghost:hover { background: var(--bg-hover); border-color: var(--text-muted); }

.btn-sm { padding: 0.35rem 0.7rem; font-size: 0.8rem; }
.btn-sm svg { width: 14px; height: 14px; }

.btn-icon { background: none; border: none; cursor: pointer; padding: 0.3rem; border-radius: var(--radius-sm); transition: var(--transition); color: var(--text-muted); display: inline-flex; }
.btn-icon svg { width: 16px; height: 16px; }
.btn-icon:hover { background: var(--bg-hover); color: var(--primary); }
.btn-icon.btn-danger:hover { background: var(--error-bg); color: var(--error); }

/* Loading / Empty */
.loading { text-align: center; padding: 3rem; }
.spinner { width: 36px; height: 36px; border: 3px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 0.8s linear infinite; margin: 0 auto; }
@keyframes spin { to { transform: rotate(360deg); } }
.empty { text-align: center; padding: 4rem 2rem; color: var(--text-muted); }

/* Exam card */
.exam-list { display: flex; flex-direction: column; gap: 0.75rem; }

.exam-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.exam-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  cursor: pointer;
  transition: var(--transition);
}

.exam-header:hover { background: var(--bg-hover); }

.exam-header-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 0;
}

.chevron {
  width: 18px;
  height: 18px;
  color: var(--text-muted);
  transition: transform 0.2s ease;
  flex-shrink: 0;
}

.chevron.open { transform: rotate(180deg); }

.exam-header-left h3 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
  margin: 0;
}

.exam-remark {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin: 0.15rem 0 0;
}

.exam-header-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
}

.module-count {
  font-size: 0.8rem;
  color: var(--text-muted);
  background: var(--bg-hover);
  padding: 0.2rem 0.6rem;
  border-radius: 12px;
}

/* Modules section */
.exam-modules {
  border-top: 1px solid var(--border);
  padding: 1rem 1.25rem;
}

.modules-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.modules-title {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-secondary);
}

/* Table */
.table-container {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
}

.data-table { width: 100%; border-collapse: collapse; }

.data-table th {
  background: var(--bg-hover);
  padding: 0.55rem 0.75rem;
  text-align: left;
  font-weight: 600;
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  border-bottom: 1px solid var(--border);
}

.data-table td {
  padding: 0.55rem 0.75rem;
  border-bottom: 1px solid var(--border-light);
  font-size: 0.85rem;
  color: var(--text);
}

.data-table tbody tr:last-child td { border-bottom: none; }
.row-alt { background: rgba(0,0,0,0.015); }
.cell-id { color: var(--text-muted); font-size: 0.8rem; }
.cell-actions { white-space: nowrap; }

.empty-row { text-align: center; color: var(--text-muted); padding: 2rem !important; }

/* Slide transition */
.slide-enter-active, .slide-leave-active { transition: all 0.2s ease; overflow: hidden; }
.slide-enter-from, .slide-leave-to { opacity: 0; max-height: 0; }
.slide-enter-to, .slide-leave-from { opacity: 1; max-height: 500px; }

/* Modal */
.modal-overlay {
  position: fixed; inset: 0; background: rgba(15,23,42,0.4); backdrop-filter: blur(4px);
  display: flex; align-items: center; justify-content: center; z-index: 1000;
}

.modal {
  background: var(--bg-card); border-radius: var(--radius-xl);
  width: 92%; max-width: 440px; max-height: 85vh; overflow-y: auto; box-shadow: var(--shadow-lg);
}

.modal-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 1.25rem 1.5rem; border-bottom: 1px solid var(--border);
}

.modal-header h2 { font-size: 1.1rem; font-weight: 600; color: var(--text); margin: 0; }

.modal-close { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 0.25rem; border-radius: var(--radius-sm); transition: var(--transition); display: flex; }
.modal-close svg { width: 18px; height: 18px; }
.modal-close:hover { background: var(--bg-hover); color: var(--text); }

.modal-body { padding: 1.25rem 1.5rem; }

.form-group { margin-bottom: 1rem; }
.form-group label { display: block; font-size: 0.85rem; font-weight: 500; color: var(--text-secondary); margin-bottom: 0.3rem; }

.form-input {
  width: 100%; padding: 0.55rem 0.75rem; border: 1px solid var(--border);
  border-radius: var(--radius); font-size: 0.875rem; color: var(--text);
  background: var(--bg-card); outline: none; transition: var(--transition);
}

.form-input:focus { border-color: var(--primary); box-shadow: 0 0 0 3px var(--primary-bg); }
.form-input.textarea { resize: vertical; min-height: 60px; }

.modal-footer {
  display: flex; justify-content: flex-end; gap: 0.5rem;
  padding: 1rem 1.5rem; border-top: 1px solid var(--border);
}

.modal-enter-active, .modal-leave-active { transition: all 0.2s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.modal-enter-from .modal, .modal-leave-to .modal { transform: scale(0.95) translateY(8px); }

/* Mobile */
@media (max-width: 768px) {
  .exam-manage { max-width: 100%; }
  h1 { font-size: 1.25rem; }
  .page-header { flex-direction: column; align-items: stretch; gap: 0.75rem; }
  .exam-header { padding: 0.85rem 1rem; }
  .exam-modules { padding: 0.75rem 1rem; }
  .table-container { overflow-x: auto; }
  .data-table { min-width: 400px; }
}
</style>
