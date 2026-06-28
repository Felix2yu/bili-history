export const useBiliApi = () => {
  const config = useRuntimeConfig()
  const { baseUrl } = useApiBase()

  const getBaseURL = () => {
    if (import.meta.server) {
      return config.backendUrl || config.public.defaultBackendUrl
    }
    return baseUrl.value
  }

  const apiFetch = $fetch.create({
    baseURL: getBaseURL(),
  })

  const getBiliHistory2024 = (
    page: number,
    size: number,
    sortOrder: number,
    tagName: string = '',
    mainCategory: string = '',
    dateRange: string = '',
    useLocalImages: boolean = false,
    business: string = ''
  ) => {
    return apiFetch('/history/all', {
      method: 'GET',
      params: {
        page,
        size,
        sort_order: sortOrder,
        tag_name: tagName,
        main_category: mainCategory,
        date_range: dateRange,
        use_local_images: useLocalImages,
        business,
      },
    })
  }

  const searchBiliHistory2024 = (
    search: string,
    searchType: string = 'all',
    page: number = 1,
    size: number = 30,
    useLocalImages: boolean = false,
    useSessdata: boolean = true
  ) => {
    return apiFetch('/history/search', {
      method: 'GET',
      params: {
        page,
        size,
        search,
        search_type: searchType,
        use_local_images: useLocalImages,
        use_sessdata: useSessdata,
      },
    })
  }

  const getAvailableYears = () => {
    return apiFetch('/history/available-years', {
      method: 'GET',
    })
  }

  const getVideoCategories = () => {
    return apiFetch('/categories/categories', {
      method: 'GET',
    })
  }

  const getMainCategories = () => {
    return apiFetch('/categories/main-categories', {
      method: 'GET',
    })
  }

  const checkServerHealth = () => {
    return apiFetch('/health', {
      method: 'GET',
    })
  }

  return {
    apiFetch,
    getBiliHistory2024,
    searchBiliHistory2024,
    getAvailableYears,
    getVideoCategories,
    getMainCategories,
    checkServerHealth,
  }
}
