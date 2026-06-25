<template>
  <div class="score-ring">
    <svg viewBox="0 0 120 120">
      <circle class="ring-bg" cx="60" cy="60" r="50"></circle>
      <circle
        class="ring-fill"
        cx="60" cy="60" r="50"
        :style="{ strokeDasharray: 314, strokeDashoffset: 314 - (314 * percentage / 100) }"
      ></circle>
    </svg>
    <div class="ring-value">
      <span class="ring-number">{{ percentage }}</span>
      <span class="ring-unit">%</span>
    </div>
  </div>
</template>

<script setup>
defineProps({
  percentage: { type: Number, required: true },
})
</script>

<style scoped>
.score-ring {
  position: relative;
  width: 120px;
  height: 120px;
  margin: 0 auto;
}

.score-ring svg {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.ring-bg {
  fill: none;
  stroke: var(--border);
  stroke-width: 8;
}

.ring-fill {
  fill: none;
  stroke: var(--success);
  stroke-width: 8;
  stroke-linecap: round;
  transition: stroke-dashoffset 0.8s ease;
}

.ring-value {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ring-number {
  font-size: 1.8rem;
  font-weight: 700;
  color: var(--text);
}

.ring-unit {
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-left: 1px;
}
</style>
