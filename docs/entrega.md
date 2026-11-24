# Entrega

Los [comandos de entrega](entrega-comandos.md) usan PowerShell y GitHub CLI. El Hito 2 genera un borrador con los tres adjuntos cuando el run del tag pasa la matriz E2E, SonarQube y los trabajos de seguridad e imagen.

1. Confirma nombre, N, consumidores y proveedores con el documento de componentes.
2. Acepta la invitación a PINGESO-CIVIL. Gitleaks detectó dos coincidencias en archivos de trabajo de `.opencode/` introducidos por el commit `6c840b7`. Confirma con el dueño si las credenciales siguen activas y rótalas si corresponde. Acuerda el saneamiento con los colaboradores. Sobre un clon de trabajo con respaldo, `git filter-repo --path .opencode/ --invert-paths` retira ese directorio de todo el historial y cambia los SHA afectados. Verifica ese clon con `gitleaks git --redact --log-opts="--all" .` y revisa los datos confidenciales antes de publicar. Usa los SHA del historial saneado en los pasos siguientes.
3. El dueño cambia el repositorio a público y lo transfiere a PINGESO-CIVIL con el nombre del servicio. Ajusta el remoto con `git remote set-url origin https://github.com/PINGESO-CIVIL/<servicio>.git`. Pide protección de main y permisos al ayudante.
4. Abre el PR del Hito 1 con el commit de mocks como punta y fusiónalo antes de incorporar el Hito 2. Confirma `calidad`, `seguridad` e `imagen` en verde para ese commit en main.
5. Crea `git tag -a hito-1 <SHA> -m "Entrega Hito 1"` y ejecuta `git push origin refs/tags/hito-1`. Comprueba el run del tag.
6. Publica la release desde `hito-1`, usa las [notas](releases/hito-1.md), enlaza el run y adjunta el `openapi.yaml` de ese tag. Comparte la release con los consumidores.
7. Abre el issue Entrega Hito 1 con la plantilla del repositorio y enlaza la release.

## Hito 2

1. En SonarQube, crea o pide al ayudante el proyecto y su Quality Gate. Usa Standard Experience y configura, para código nuevo, 0 bugs, 0 vulnerabilidades, 100% de hotspots revisados, duplicación estrictamente menor que 3% y cobertura de al menos 60%. CI exige las cinco condiciones con medidas para su análisis. En GitHub configura `SONAR_HOST_URL` y `SONAR_PROJECT_KEY` como variables y `SONAR_TOKEN` como secreto. Añade el enlace del proyecto al README.
2. Incorpora el Hito 2 a main después de publicar `hito-1`. Define los datos reales del proveedor de identidad, confirma los encabezados y la cookie de sesión, y revisa los proveedores que declare el documento de componentes. Si el Hito 1 se integró con un commit distinto, incorpora los cambios posteriores mediante un PR nuevo sobre main.
3. Verifica los tres jobs en verde sobre el SHA final. Descarga los artefactos `evidencia-calidad-hito-2` y `sbom-cyclonedx`. El primer artefacto incluye la comparación con `hito-1`, cobertura E2E y log de la mutación.
4. En Actions, ejecuta manualmente **E2E sin protección de versión** contra el mismo código de Hito 2. Verifica que el run rojo muestre específicamente `respuestas [200 200], esperadas [200 409]`. Guarda su enlace.
5. Ejecuta `git tag -a hito-2 <SHA> -m "Entrega Hito 2"` y `git push origin refs/tags/hito-2`. Verifica el run del tag y publica el borrador creado por CI con las [notas](releases/hito-2.md), `openapi.yaml`, `sbom.cdx.json` y `oasdiff-hito-1.txt`.
6. Abre el issue Entrega Hito 2. Enlaza la release, Quality Gate de ese SHA, oasdiff, golden y run rojo. Si se necesita una excepción de Trivy, incorpora CVE, justificación, responsable y fecha de revisión en `.trivyignore.yaml`, con aprobación del ayudante antes de configurarla en los escaneos.
