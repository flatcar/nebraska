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

.PHONY: test-storybook-ci
test-storybook-ci:
	# Install dependencies
	npm ci

	# Install Playwright
	npx playwright install --with-deps

	# Build Storybook
	npm run build-storybook --quiet

	# Serve Storybook and run tests in parallel
	npx concurrently -k -s first -n "SB,TEST" -c "magenta,blue" \
		"npx http-server storybook-static --port 6006 --silent" \
		"npx wait-on tcp:127.0.0.1:6006 && npm run test-storybook"

.PHONY: ci
ci: install-ci lint test test-storybook-ci build tsc

node_modules:
	npm install
