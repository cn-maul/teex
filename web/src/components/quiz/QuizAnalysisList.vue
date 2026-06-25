<template>
  <div class="analysis-list">
    <h3 class="analysis-title">题目解析</h3>
    <div
      v-for="(q, idx) in questions"
      :key="q.id"
      class="analysis-item"
      :class="getItemClass(idx)"
    >
      <div class="analysis-header">
        <span class="analysis-number">{{ idx + 1 }}</span>
        <span class="analysis-type">{{ getTypeLabel(q.type) }}</span>
        <span class="analysis-status" v-if="isExamMode && !results[idx]">未作答</span>
        <span class="analysis-status status-correct" v-else-if="results[idx]?.is_correct">✓ 正确</span>
        <span class="analysis-status status-wrong" v-else-if="results[idx]">✗ 错误</span>
      </div>
      <div class="analysis-content">{{ q.content }}</div>
      <div class="analysis-detail">
        <span class="analysis-user-answer" v-if="results[idx]">你的答案：<strong>{{ results[idx].user_input || '（未作答）' }}</strong></span>
        <span class="analysis-correct-answer">正确答案：<strong>{{ q.answer }}</strong></span>
      </div>
      <div class="analysis-explanation" v-if="q.analysis">{{ q.analysis }}</div>
      <div class="analysis-options">
        <div
          v-for="(option, oi) in parseOptions(q.options)"
          :key="oi"
          class="analysis-option"
          :class="getOptionClass(idx, option, q)"
        >
          <span class="aopt-letter">{{ getOptionLetter(option) }}</span>
          <span class="aopt-text">{{ getOptionText(option) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { getTypeLabel, parseOptions, getOptionLetter, getOptionText } from '../../utils/quiz'

const props = defineProps({
  questions: { type: Array, required: true },
  results: { type: Array, required: true },
  isExamMode: { type: Boolean, default: false },
})

function getItemClass(idx) {
  const r = props.results[idx]
  if (!r && props.isExamMode) return 'analysis-unanswered'
  if (r?.is_correct) return 'analysis-correct'
  if (r) return 'analysis-wrong'
  return ''
}

function getOptionClass(idx, option, question) {
  const letter = getOptionLetter(option)
  const classes = []

  if (props.isExamMode) {
    // 考试模式：根据用户选择和正确答案判断
    const userAnswer = props.results[idx]?.user_input || ''
    const isSelected = userAnswer.split(',').map(s => s.trim()).includes(letter)
    const isCorrect = (question.answer || '').split(',').map(s => s.trim()).includes(letter)
    if (isSelected) classes.push('aopt-selected')
    if (isCorrect) classes.push('aopt-correct')
    if (isSelected && !isCorrect) classes.push('aopt-wrong')
  } else {
    // 解析模式
    const userAnswer = props.results[idx]?.user_input || ''
    const correctLetters = new Set((question.answer || '').split(',').map(s => s.trim()))
    if (userAnswer.split(',').map(s => s.trim()).includes(letter)) classes.push('aopt-selected')
    if (correctLetters.has(letter)) classes.push('aopt-correct')
    if (userAnswer.split(',').map(s => s.trim()).includes(letter) && !correctLetters.has(letter)) classes.push('aopt-wrong')
  }

  return classes.join(' ')
}
</script>

<style scoped>
.analysis-list {
  margin-top: 1.5rem;
}

.analysis-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 1rem;
}

.analysis-item {
  background: var(--card-bg);
  border-radius: 12px;
  padding: 1rem;
  margin-bottom: 0.75rem;
  border-left: 3px solid var(--border);
}

.analysis-item.analysis-correct { border-left-color: var(--success); }
.analysis-item.analysis-wrong { border-left-color: var(--error); }
.analysis-item.analysis-unanswered { border-left-color: var(--text-muted); }

.analysis-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.analysis-number {
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--text-muted);
  min-width: 1.5rem;
}

.analysis-type {
  font-size: 0.7rem;
  background: var(--bg-secondary);
  padding: 0.1rem 0.4rem;
  border-radius: 10px;
  color: var(--text-muted);
}

.analysis-status {
  font-size: 0.75rem;
  margin-left: auto;
  font-weight: 600;
}

.status-correct { color: var(--success); }
.status-wrong { color: var(--error); }

.analysis-content {
  font-size: 0.9rem;
  color: var(--text);
  margin-bottom: 0.5rem;
  line-height: 1.5;
}

.analysis-detail {
  display: flex;
  gap: 1rem;
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-bottom: 0.5rem;
  flex-wrap: wrap;
}

.analysis-explanation {
  font-size: 0.8rem;
  color: var(--text-muted);
  background: var(--bg-secondary);
  padding: 0.5rem 0.75rem;
  border-radius: 8px;
  margin-top: 0.5rem;
  line-height: 1.5;
}

.analysis-options {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  margin-top: 0.5rem;
}

.analysis-option {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.35rem 0.6rem;
  border-radius: 8px;
  font-size: 0.8rem;
  color: var(--text-muted);
  border: 1px solid transparent;
}

.aopt-selected { background: rgba(99, 102, 241, 0.08); border-color: var(--primary); color: var(--text); }
.aopt-correct { background: rgba(16, 185, 129, 0.08); border-color: var(--success); color: var(--success); }
.aopt-wrong { background: rgba(239, 68, 68, 0.08); border-color: var(--error); color: var(--error); }

.aopt-letter {
  font-weight: 700;
  min-width: 1.2rem;
}
</style>
