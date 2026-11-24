# Arquitectura y propiedad de datos

`cmd/api` ensambla el router, la autenticación y el repositorio. La petición sigue Handler → Service → Repository. Los handlers sólo llaman al service. `internal/api` contiene los DTOs y la política de rutas generada. El almacenamiento se inyecta mediante la interfaz Repository.

Plantillas es dueño de las plantillas y sus versiones. Las identidades pertenecen al proveedor de identidad, y Plantillas guarda referencias UUID del propietario y autor. El servicio accede exclusivamente a su propio almacenamiento.

PostgreSQL almacena plantillas, versiones, auditoría y registros de idempotencia. La aplicación usa las consultas de `db/queries` mediante el código sqlc de `internal/repository`. El adaptador `PostgresRepository` confirma la versión, su evento de auditoría y, cuando corresponde, la respuesta de idempotencia en una sola transacción. El borrado es lógico y conserva el historial. El proveedor de identidad autentica al usuario y emite el JWT. La API obtiene claves públicas del JWKS con timeout de 5 s, propaga el identificador de petición en esa llamada y, tras validar el Bearer, puede crear una cookie de sesión local.

El dominio carece de montos, cupos, inventario y escrituras en otros servicios. La protección de concurrencia del recurso usa `version_esperada`.

Se conserva la estructura `apps/api` y `apps/web` hasta cotejarla con la plantilla oficial. El código actual carece de integraciones de negocio salientes. El documento de componentes debe confirmar los consumidores, proveedores, nombre oficial y puertos.
