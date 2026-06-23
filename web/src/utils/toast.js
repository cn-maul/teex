export function showToast(message, type = 'success') {
  // Remove any existing toast
  const existing = document.querySelector('.vue-toast-notification')
  if (existing) existing.remove()

  // Inject animation styles once
  if (!document.querySelector('#vue-toast-style')) {
    const style = document.createElement('style')
    style.id = 'vue-toast-style'
    style.textContent = `
      @keyframes vue-toast-in { from { opacity: 0; transform: translateX(-50%) translateY(20px); } to { opacity: 1; transform: translateX(-50%) translateY(0); } }
      @keyframes vue-toast-out { from { opacity: 1; transform: translateX(-50%) translateY(0); } to { opacity: 0; transform: translateX(-50%) translateY(20px); } }
    `
    document.head.appendChild(style)
  }

  const toast = document.createElement('div')
  toast.className = 'vue-toast-notification'
  toast.style.cssText = `
    position: fixed; bottom: 24px; left: 50%; transform: translateX(-50%);
    padding: 0.65rem 1.25rem; border-radius: 12px;
    font-size: 0.85rem; font-weight: 500; color: white; z-index: 10000;
    box-shadow: 0 4px 12px rgba(0,0,0,0.15);
    animation: vue-toast-in 0.3s ease;
    background: ${type === 'success' ? 'var(--success, #22c55e)' : type === 'error' ? 'var(--error, #ef4444)' : 'var(--primary, #3b82f6)'};
  `
  toast.textContent = message
  document.body.appendChild(toast)

  setTimeout(() => {
    toast.style.animation = 'vue-toast-out 0.3s ease forwards'
    setTimeout(() => toast.remove(), 300)
  }, 3000)
}
