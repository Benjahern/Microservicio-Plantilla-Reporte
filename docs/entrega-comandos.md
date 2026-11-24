# Entrega de los Hitos 1 y 2 en GitHub

Ejecuta estos comandos en PowerShell con Git, Git Filter Repo, Gitleaks y GitHub CLI instalados. Completa los valores solicitados con la información del curso.

## 1. Sanear una copia y preparar el repositorio

Gitleaks detectó dos coincidencias en archivos `.opencode/` del commit inicial `6c840b7`. Antes de publicar, acuerda el saneamiento con el dueño y los colaboradores. Verifica si las credenciales siguen activas, rótalas si corresponde y revisa si el historial contiene datos internos del cliente.

```powershell
$Source = 'C:\Users\conyv\Desktop\Semestre 2-2026\Microservicio-Plantilla-Reporte'
$Clean = Join-Path ([System.IO.Path]::GetTempPath()) ('plantillas-entrega-' + [guid]::NewGuid())
git clone --no-local --branch entrega/hito-2 $Source $Clean
if ($LASTEXITCODE -ne 0) { throw 'Falló el clon' }
Set-Location $Clean
git branch entrega/hito-1 origin/entrega/hito-1
if ($LASTEXITCODE -ne 0) { throw 'Falta la rama Hito 1' }
git filter-repo --path .opencode/ --invert-paths
if ($LASTEXITCODE -ne 0) { throw 'Falló el saneamiento' }
gitleaks git --redact=100 --log-opts='--all' .
if ($LASTEXITCODE -ne 0) { throw 'El historial aún no está listo para publicar' }
$Hito1 = git rev-parse entrega/hito-1
$Hito2Tip = git rev-parse entrega/hito-2

gh auth login --hostname github.com
$Servicio = Read-Host 'Nombre oficial del servicio'
$Repo = "PINGESO-CIVIL/$Servicio"
gh repo create "$Repo" --public
if ($LASTEXITCODE -ne 0) { throw 'Confirma el repositorio con el dueño antes de subir código' }
git remote add curso "https://github.com/$Repo.git"
git push curso 'entrega/hito-1:refs/heads/main'
gh repo edit "$Repo" --default-branch main
```

Si el curso ya asignó un repositorio o exige transferir el existente, resuelve esa operación con el dueño y usa la URL del repositorio confirmado en vez de crear otro. Solicita al ayudante los permisos y la protección de `main` y de los tags `hito-*` según el curso.

## 2. Entregar Hito 1

En Actions confirma que el push a `main` del SHA `$Hito1` esté en verde. Después crea el tag y espera el run del tag.

```powershell
gh run list --repo "$Repo" --workflow ci.yml --commit "$Hito1"
git tag -a hito-1 "$Hito1" -m 'Entrega Hito 1'
git push curso refs/tags/hito-1
gh run list --repo "$Repo" --workflow ci.yml --commit "$Hito1" --event push
$RunHito1 = Read-Host 'ID del run del tag hito-1'
gh run watch "$RunHito1" --repo "$Repo" --exit-status
if ($LASTEXITCODE -ne 0) { throw 'La verificación de hito-1 no aprobó' }
git archive --format=zip -o hito1-artefactos.zip hito-1 contracts/openapi.yaml docs/releases/hito-1.md
Expand-Archive -LiteralPath hito1-artefactos.zip -DestinationPath hito1-artefactos
gh release create hito-1 'hito1-artefactos/contracts/openapi.yaml' --repo "$Repo" --verify-tag --title 'Hito 1' --notes-file 'hito1-artefactos/docs/releases/hito-1.md'
```

Edita las notas para enlazar el run del tag, el SHA y los modos de falla. Comparte la release con los consumidores confirmados. En el paso 5 se incluye el comando para abrir el issue.

## 3. Configurar SonarQube antes de integrar Hito 2

En SonarQube crea el proyecto y asígnale un Quality Gate para **Standard Experience**. Sobre código nuevo agrega condiciones que fallen con bugs > 0, vulnerabilidades > 0, hotspots revisados < 100 %, cobertura < 60 % y duplicación > 2,99 %. Desactiva la omisión de cobertura y duplicación para cambios pequeños si el servidor la aplica. CI exige las cinco condiciones y además verifica el valor de duplicación estrictamente menor que 3 % en el análisis del run. Usa un token con permiso de analizar ese proyecto y consultar el resultado de su análisis por API.

```powershell
$SonarUrl = Read-Host 'URL HTTPS de SonarQube'
$SonarProject = Read-Host 'Clave del proyecto SonarQube'
gh variable set SONAR_HOST_URL --repo "$Repo" --body "$SonarUrl"
gh variable set SONAR_PROJECT_KEY --repo "$Repo" --body "$SonarProject"
gh secret set SONAR_TOKEN --repo "$Repo"
```

