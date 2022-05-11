.PHONY: frontend
frontend: build

.PHONY: watch
watch: run

.PHONY: run
run: node_modules
	npm start

.PHONY: install
install: node_modules

.PHONY: install-ci
install-ci:
	npm ci

.PHONY: build
build: install
	npm run build

.PHONY: test
test:
	npm run test -- --coverage

.PHONY: lint
lint:
	npm run lint

.PHONY: tsc
tsc:
	npm run tsc

.PHONY: i18n
i18n:
	npm run i18n

.PHONY: ci
ci: install-ci lint test build tsc

node_modules:
	npm install
