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
	# Build Storybook
	npm run build-storybook --quiet

	# Serve Storybook and run tests in parallel
	npx concurrently -k -s first -n "SB,TEST" -c "magenta,blue" \
		"npx http-server storybook-static --port 6006 --silent" \
		"npx wait-on tcp:127.0.0.1:6006 && npm run test-storybook"

.PHONY: test-playwright-ci
test-playwright-ci:
	# Run Playwright tests and capture failure
	npx playwright test || touch playwright-failed

	# Check if tests failed and create a report
	@if [ -f "playwright-failed" ]; then \
		echo "Playwright tests failed."; \
		printf "%s\n" \
		"🔍 To check failed test in GitHub Actions:" \
		"" \
		"1️⃣ **Download the HTML Report:**" \
		"   - Go to the **Artifacts** section in GitHub Actions." \
		"   - Click on **playwright-report** to download the ZIP file." \
		"" \
		"2️⃣ **View the Report:**" \
		"   - Extract the ZIP file in a folder with Playwright installed." \
		"   - Run: **npx playwright show-report <extracted-folder-name>**" \
		"" \
		"3️⃣ **Analyze Failures with Trace Viewer:**" \
		"   - Open the report and click the **trace** icon next to a failed test." \
		"   - Inspect each test action to diagnose the issue." \
		"" \
		"📖 More details: https://playwright.dev/docs/test-reporters#html-reporter"; \
		exit 1; \
	fi

.PHONY: ci
ci: install-ci lint test install-playwright test-storybook-ci test-playwright-ci build tsc

node_modules:
	npm install
