<template>
  <div class="user-manage">
    <div class="page-header">
      <h1>人员管理</h1>
      <button class="btn btn-primary" @click="openCreateModal()">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
        新增用户
      </button>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>

    <div v-else class="user-table-wrapper">
      <table class="user-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>用户名</th>
            <th>昵称</th>
            <th>角色</th>
            <th>注册时间</th>
            <th style="width: 120px">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id">
            <td>{{ user.id }}</td>
            <td>{{ user.username }}</td>
            <td>{{ user.nickname || '-' }}</td>
            <td>
              <span class="role-tag" :class="user.role === 'admin' ? 'role-admin' : 'role-user'">
                {{ user.role === 'admin' ? '管理员' : '普通用户' }}
              </span>
            </td>
            <td>{{ formatDate(user.created_at) }}</td>
            <td class="cell-actions">
              <button class="btn-icon" @click="openEditModal(user)" title="编辑">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              </button>
              <button v-if="user.id !== authStore.user?.id" class="btn-icon btn-danger" @click="confirmDelete(user)" title="删除">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="users.length === 0" class="empty">暂无用户数据</div>
    </div>

    <!-- 新增/编辑弹窗 -->
    <Transition name="modal">
      <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
        <div class="modal">
          <div class="modal-header">
            <h2>{{ editingUser ? '编辑用户' : '新增用户' }}</h2>
            <button class="modal-close" @click="closeModal">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>
          <div class="modal-body">
            <div v-if="!editingUser" class="form-group">
              <label>用户名</label>
              <input v-model="form.username" class="form-input" placeholder="3-50 个字符" />
            </div>
            <div v-if="!editingUser" class="form-group">
              <label>密码</label>
              <input v-model="form.password" type="password" class="form-input" placeholder="至少 6 位" />
            </div>
            <div class="form-group">
              <label>昵称{{ editingUser ? '' : '（可选）' }}</label>
              <input v-model="form.nickname" class="form-input" placeholder="显示名称" />
            </div>
            <div class="form-group">
              <label>角色</label>
              <select v-model="form.role" class="form-input" :disabled="editingUser && editingUser.id === authStore.user?.id">
                <option value="user">普通用户</option>
                <option value="admin">管理员</option>
              </select>
            </div>
            <div v-if="editingUser" class="form-group">
              <label>重置密码（留空则不修改）</label>
              <input v-model="form.new_password" type="password" class="form-input" placeholder="至少 6 位" />
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-ghost" @click="closeModal">取消</button>
            <button class="btn btn-primary" @click="saveUser" :disabled="saving">
              {{ saving ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getUsers, adminCreateUser, adminUpdateUser, adminDeleteUser } from '../api'
import { showToast } from '../utils/toast'
import { formatDate } from '../utils/format'
import { useAuthStore } from '../stores/auth'
import { useConfirm } from '../utils/confirm'

const { showConfirm } = useConfirm()

const authStore = useAuthStore()
const users = ref([])
const loading = ref(true)
const saving = ref(false)
const showModal = ref(false)
const editingUser = ref(null)
const form = ref({ username: '', password: '', nickname: '', new_password: '', role: 'user' })

onMounted(async () => {
  await loadUsers()
})

async function loadUsers() {
  loading.value = true
  try {
    const res = await getUsers()
    users.value = res.data.data || []
  } catch (err) {
    console.error('Failed to load users:', err)
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  editingUser.value = null
  form.value = { username: '', password: '', nickname: '', new_password: '', role: 'user' }
  showModal.value = true
}

function openEditModal(user) {
  editingUser.value = user
  form.value = { username: user.username, password: '', nickname: user.nickname || '', new_password: '', role: user.role || 'user' }
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  editingUser.value = null
}

async function saveUser() {
  saving.value = true
  try {
    if (editingUser.value) {
      const data = {}
      if (form.value.nickname) data.nickname = form.value.nickname
      if (form.value.new_password) data.new_password = form.value.new_password
      if (form.value.role !== editingUser.value.role) data.role = form.value.role
      if (!data.nickname && !data.new_password && !data.role) {
        showToast('请提供需要修改的信息', 'error')
        saving.value = false
        return
      }
      await adminUpdateUser(editingUser.value.id, data)
      showToast('用户信息已更新')
    } else {
      if (!form.value.username || !form.value.password) {
        showToast('请填写用户名和密码', 'error')
        saving.value = false
        return
      }
      await adminCreateUser({
        username: form.value.username,
        password: form.value.password,
        nickname: form.value.nickname || undefined,
        role: form.value.role,
      })
      showToast('用户创建成功')
    }
    closeModal()
    await loadUsers()
  } catch (err) {
    showToast(err.response?.data?.error || '操作失败', 'error')
  } finally {
    saving.value = false
  }
}

async function confirmDelete(user) {
  if (!await showConfirm({ message: `确定要删除用户「${user.username}」吗？`, dangerMode: true })) return
  try {
    await adminDeleteUser(user.id)
    showToast('用户已删除')
    await loadUsers()
  } catch (err) {
    showToast(err.response?.data?.error || '删除失败', 'error')
  }
}

</script>

<style scoped>
.user-manage {
  max-width: 900px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.page-header h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text);
  margin: 0;
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

.btn svg {
  width: 16px;
  height: 16px;
}

.btn-primary { background: var(--primary); color: white; }

.btn-ghost { background: var(--bg-card); color: var(--text-secondary); border: 1px solid var(--border); }
.btn-ghost:hover:not(:disabled) { background: var(--bg-hover); border-color: var(--text-muted); }

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.3rem;
  border-radius: var(--radius-sm);
  transition: var(--transition);
  color: var(--text-muted);
  display: inline-flex;
}

.btn-icon svg { width: 16px; height: 16px; }
.btn-icon:hover { background: var(--bg-hover); color: var(--primary); }
.btn-icon.btn-danger:hover { background: var(--error-bg); color: var(--error); }

/* Loading / Empty */
.spinner { width: 32px; height: 32px; border: 3px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin 0.8s linear infinite; margin: 0 auto; }

/* Table */
.user-table-wrapper {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  overflow: hidden;
}

.user-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.user-table th {
  text-align: left;
  padding: 0.75rem 1rem;
  background: var(--bg);
  color: var(--text-muted);
  font-weight: 600;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  border-bottom: 1px solid var(--border);
}

.user-table td {
  padding: 0.7rem 1rem;
  color: var(--text);
  border-bottom: 1px solid var(--border-light);
}

.user-table tbody tr:last-child td {
  border-bottom: none;
}

.user-table tbody tr:hover {
  background: var(--bg-hover);
}

.cell-actions { white-space: nowrap; }

.role-tag {
  display: inline-block;
  padding: 0.15rem 0.55rem;
  border-radius: 10px;
  font-size: 0.75rem;
  font-weight: 600;
}

.role-admin {
  background: #fef3c7;
  color: #92400e;
}

.role-user {
  background: var(--primary-bg);
  color: var(--primary);
}

.empty {
  text-align: center;
  padding: 2rem;
  color: var(--text-muted);
  font-size: 0.85rem;
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
  max-width: 440px;
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

.modal-close svg { width: 18px; height: 18px; }
.modal-close:hover { background: var(--bg-hover); color: var(--text); }

.modal-body { padding: 1.25rem 1.5rem; }

.form-group { margin-bottom: 1rem; }
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
  box-sizing: border-box;
}

.form-input:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-bg);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border);
}

/* Modal transitions */
.modal-enter-active,
.modal-leave-active {
  transition: all 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal,
.modal-leave-to .modal {
  transform: scale(0.95) translateY(8px);
}

/* Mobile */
@media (max-width: 768px) {
  .user-manage { max-width: 100%; }
  .page-header { flex-direction: column; align-items: stretch; gap: 0.75rem; }
  .user-table-wrapper { overflow-x: auto; }
  .user-table { min-width: 600px; }
}
</style>
