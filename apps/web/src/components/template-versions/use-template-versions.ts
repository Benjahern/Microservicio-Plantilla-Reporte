import { computed, ref, watch } from "vue"
import { errorMessage } from "@/shared/http-client.api"
import type { ApiContext, Pagination, Version } from "@/types/template"
import { getVersion, listVersions } from "./template-versions.api"

export function useTemplateVersions(context: () => ApiContext, templateId: () => string, limit = 10) {
  const versions = ref<Version[]>([])
  const pagination = ref<Pagination>({ page: 1, limit, total: 0 })
  const selected = ref<Version | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const totalPages = computed(() =>
    Math.max(1, Math.ceil(pagination.value.total / pagination.value.limit)),
  )

  async function load(page = 1) {
    loading.value = true
    error.value = null
    try {
      const result = await listVersions(context(), templateId(), page, limit)
      versions.value = result.data
      pagination.value = result.pagination
    } catch (e) {
      error.value = errorMessage(e)
    } finally {
      loading.value = false
    }
  }

  async function select(number: number) {
    error.value = null
    try {
      selected.value = await getVersion(context(), templateId(), number)
    } catch (e) {
      error.value = errorMessage(e)
    }
  }

  watch(
    templateId,
    () => {
      selected.value = null
      void load(1)
    },
    { immediate: true },
  )

  return { versions, pagination, totalPages, selected, loading, error, load, select }
}
