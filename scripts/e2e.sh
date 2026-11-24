#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
: "${TEST_DATABASE_URL:?Define TEST_DATABASE_URL sobre PostgreSQL de pruebas con Goose aplicado}"
scratch=$(mktemp -d)
trap 'rm -rf -- "$scratch"' EXIT
tree=$root
mode=${1:-normal}

case "$mode" in
  normal) ;;
  sin-proteccion)
    tree="$scratch/mutacion"
    mkdir -p "$tree/apps"
    cp -R "$root/apps/api" "$tree/apps/api"
    cp -R "$root/db" "$root/e2e" "$root/sqlc.yaml" "$tree/"
    # Mutación aislada: una versión más reciente también coincide con la versión esperada.
    target="$tree/db/queries/plantillas.sql"
    test "$(grep -F -c 'p.version_actual = sqlc.arg(version_esperada)' "$target")" -eq 1
    sed -i 's/p.version_actual = sqlc.arg(version_esperada)/p.version_actual >= sqlc.arg(version_esperada)/' "$target"
    (cd "$tree" && sqlc generate)
    ;;
  *) printf 'Modo inválido: %s\n' "$mode" >&2; exit 2 ;;
esac

mkdir -p "$root/coverage" "$scratch/coverage"
export E2E_API_BIN="$scratch/api"
export E2E_COVERAGE_DIR="$scratch/coverage"
cd "$tree/apps/api"
go build -cover -covermode=atomic -o "$E2E_API_BIN" ./cmd/api

if [[ "$mode" == sin-proteccion ]]; then
  TZ=UTC go test -count=1 -v ./e2e
  exit
fi

for zone in UTC Pacific/Auckland; do
  printf 'E2E con TZ=%s\n' "$zone"
  TZ="$zone" go test -count=1 -v ./e2e
done
go tool covdata textfmt -i="$E2E_COVERAGE_DIR" -o="$root/coverage/e2e.cover"
go tool cover -func="$root/coverage/e2e.cover" | tee "$root/coverage/resumen.txt"
sed -i 's#^github.com/example/microservicio-plantilla-reporte/#apps/api/#' "$root/coverage/e2e.cover"
