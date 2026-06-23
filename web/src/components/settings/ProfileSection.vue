<template>
  <div class="settings-section">
    <div class="section-header">
      <h2>账户信息</h2>
      <p class="section-desc">管理你的个人资料</p>
    </div>

    <div class="profile-card">
      <div class="profile-avatar">👤</div>
      <div class="profile-info">
        <div class="profile-name">
          {{ user?.nickname || user?.username }}
          <span class="role-badge" :class="user?.role === 'admin' ? 'role-admin' : 'role-user'">
            {{ user?.role === 'admin' ? '管理员' : '普通用户' }}
          </span>
        </div>
        <div class="profile-meta">
          <span>用户名：{{ user?.username }}</span>
          <span v-if="user?.created_at">注册时间：{{ formatDate(user.created_at) }}</span>
        </div>
      </div>
    </div>

    <!-- 修改昵称 -->
    <div class="setting-item">
      <div class="setting-info">
        <label class="setting-label">昵称</label>
        <p class="setting-desc">当前昵称：{{ user?.nickname || user?.username }}</p>
      </div>
      <div class="setting-control">
        <div v-if="!editingNickname" class="inline-edit">
          <button class="btn btn-ghost" @click="$emit('start-edit-nickname')">修改</button>
        </div>
        <div v-else class="inline-edit-form">
          <input :value="newNickname" @input="$emit('update:newNickname', $event.target.value)" type="text" placeholder="输入新昵称" maxlength="50" class="edit-input" />
          <button class="btn btn-primary btn-sm" @click="$emit('save-nickname')" :disabled="!newNickname.trim()">保存</button>
          <button class="btn btn-ghost btn-sm" @click="$emit('cancel-edit-nickname')">取消</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  user: { type: Object, default: null },
  editingNickname: { type: Boolean, default: false },
  newNickname: { type: String, default: '' },
  formatDate: { type: Function, required: true },
})

defineEmits(['start-edit-nickname', 'update:newNickname', 'save-nickname', 'cancel-edit-nickname'])
</script>

<style scoped>
.settings-section {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  padding: 1.5rem;
  margin-bottom: 1rem;
}

.section-header {
  margin-bottom: 1.25rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--border-light);
}

.section-header h2 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 0.15rem;
}

.section-desc {
  font-size: 0.85rem;
  color: var(--text-muted);
}

.profile-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding-bottom: 1rem;
  margin-bottom: 0.5rem;
  border-bottom: 1px solid var(--border-light);
}

.profile-avatar {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  background: var(--bg-hover);
  border-radius: 50%;
  flex-shrink: 0;
}

.profile-info {
  flex: 1;
  min-width: 0;
}

.profile-name {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.role-badge {
  font-size: 0.7rem;
  font-weight: 600;
  padding: 0.1rem 0.5rem;
  border-radius: 10px;
}

.role-admin {
  background: #fef3c7;
  color: #92400e;
}

.role-user {
  background: var(--primary-bg);
  color: var(--primary);
}

.profile-meta {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-top: 0.2rem;
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
  padding: 0.85rem 0;
}

.setting-item + .setting-item {
  border-top: 1px solid var(--border-light);
}

.setting-info {
  flex: 1;
  min-width: 0;
}

.setting-label {
  display: block;
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--text);
  margin-bottom: 0.1rem;
}

.setting-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.setting-control {
  flex-shrink: 0;
}

.inline-edit-form {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.edit-input {
  padding: 0.4rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  font-size: 0.85rem;
  width: 180px;
  outline: none;
  transition: border-color 0.2s;
}

.edit-input:focus {
  border-color: var(--primary);
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

.btn-sm {
  padding: 0.35rem 0.75rem;
  font-size: 0.8rem;
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

@media (max-width: 768px) {
  .setting-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }

  .profile-meta {
    flex-direction: column;
    gap: 0.2rem;
  }
}
</style>
