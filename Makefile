SHELL := /bin/bash

.PHONY: generate verify verify-generated verify-proto verify-status verify-library verify-manager verify-ops

generate:
	pnpm generate

verify: verify-generated verify-proto verify-status verify-library verify-manager verify-ops

verify-generated: generate
	git diff --exit-code -- gen/go packages/status-contracts-web/src/gen packages/status-contracts-dart/lib/gen packages/library-contracts-web/src/gen services/manager/src/gen packages/manager-contracts-dart/lib/gen

verify-proto:
	pnpm check:proto

verify-status:
	test -z "$$(find services/status -type f -name '*.go' -print0 | xargs -0 gofmt -l)"
	cd services/status && go vet ./... && go build ./...
	pnpm --filter @realtime-me/status-web check

verify-library:
	test -z "$$(find services/library -path services/library/vendor -prune -o -type f -name '*.go' -print0 | xargs -0 gofmt -l)"
	cd services/library && go vet ./... && go build ./...
	pnpm --filter '@realtime-me/library-*' --if-present typecheck
	pnpm --filter '@realtime-me/library-*' --if-present build

verify-manager:
	pnpm --filter @realtime-me/manager check
	pnpm --filter @realtime-me/manager build

verify-ops:
	bash -n deploy/library/scripts/*.sh deploy/library/operator/*.sh deploy/manager/scripts/*.sh
	shellcheck --severity=warning deploy/library/scripts/*.sh deploy/library/operator/*.sh deploy/manager/scripts/*.sh
	PYTHONPATH=deploy/library/operator python3 -B -c 'import compose_expected, compose_policy, compose_rendered_policy, compose_source_policy'
	python3 -B deploy/library/operator/validate-compose.py source deploy/library/docker-compose.yml
