@echo off

echo Starting PhotoPrism and MariaDB...

docker compose up -d
docker compose logs -f

echo Done.