Agrega el enlace del proyecto al README cuando esté disponible. Confirma `N`, plantilla, consumidores y proveedores, y prepara la configuración real de JWT/JWKS y HTTPS con el proveedor de identidad.

Configura en el servidor una definición de código nuevo que produzca medidas sobre los cambios del Hito 2. Si necesitas un análisis de referencia, establece una línea base con el commit del Hito 1 antes de fusionar. Con SonarScanner CLI instalado, puedes cargarlo así y solicitar al ayudante que lo use como período de referencia según la edición del servidor:

```powershell
$Baseline = Join-Path $Clean 'sonar-baseline-hito1'
git worktree add --detach "$Baseline" hito-1
$Secret = Read-Host 'Token SonarQube para el análisis base' -AsSecureString
$env:SONAR_TOKEN = [System.Net.NetworkCredential]::new('', $Secret).Password
$env:SONAR_HOST_URL = $SonarUrl
Push-Location "$Baseline"
try {
  sonar-scanner "-Dsonar.projectKey=$SonarProject" '-Dsonar.projectVersion=hito-1' '-Dsonar.sources=apps/api,apps/web/src'
  if ($LASTEXITCODE -ne 0) { throw 'Falló el análisis base' }
} finally {
  Pop-Location
  Remove-Item Env:SONAR_TOKEN
}
```

El workflow evalúa SonarQube en los pushes a `main` y al tag `hito-2`. Las verificaciones de PR ejecutan la matriz y el resto de los controles sin requerir secretos.

## 4. Integrar y publicar Hito 2

```powershell
git push curso 'entrega/hito-2:refs/heads/entrega/hito-2'
gh pr create --repo "$Repo" --base main --head entrega/hito-2 --title 'Entrega Hito 2' --body-file docs/releases/hito-2.md
$PR = Read-Host 'Número del PR de Hito 2'
gh pr checks "$PR" --repo "$Repo" --watch
if ($LASTEXITCODE -ne 0) { throw 'El PR todavía tiene verificaciones pendientes o fallidas' }
gh pr merge "$PR" --repo "$Repo" --merge --match-head-commit "$Hito2Tip"
git fetch curso main
$Hito2 = git rev-parse curso/main
gh run list --repo "$Repo" --workflow ci.yml --commit "$Hito2"
$RunMain = Read-Host 'ID del run verde de main para ese SHA'
gh run watch "$RunMain" --repo "$Repo" --exit-status
if ($LASTEXITCODE -ne 0) { throw 'El SHA de main no aprobó' }
git tag -a hito-2 "$Hito2" -m 'Entrega Hito 2'
git push curso refs/tags/hito-2
gh run list --repo "$Repo" --workflow ci.yml --commit "$Hito2" --event push
$RunHito2 = Read-Host 'ID del run del tag hito-2'
gh run watch "$RunHito2" --repo "$Repo" --exit-status
if ($LASTEXITCODE -ne 0) { throw 'El run del tag no aprobó' }
gh workflow run e2e-sin-proteccion.yml --repo "$Repo" --ref hito-2
gh run list --repo "$Repo" --workflow e2e-sin-proteccion.yml --commit "$Hito2"
gh release view hito-2 --repo "$Repo" --json isDraft,assets,url
```

Confirma el run rojo de la mutación y su mensaje `respuestas [200 200], esperadas [200 409]`. Completa en las notas del borrador los enlaces al run rojo, Quality Gate y golden. Verifica que estén adjuntos `openapi.yaml`, `sbom.cdx.json` y `oasdiff-hito-1.txt`. Publica con:

```powershell
gh release edit hito-2 --repo "$Repo" --draft=false
```

## 5. Abrir los issues de entrega

Ejecuta el bloque con el número de hito correspondiente en cuanto publiques su release. Examina el cuerpo generado y añade los enlaces a las evidencias antes de enviarlo.

```powershell
$NroHito = Read-Host 'Hito a entregar (1 o 2)'
if ($NroHito -notin @('1', '2')) { throw 'Hito inválido' }
$Tag = "hito-$NroHito"
$Release = gh release view "$Tag" --repo "$Repo" --json url --jq .url
if ($LASTEXITCODE -ne 0) { throw 'Falta publicar la release' }
$Template = Get-Content -Raw -LiteralPath (Join-Path $Source ".github/ISSUE_TEMPLATE/entrega-hito-$NroHito.md")
$Body = ($Template -replace '(?s)\A---\r?\n.*?\r?\n---\r?\n', '') -replace '(?m)^Release:.*$', "Release: $Release"
$Draft = Join-Path $Clean "$Tag-issue.md"
Set-Content -LiteralPath $Draft -Value $Body -Encoding utf8
notepad $Draft
# Después de completar y guardar los enlaces:
gh issue create --repo "$Repo" --title "Entrega Hito $NroHito" --body-file "$Draft"
```
