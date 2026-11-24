import { computed, ref, watch } from "vue"
import { errorMessage } from "@/shared/http-client.api"
import type { ApiContext, Pagination, Template } from "@/types/template"
import { deleteTemplate, listTemplates } from "./templates-list.api"

export function useTemplatesList(context: () => ApiContext, limit = 10) {
  const items = ref<Template[]>([])
  const pagination = ref<Pagination>({ page: 1, limit, total: 0 })
  const loading = ref(false)
  const error = ref<string | null>(null)

  const totalPages = computed(() =>
    Math.max(1, Math.ceil(pagination.value.total / pagination.value.limit)),
  )

  async function load(page = pagination.value.page) {
    const ctx = context()
    loading.value = true
    error.value = null
    try {
      const result = await listTemplates(ctx, page, limit)
      items.value = result.data
      pagination.value = result.pagination
    } catch (e) {
      items.value = []
      error.value = errorMessage(e)
    } finally {
      loading.value = false
    }
  }

  async function remove(id: string) {
    error.value = null
    try {
      await deleteTemplate(context(), id)
      // if it was the last item on the page, go back one page
      const page =
        items.value.length === 1 && pagination.value.page > 1
          ? pagination.value.page - 1
          : pagination.value.page
      await load(page)
      return true
    } catch (e) {
      error.value = errorMessage(e)
      return false
    }
  }

  watch(
    () => [context().baseUrl, context().token],
    () => load(1),
    { immediate: true },
  )

  return { items, pagination, totalPages, loading, error, load, remove }
}
