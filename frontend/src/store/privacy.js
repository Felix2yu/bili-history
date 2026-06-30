import { ref } from 'vue'

// 从 localStorage 读取初始状态（仅客户端）
const isPrivacyMode = ref(typeof window !== 'undefined' ? localStorage.getItem('privacyMode') === 'true' : false)

export const usePrivacyStore = () => {
  const togglePrivacyMode = () => {
    isPrivacyMode.value = !isPrivacyMode.value
    // 保存到 localStorage
    localStorage.setItem('privacyMode', isPrivacyMode.value.toString())
  }
  
  const setPrivacyMode = (value) => {
    isPrivacyMode.value = value
    // 保存到 localStorage
    localStorage.setItem('privacyMode', value.toString())
  }

  return {
    isPrivacyMode,
    togglePrivacyMode,
    setPrivacyMode
  }
} 