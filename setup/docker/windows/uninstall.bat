@echo off
setlocal
:PROMPT
SET /P AREYOUSURE=Are you sure (Y/[N])?
IF /I "%AREYOUSURE%" NEQ "Y" GOTO END

echo Stopping PhotoPrism and MariaDB...

docker compose down -v

echo Removing Docker images...

docker compose rm -s -v

:END
endlocal