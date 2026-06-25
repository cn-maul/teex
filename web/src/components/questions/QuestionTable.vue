<template>
  <div>
    <!-- 题目列表 -->
    <div class="table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th style="width: 40px">
              <input type="checkbox" class="row-checkbox" :checked="allSelected" @change="$emit('toggle-select-all')" />
            </th>
            <th style="width: 56px">ID</th>
            <th style="width: 80px">题型</th>
            <th>题干</th>
            <th style="width: 72px">答案</th>
            <th style="width: 80px">难度</th>
            <th style="width: 96px">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(q, idx) in questions" :key="q.id" :class="{ 'row-alt': idx % 2 === 1, 'row-selected': selectedIds.has(q.id) }">
            <td>
              <input type="checkbox" class="row-checkbox" :checked="selectedIds.has(q.id)" @change="$emit('toggle-select', q.id)" />
            </td>
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
              <button class="btn-icon" @click="$emit('edit', q)" title="编辑">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
              </button>
              <button class="btn-icon btn-icon-danger" @click="$emit('delete', q)" title="删除">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
              </button>
            </td>
          </tr>
          <tr v-if="questions.length === 0 && !loading">
            <td colspan="7" class="empty-row">暂无题目</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页 -->
    <div class="pagination" v-if="total > pageSize">
      <button class="btn-page" :disabled="currentPage <= 1" @click="$emit('change-page', currentPage - 1)">
        ← 上一页
      </button>
      <input
        type="number"
        :value="currentPage"
        @keyup.enter="goToPage($event.target.value)"
        min="1"
        :max="totalPages"
        class="page-input"
        placeholder="页码"
      />
      <span class="page-info">第 {{ currentPage }} / {{ totalPages }} 页（共 {{ total }} 题）</span>
      <button class="btn-page" :disabled="currentPage >= totalPages" @click="$emit('change-page', currentPage + 1)">
        下一页 →
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { getTypeLabel } from '../../utils/quiz'

const props = defineProps({
  questions: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
  total: { type: Number, default: 0 },
  currentPage: { type: Number, default: 1 },
  pageSize: { type: Number, default: 20 },
  selectedIds: { type: Object, default: () => new Set() },
  allSelected: { type: Boolean, default: false },
})

const emit = defineEmits(['change-page', 'toggle-select', 'toggle-select-all', 'edit', 'delete'])

const totalPages = computed(() => Math.ceil(props.total / props.pageSize))

function truncate(str, len) {
  if (!str) return ''
  return str.length > len ? str.substring(0, len) + '...' : str
}

function goToPage(val) {
  const p = parseInt(val, 10)
  if (isNaN(p) || p < 1 || p > totalPages.value) return
  emit('change-page', p)
}
</script>

<style scoped>
.table-container {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  overflow: hidden;
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

.row-alt {
  background: #fafbfc;
}

.type-badge {
  background: var(--primary-bg);
  color: var(--primary);
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  white-space: nowrap;
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

.row-checkbox {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: var(--primary);
}

.row-selected {
  background: #eff6ff !important;
}

.empty-row {
  text-align: center;
  color: var(--text-muted);
  padding: 2.5rem !important;
}

@media (max-width: 768px) {
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
}
</style>
