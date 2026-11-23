<script setup lang="ts">
import { useTemplatesList } from "./use-templates-list"
import type { ApiContext } from "@/types/template"

const props = defineProps<{ context: ApiContext }>()
const emit = defineEmits<{
  edit: [id: string]
  versions: [id: string]
  deleted: [id: string]
}>()

const { items, pagination, totalPages, loading, error, load, remove } =
  useTemplatesList(() => props.context)

async function confirmDelete(id: string, name: string) {
  if (!window.confirm(`¿Eliminar la plantilla "${name}"?`)) return
  if (await remove(id)) emit("deleted", id)
}

defineExpose({ refresh: () => load() })
</script>

<template>
  <section class="templates-list">
    <header class="header">
      <h2>Plantillas</h2>
      <button type="button" :disabled="loading" @click="load()">Actualizar</button>
    </header>

    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <p v-else-if="loading">Cargando…</p>
    <p v-else-if="items.length === 0">No hay plantillas.</p>

    <table v-else>
      <thead>
        <tr>
          <th>Nombre</th>
          <th>Versión</th>
          <th>Estado</th>
          <th>Actualizada</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in items" :key="item.id">
          <td>
            <strong>{{ item.name }}</strong>
            <small v-if="item.description">{{ item.description }}</small>
          </td>
          <td>v{{ item.currentVersion }}</td>
          <td>{{ item.active ? "Activa" : "Inactiva" }}</td>
          <td>{{ new Date(item.updatedAt).toLocaleString("es-CL") }}</td>
          <td class="actions">
            <button type="button" @click="emit('edit', item.id)">Editar</button>
            <button type="button" @click="emit('versions', item.id)">Versiones</button>
            <button type="button" class="danger" @click="confirmDelete(item.id, item.name)">
              Eliminar
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <footer v-if="items.length" class="pagination">
      <button type="button" :disabled="pagination.page <= 1 || loading" @click="load(pagination.page - 1)">
        Anterior
      </button>
      <span>Página {{ pagination.page }} de {{ totalPages }} ({{ pagination.total }} en total)</span>
      <button type="button" :disabled="pagination.page >= totalPages || loading" @click="load(pagination.page + 1)">
        Siguiente
      </button>
    </footer>
  </section>
</template>

<style scoped>
.header, .pagination { display: flex; align-items: center; justify-content: space-between; gap: 1rem; }
table { width: 100%; border-collapse: collapse; margin: 1rem 0; }
th, td { text-align: left; padding: 0.5rem; border-bottom: 1px solid #d9e2e1; vertical-align: top; }
small { display: block; color: #5b6770; }
.actions { display: flex; gap: 0.4rem; flex-wrap: wrap; }
.error, .danger { color: #b3261e; }
</style>
