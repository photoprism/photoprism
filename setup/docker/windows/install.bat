@echo off

Rem If you don't have Docker installed yet, please follow this guide to download
Rem and install Docker Desktop before you proceed:
Rem
Rem   https://docs.docker.com/desktop/install/windows-install/
Rem
Rem With Docker up and running, change to the directory where you want to install PhotoPrism,
Rem and then run the following commands in a terminal (command prompt) to download our
Rem configuration examples and start PhotoPrism on your local PC:
Rem
Rem   curl.exe -o install.bat https://dl.photoprism.app/docker/windows/install.bat
Rem   install.bat

echo Checking Docker version...

docker --version
docker compose version

echo Downloading config files...

curl.exe -o docker-compose.yml https://dl.photoprism.app/docker/windows/docker-compose.yml
curl.exe -o start.bat https://dl.photoprism.app/docker/windows/start.bat
curl.exe -o stop.bat https://dl.photoprism.app/docker/windows/stop.bat
curl.exe -o uninstall.bat https://dl.photoprism.app/docker/windows/uninstall.bat

dir

echo Pulling Docker images...

docker compose pull

echo Starting PhotoPrism and MariaDB...

docker compose up -d
timeout /t 20

echo Please open the Web UI by navigating to http://localhost:2342/. You should see a login screen.
echo Then log in with the user 'admin' and the password you have specified in PHOTOPRISM_ADMIN_PASSWORD (default is 'insecure').
echo You can change it on the account settings page. If you enable public mode, authentication will be disabled.
echo Enjoy!

START http://localhost:2342/