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

.PHONY: install-playwright
install-playwright:
	npx playwright install --with-deps

.PHONY: build
build: install
	NODE_ENV=production npm run build

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
	npm run build-storybook:ci && \
		npx concurrently -k -s first -n SB,TEST "npm run serve-storybook:ci" "npm run test-storybook:ci"

.PHONY: test-playwright-ci
test-playwright-ci:
	# Run Playwright tests and capture failure
	npx playwright test 

.PHONY: ci
ci: install-ci lint test install-playwright test-storybook-ci test-playwright-ci tsc build 

node_modules:
	npm install
