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

# OIDC test infrastructure commands
OIDC_COMPOSE_CMD = cd ../backend && docker compose -f docker-compose.base.yaml -f docker-compose.oidc-test.yaml

.PHONY: oidc-up
oidc-up:
	$(OIDC_COMPOSE_CMD) up --build -d

.PHONY: oidc-down
oidc-down:
	$(OIDC_COMPOSE_CMD) down

# Regular test infrastructure commands  
REGULAR_COMPOSE_CMD = cd ../backend && docker compose -f docker-compose.base.yaml -f docker-compose.test.yaml

.PHONY: regular-test-down
regular-test-down:
	$(REGULAR_COMPOSE_CMD) down

.PHONY: test-playwright-ci
test-playwright-ci:
	# Run Playwright tests and capture failure, then ensure cleanup
	npx playwright test || ($(MAKE) regular-test-down && exit 1)
	$(MAKE) regular-test-down

.PHONY: test-oidc
test-oidc:
	npm run test:oidc

.PHONY: test-oidc-ci
test-oidc-ci: oidc-up
	# Run OIDC tests (Playwright setup handles service readiness checks)
	npx playwright test --config=playwright.oidc.config.ts || ($(MAKE) oidc-down && exit 1)
	$(MAKE) oidc-down 

.PHONY: ci
ci: install-ci lint test install-playwright test-storybook-ci test-playwright-ci test-oidc-ci tsc build 

node_modules:
	npm install
