# Entrega

1. Confirma nombre, N, consumidores y proveedores con el documento de componentes.
2. Acepta la invitación a PINGESO-CIVIL. Ejecuta `gitleaks git --redact .` y revisa los datos confidenciales en todo el historial. Si se encuentra un secreto, rótalo y sanea el historial antes de publicarlo.
3. El dueño cambia el repositorio a público y lo transfiere a PINGESO-CIVIL con el nombre del servicio. Ajusta el remoto con `git remote set-url origin https://github.com/PINGESO-CIVIL/<servicio>.git`. Pide protección de main y permisos al ayudante.
4. Abre el PR del Hito 1 con el commit de mocks como punta y fusiónalo antes de incorporar el Hito 2. Confirma `calidad`, `seguridad` e `imagen` en verde para ese commit en main.
5. Crea `git tag -a hito-1 <SHA> -m "Entrega Hito 1"` y ejecuta `git push origin refs/tags/hito-1`. Comprueba el run del tag.
6. Publica la release desde `hito-1`, usa las [notas](releases/hito-1.md), enlaza el run y adjunta el `openapi.yaml` de ese tag. Comparte la release con los consumidores.
7. Abre el issue Entrega Hito 1 con la plantilla del repositorio y enlaza la release.
