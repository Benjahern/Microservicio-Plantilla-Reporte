<script setup lang="ts">
import { ref } from "vue"
import TemplatesList from "@/components/templates-list/templates-list.vue"
import TemplateEditor from "@/components/template-editor/template-editor.vue"
import TemplateVersions from "@/components/template-versions/template-versions.vue"
import type { ApiContext } from "@/types/template"

type View = "list" | "editor" | "versions"

const context: ApiContext = { baseUrl: "" }

const view = ref<View>("list")
const templateId = ref<string | undefined>()
const list = ref<InstanceType<typeof TemplatesList> | null>(null)

function openNew() {
  templateId.value = undefined
  view.value = "editor"
}
function openEditor(id: string) {
  templateId.value = id
  view.value = "editor"
}
function openVersions(id: string) {
  templateId.value = id
  view.value = "versions"
}
function backToList() {
  view.value = "list"
  void list.value?.refresh()
}
</script>

<template>
  <main class="app">
    <header class="toolbar">
      <h1>Plantillas de reporte <small>(app de desarrollo)</small></h1>
      <button v-if="view === 'list'" type="button" @click="openNew">Nueva plantilla</button>
    </header>

    <TemplatesList v-show="view === 'list'" ref="list" :context="context" @edit="openEditor" @versions="openVersions" />
    <TemplateEditor
      v-if="view === 'editor'"
      :context="context"
      :template-id="templateId"
      @saved="backToList"
      @cancel="backToList"
    />
    <TemplateVersions
      v-if="view === 'versions' && templateId"
      :context="context"
      :template-id="templateId"
      @close="backToList"
    />
  </main>
</template>

<style>
body { margin: 0; font-family: system-ui, sans-serif; color: #2f3640; }
button { font: inherit; padding: 0.35rem 0.75rem; cursor: pointer; }
</style>

<style scoped>
.app { max-width: 60rem; margin: 0 auto; padding: 1.5rem; display: grid; gap: 1.5rem; }
.toolbar { display: grid; gap: 0.75rem; }
h1 { margin: 0; font-size: 1.5rem; color: #007a6e; }
h1 small { font-size: 0.8rem; color: #5b6770; }
</style>
