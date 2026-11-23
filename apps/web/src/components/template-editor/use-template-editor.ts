import { computed, reactive, ref, watch } from "vue"
import { ApiError, errorMessage } from "@/shared/http-client.api"
import type { ApiContext, Content, Template } from "@/types/template"
import { createTemplate, getTemplate, updateTemplate } from "./template-editor.api"

// Initial JSON shown to the user: "titulo" / "secciones" are keys of the report content schema.
const INITIAL_CONTENT = '{\n  "titulo": "Nuevo reporte",\n  "secciones": []\n}'

export function useTemplateEditor(context: () => ApiContext, templateId: () => string | undefined) {
  const form = reactive({ name: "", description: "", active: true, contentText: INITIAL_CONTENT })
  const currentVersion = ref<number | null>(null)
  const loading = ref(false)
  const saving = ref(false)
  const error = ref<string | null>(null)
  const contentError = ref<string | null>(null)

  const isEditing = computed(() => Boolean(templateId()))

  function apply(t: Template) {
    form.name = t.name
    form.description = t.description
    form.active = t.active
    form.contentText = JSON.stringify(t.content, null, 2)
    currentVersion.value = t.currentVersion
  }

  function reset() {
    form.name = ""
    form.description = ""
    form.active = true
    form.contentText = INITIAL_CONTENT
    currentVersion.value = null
    error.value = null
    contentError.value = null
  }

  async function load() {
    const id = templateId()
    if (!id) return reset()
    loading.value = true
    error.value = null
    try {
      apply(await getTemplate(context(), id))
    } catch (e) {
      error.value = errorMessage(e)
    } finally {
      loading.value = false
    }
  }

  function parseContent(): Content | null {
    contentError.value = null
    try {
      const v: unknown = JSON.parse(form.contentText)
      if (v === null || typeof v !== "object" || Array.isArray(v) || Object.keys(v).length === 0) {
        contentError.value = "El contenido debe ser un objeto JSON no vacío."
        return null
      }
      return v as Content
    } catch {
      contentError.value = "El contenido no es un JSON válido."
      return null
    }
  }

  async function save(): Promise<Template | null> {
    error.value = null
    const name = form.name.trim()
    if (!name || name.length > 150) {
      error.value = "El nombre es obligatorio y admite hasta 150 caracteres."
      return null
    }
    if (form.description.length > 1000) {
      error.value = "La descripción admite hasta 1000 caracteres."
      return null
    }
    const content = parseContent()
    if (!content) return null

    saving.value = true
    try {
      const base = { name, description: form.description, content, active: form.active }
      const id = templateId()
      const saved = id
        ? await updateTemplate(context(), id, { ...base, expectedVersion: currentVersion.value ?? 1 })
        : await createTemplate(context(), base)
      apply(saved)
      return saved
    } catch (e) {
      error.value =
        e instanceof ApiError && e.status === 409
          ? "Otra persona modificó la plantilla. Recarga para ver la versión actual."
          : errorMessage(e)
      return null
    } finally {
      saving.value = false
    }
  }

  watch(templateId, load, { immediate: true })

  return { form, currentVersion, isEditing, loading, saving, error, contentError, save, load }
}
