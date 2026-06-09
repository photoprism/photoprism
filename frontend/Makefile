# Copyright © 2018 - 2026 PhotoPrism UG. All rights reserved.
#
# Questions? Email us at hello@photoprism.app or visit our website to learn
# more about our team, products and services: https://www.photoprism.app/

UID := $(shell id -u)
GID := $(shell id -g)
PWD := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
REPO_ROOT := $(abspath $(PWD)/..)
NODE_IMAGE ?= node:lts
export NPM_CONFIG_IGNORE_SCRIPTS ?= true

ifneq (,$(wildcard $(REPO_ROOT)/.telemetry))
include $(REPO_ROOT)/.telemetry
export $(shell sed -n 's/^[[:space:]]*\([A-Z_][A-Z0-9_]*\)=.*/\1/p' $(REPO_ROOT)/.telemetry)
endif

define DOCKER_NPM
docker run --rm -it \
  -v "$(REPO_ROOT)":/app \
  -w /app \
  --user "$(UID):$(GID)" \
  --read-only --tmpfs /tmp \
  -v ~/.cache/npm:/npm-cache \
  --env-file "$(REPO_ROOT)/.telemetry" \
  -e NPM_CONFIG_CACHE=/npm-cache \
  -e NPM_CONFIG_IGNORE_SCRIPTS=true \
  --cap-drop ALL --security-opt no-new-privileges --pids-limit 256 \
  $(NODE_IMAGE) sh -lc
endef

# Declare "make" targets.
all: install build
dep: install-npm install-testcafe install
dep-list:
	npx npm-check-updates
help: list
list:
	@awk '/^[[:alnum:]]+[^[:space:]]+:/ {printf "%s",substr($$1,1,length($$1)-1); if (match($$0,/#/)) {desc=substr($$0,RSTART+1); sub(/^[[:space:]]+/,"",desc); printf " - %s\n",desc} else printf "\n" }' "$(firstword $(MAKEFILE_LIST))"
notice:
	@echo "Creating license report for frontend dependencies..."
	license-report --only=prod --config=.report.json > NOTICE
install-npm:
	# Split global installs and disable package scripts for safety.
	sudo npm install -g --ignore-scripts --no-audit --no-fund --no-update-notifier npm@latest
	sudo npm install -g --ignore-scripts --no-audit --no-fund --no-update-notifier npm-check-updates@latest license-report@latest
install-testcafe:
	npm install -g --ignore-scripts --no-audit --no-fund --no-update-notifier testcafe@latest
install-eslint:
	npm install -g --ignore-scripts --no-audit --no-fund --no-update-notifier eslint globals @eslint/eslintrc @eslint/js eslint-config-prettier eslint-formatter-pretty eslint-plugin-html eslint-plugin-import eslint-plugin-node eslint-plugin-promise eslint-plugin-vue eslint-webpack-plugin vue-eslint-parser prettier
upgrade:
	$(info Securely upgrading NPM dependencies...)
	$(DOCKER_NPM) 'npx -y npm@latest update --save --package-lock --ignore-scripts --no-audit --no-fund --no-update-notifier && npx -y npm@latest install --ignore-scripts --no-audit --no-fund --no-update-notifier'
ci: npm-ci
npm-ci:
	@$(MAKE) -C .. --no-print-directory npm-ci
install: npm-install
npm-install:
	@$(MAKE) -C .. --no-print-directory npm-install
update: npm-update npm-install
npm-update:
	@$(MAKE) -C .. --no-print-directory npm-update
audit: npm-audit
npm-audit:
	@$(MAKE) -C .. --no-print-directory npm-audit
security-check: # Scan for missing --ignore-scripts and unsafe v-html
	npm run -s security:scan
watch:
	npm run watch
build:
	npm run build
lint:
	npm run lint || true
fmt:
	npm run fmt
test:
	npm run test
test-coverage:
	npm run test-coverage
test-component:
	npm run test-component
testcafe:
	npm run testcafe
acceptance-local:
	npm run acceptance-local
gettext-extract:
	# Script sets GETTEXT_MERGE=0 by default so vue-gettext skips msgmerge; scripts/gettext-merge.sh handles merging.
	npm run gettext-extract
gettext-compile:
	npm run gettext-compile

# Declare all targets as "PHONY", see https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html.
MAKEFLAGS += --always-make
