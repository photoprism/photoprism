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

# Overview of the most commonly used targets, shown by "make help".
# Keep this in sync when renaming or adding an important target; "make list" shows all of them.
define HELP_TEXT
PhotoPrism Frontend Makefile

Usage: make -C frontend <target> [target ...]

Dependencies:
  install          Install the NPM dependencies
  ci               Clean install of the NPM dependencies from the lockfile
  update           Update the NPM dependencies and their version ranges
  upgrade          Update the NPM dependencies in a sandboxed container
  dep-list         List the NPM dependencies that have newer versions
  audit            Check the NPM dependencies for known vulnerabilities
  notice           Regenerate the NOTICE file for the frontend dependencies

Build:
  build            Build the frontend for production
  watch            Rebuild the frontend when source files change

Test:
  test             Run the unit tests with Vitest
  test-coverage    Run the unit tests with a code coverage report
  test-component   Run the component tests only
  acceptance-local Run the TestCafe acceptance tests against a local server

Format & Lint:
  fmt              Fix the JS with ESLint and format the CSS with Prettier
  lint             Check the JS with ESLint and the CSS with Prettier
  security-check   Scan for unsafe NPM commands and HTML injection sinks

Translations:
  gettext-extract  Extract the translation strings into the catalogs
  gettext-compile  Compile the translation catalogs

Other:
  list             List all targets, including those not shown here
endef
export HELP_TEXT

# Declare "make" targets.
all: install build
dep: install-npm install-testcafe install
dep-list:
	npx npm-check-updates
help:
	@printf '%s\n' "$$HELP_TEXT"
list:
	@awk '/^define /{skip=1} /^endef/{skip=0;next} skip{next} /^[[:alnum:]]+[^[:space:]]+:/ {printf "%s",substr($$1,1,length($$1)-1); if (match($$0,/#/)) {desc=substr($$0,RSTART+1); sub(/^[[:space:]]+/,"",desc); printf " - %s\n",desc} else printf "\n" }' "$(firstword $(MAKEFILE_LIST))"
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
