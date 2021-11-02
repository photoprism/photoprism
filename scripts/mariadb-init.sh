#!/usr/bin/env bash

# Create default databases
cat << EOF
CREATE DATABASE IF NOT EXISTS keycloak;
CREATE USER IF NOT EXISTS keycloak@'%' IDENTIFIED BY 'keycloak';
GRANT ALL PRIVILEGES ON keycloak.* TO keycloak@'%';

CREATE DATABASE IF NOT EXISTS photoprism_latest;
CREATE USER IF NOT EXISTS photoprism_latest@'%' IDENTIFIED BY 'photoprism_latest';
GRANT ALL PRIVILEGES ON photoprism_latest.* TO photoprism_latest@'%';

CREATE DATABASE IF NOT EXISTS photoprism_preview;
CREATE USER IF NOT EXISTS photoprism_preview@'%' IDENTIFIED BY 'photoprism_preview';
GRANT ALL PRIVILEGES ON photoprism_preview.* TO photoprism_preview@'%';

CREATE DATABASE IF NOT EXISTS acceptance;
CREATE USER IF NOT EXISTS acceptance@'%' IDENTIFIED BY 'acceptance';
GRANT ALL PRIVILEGES ON acceptance.* TO acceptance@'%';

EOF

# Create additional test databases
for USER_ID in $(seq -f "%02g" 1 5)
do
	echo "CREATE DATABASE IF NOT EXISTS photoprism_$USER_ID;"
	echo "CREATE USER IF NOT EXISTS photoprism_$USER_ID@'%' IDENTIFIED BY 'photoprism_$USER_ID';";
	echo "GRANT ALL PRIVILEGES ON photoprism_$USER_ID.* TO photoprism_$USER_ID@'%';"
done

cat << EOF

FLUSH PRIVILEGES;
EOF
