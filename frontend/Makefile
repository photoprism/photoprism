# Copyright © 2018 - 2025 PhotoPrism UG. All rights reserved.
#
# Questions? Email us at hello@photoprism.app or visit our website to learn
# more about our team, products and services: https://www.photoprism.app/

UID := $(shell id -u)
GID := $(shell id -g)
PWD := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
NODE_IMAGE ?= node:lts

define DOCKER_NPM
docker run --rm -it \
  -v "$(PWD)":/app \
  -w /app \
  --user "$(UID):$(GID)" \
  --read-only --tmpfs /tmp \
  -v ~/.cache/npm:/npm-cache \
  -e NPM_CONFIG_CACHE=/npm-cache \
  -e NPM_CONFIG_AUDIT=false \
  -e NPM_CONFIG_FUND=false \
  -e NO_UPDATE_NOTIFIER=true \
  --cap-drop ALL --security-opt no-new-privileges --pids-limit 256 \
  $(NODE_IMAGE) sh -lc
endef

# Declare "make" targets.
all: install build
dep: install-npm install-testcafe install
dep-list:
	npx npm-check-updates
notice:
	@echo "Creating license report for frontend dependencies..."
	license-report --only=prod --config=.report.json > NOTICE
install-npm:
	sudo npm install --unsafe-perm=true --allow-root -g npm@latest npm-check-updates@latest license-report@latest
install-testcafe:
	npm install -g testcafe@latest
install-eslint:
	npm install -g eslint globals @eslint/eslintrc @eslint/js eslint-config-prettier eslint-formatter-pretty eslint-plugin-html eslint-plugin-import eslint-plugin-node eslint-plugin-prettier eslint-plugin-promise eslint-plugin-vue eslint-webpack-plugin vue-eslint-parser prettier
upgrade:
	$(info Securely upgrading NPM dependencies...)
	$(DOCKER_NPM) 'npx -y npm@latest update --save --ignore-scripts --no-update-notifier && npx -y npm@latest install --ignore-scripts --no-audit --no-fund --no-update-notifier'
npm-install:
	$(info Installing NPM dependencies...)
	npm install --ignore-scripts --no-update-notifier --no-audit --no-audit --no-fund
install: npm-install
npm-update:
	$(info Updating NPM dependencies in package.lock and package-lock.json...)
	npm update --save --package-lock --ignore-scripts --no-update-notifier --no-audit --no-fund
update: npm-update npm-install
watch:
	npm run watch
build:
	npm run build
lint:
	npm run lint
fmt:
	npm run fmt
test:
	npm run test
testcafe:
	npm run testcafe
acceptance-local:
	npm run acceptance-local
gettext-extract:
	npm run gettext-extract
gettext-compile:
	npm run gettext-compile

# Declare all targets as "PHONY", see https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html.
MAKEFLAGS += --always-make
