#!/bin/bash
set -e

echo "build: admin UI"
cd admin-fe
bun run build
cd ..

echo "build: player embed"
cd packages/player
bun run build:embed
cd ../..

echo "copying to go"
mkdir -p internal/api/dist/embed

# Admin UI -> internal/api/dist/index.html
cp admin-fe/dist/index.html internal/api/dist/index.html

# Player embed -> internal/api/dist/embed/index.html
cp packages/player/dist/embed/index.html internal/api/dist/embed/index.html

echo done
