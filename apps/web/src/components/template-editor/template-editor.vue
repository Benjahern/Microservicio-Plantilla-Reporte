<script setup lang="ts">
import { useTemplateEditor } from "./use-template-editor"
import type { ApiContext, Template } from "@/types/template"

const props = defineProps<{ context: ApiContext; templateId?: string }>()
const emit = defineEmits<{ saved: [template: Template]; cancel: [] }>()

const { form, currentVersion, isEditing, loading, saving, error, contentError, save, load } =
  useTemplateEditor(
    () => props.context,
    () => props.templateId,
  )

async function submit() {
  const saved = await save()
  if (saved) emit("saved", saved)
}
</script>

<template>
  <section class="template-editor">
    <h2>
      {{ isEditing ? "Editar plantilla" : "Nueva plantilla" }}
      <small v-if="currentVersion">(versión actual v{{ currentVersion }})</small>
    </h2>

    <p v-if="loading">Cargando…</p>
    <form v-else @submit.prevent="submit">
      <label>
        Nombre
        <input v-model="form.name" type="text" maxlength="150" required />
      </label>

      <label>
        Descripción
        <textarea v-model="form.description" rows="2" maxlength="1000" />
      </label>

      <label class="check">
        <input v-model="form.active" type="checkbox" />
        Activa
      </label>

      <label>
        Contenido (JSON)
        <textarea v-model="form.contentText" class="code" rows="12" spellcheck="false" />
      </label>
      <p v-if="contentError" class="error" role="alert">{{ contentError }}</p>
      <p v-if="error" class="error" role="alert">{{ error }}</p>

      <div class="actions">
        <button type="submit" :disabled="saving">
          {{ saving ? "Guardando…" : isEditing ? "Guardar nueva versión" : "Crear" }}
        </button>
        <button type="button" @click="emit('cancel')">Cancelar</button>
        <button v-if="isEditing" type="button" @click="load()">Recargar</button>
      </div>
    </form>
  </section>
</template>

<style scoped>
form { display: grid; gap: 0.8rem; }
label { display: grid; gap: 0.25rem; font-weight: 600; }
.check { display: flex; align-items: center; gap: 0.5rem; font-weight: 400; }
input[type="text"], textarea { font: inherit; padding: 0.4rem; }
.code { font-family: ui-monospace, monospace; font-size: 0.85rem; }
.actions { display: flex; gap: 0.5rem; }
.error { color: #b3261e; margin: 0; }
small { font-weight: 400; color: #5b6770; }
</style>
