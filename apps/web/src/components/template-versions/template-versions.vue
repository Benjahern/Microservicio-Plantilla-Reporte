<script setup lang="ts">
import { useTemplateVersions } from "./use-template-versions"
import type { ApiContext } from "@/types/template"

const props = defineProps<{ context: ApiContext; templateId: string }>()
const emit = defineEmits<{ close: [] }>()

const { versions, pagination, totalPages, selected, loading, error, load, select } =
  useTemplateVersions(
    () => props.context,
    () => props.templateId,
  )
</script>

<template>
  <section class="template-versions">
    <header class="header">
      <h2>Historial de versiones</h2>
      <button type="button" @click="emit('close')">Cerrar</button>
    </header>

    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <p v-if="loading">Cargando…</p>

    <ul v-else class="list">
      <li v-for="v in versions" :key="v.number">
        <button type="button" :class="{ active: selected?.number === v.number }" @click="select(v.number)">
          <strong>v{{ v.number }}</strong> · {{ v.name }}
          <small>{{ new Date(v.createdAt).toLocaleString("es-CL") }} · {{ v.authorId }}</small>
        </button>
      </li>
    </ul>

    <footer v-if="versions.length" class="pagination">
      <button type="button" :disabled="pagination.page <= 1" @click="load(pagination.page - 1)">Anterior</button>
      <span>Página {{ pagination.page }} de {{ totalPages }}</span>
      <button type="button" :disabled="pagination.page >= totalPages" @click="load(pagination.page + 1)">
        Siguiente
      </button>
    </footer>

    <article v-if="selected" class="detail">
      <h3>v{{ selected.number }} — {{ selected.name }}</h3>
      <p>{{ selected.description || "Sin descripción" }} ({{ selected.active ? "activa" : "inactiva" }})</p>
      <pre>{{ JSON.stringify(selected.content, null, 2) }}</pre>
    </article>
  </section>
</template>

<style scoped>
.header, .pagination { display: flex; align-items: center; justify-content: space-between; gap: 1rem; }
.list { list-style: none; padding: 0; display: grid; gap: 0.4rem; }
.list button { width: 100%; text-align: left; padding: 0.5rem; }
.list button.active { border-color: #007a6e; background: #e0f0ee; }
small { display: block; color: #5b6770; }
pre { background: #f4f7f7; padding: 0.75rem; overflow-x: auto; }
.error { color: #b3261e; }
</style>
