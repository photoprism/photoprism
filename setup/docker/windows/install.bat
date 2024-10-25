@echo off

Rem With Docker up and running, change to the directory where you want to install PhotoPrism,
Rem and then run the following commands in a terminal (command prompt) to download our
Rem configuration examples and start PhotoPrism on your local PC:
Rem
Rem   curl.exe -o install.bat https://dl.photoprism.app/docker/windows/install.bat
Rem   install.bat

echo If you don't have Docker installed yet, please follow this guide to download
echo and install Docker Desktop before you proceed:
echo:
echo   https://docs.docker.com/desktop/install/windows-install/

timeout /t 10

echo:
echo Checking Docker version...

docker --version
docker compose version

echo:
echo Downloading config files...

curl.exe -s -o compose.yaml https://dl.photoprism.app/docker/windows/compose.yaml
curl.exe -s -o start.bat https://dl.photoprism.app/docker/windows/start.bat
curl.exe -s -o stop.bat https://dl.photoprism.app/docker/windows/stop.bat
curl.exe -s -o uninstall.bat https://dl.photoprism.app/docker/windows/uninstall.bat
curl.exe -s -o update.bat https://dl.photoprism.app/docker/windows/update.bat

echo:
echo Pulling Docker images...

docker compose pull

echo:
echo Starting PhotoPrism and MariaDB...

docker compose up -d
timeout /t 20

echo:
echo You should now be able to log in with the user "admin" when navigating to the following URL:
echo:
echo   http://localhost:2342/
echo:
echo The initial password is "insecure". Please change it under Settings ^> Account before you proceed.

START http://localhost:2342/
