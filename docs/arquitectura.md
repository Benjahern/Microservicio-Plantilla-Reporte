# Arquitectura y propiedad de datos

`cmd/api` ensambla el router, la autenticación y el repositorio. La petición sigue Handler → Service → Repository. Los handlers sólo llaman al service. `internal/api` contiene los DTOs y la política de rutas generada. El almacenamiento se inyecta mediante la interfaz Repository.

Plantillas es dueño de las plantillas y sus versiones. Las identidades pertenecen al proveedor de identidad, y Plantillas guarda referencias UUID del propietario y autor. El servicio accede exclusivamente a su propio almacenamiento.

Se conserva la estructura `apps/api` y `apps/web` hasta cotejarla con la plantilla oficial. El código actual carece de integraciones de negocio salientes. El documento de componentes debe confirmar los consumidores, proveedores, nombre oficial y puertos.
