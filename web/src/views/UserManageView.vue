<template>
  <div class="user-manage">
    <div class="page-header">
      <h1>人员管理</h1>
      <p class="page-desc">管理系统中的所有用户账户</p>
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
          </tr>
        </tbody>
      </table>

      <div v-if="users.length === 0" class="empty">暂无用户数据</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getUsers } from '../api'

const users = ref([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await getUsers()
    users.value = res.data.data || []
  } catch (err) {
    console.error('Failed to load users:', err)
  } finally {
    loading.value = false
  }
})

function formatDate(dateStr) {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.user-manage {
  max-width: 900px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 1.5rem;
}

.page-header h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text);
  margin: 0 0 0.25rem 0;
}

.page-desc {
  font-size: 0.85rem;
  color: var(--text-muted);
  margin: 0;
}

.loading {
  text-align: center;
  padding: 3rem;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto;
}

@keyframes spin { to { transform: rotate(360deg); } }

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

@media (max-width: 768px) {
  .user-table-wrapper {
    overflow-x: auto;
  }

  .user-table {
    min-width: 500px;
  }
}
</style>
