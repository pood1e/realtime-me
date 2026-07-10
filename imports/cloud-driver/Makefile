SHELL := /bin/bash

.PHONY: generate verify verify-proto verify-api verify-web verify-ops build-api build-web

generate:
	cd proto && buf generate

build-api:
	cd api && go build ./...

build-web:
	pnpm run build:web

verify: verify-proto verify-api verify-web verify-ops

verify-proto:
	cd proto && buf lint

verify-api:
	test -z "$$(find api -path api/vendor -prune -o -type f -name '*.go' -print0 | xargs -0 gofmt -l)"
	cd api && go vet ./... && go build ./...

verify-web:
	pnpm exec prettier --check 'web/apps/**/*.{ts,tsx,css,json,html}' 'web/packages/shared/**/*.{ts,tsx,css,json}' 'web/build/**/*.ts'
	pnpm run typecheck:web
	pnpm run build:web

verify-ops:
	bash -n ops/scripts/*.sh ops/operator/*.sh
	shellcheck --severity=warning ops/scripts/*.sh ops/operator/*.sh
