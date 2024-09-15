# Copyright © 2018 - 2024 PhotoPrism UG. All rights reserved.
#
# Questions? Email us at hello@photoprism.app or visit our website to learn
# more about our team, products and services: https://www.photoprism.app/

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
	sudo npm install --unsafe-perm=true --allow-root -g testcafe@3.6.2
install:
	npm install --no-update-notifier --no-audit
update:
	npm update
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
upgrade:
	npm run upgrade
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
