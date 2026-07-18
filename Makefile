SHELL := /bin/bash

.PHONY: generate generate-proto generate-mobile verify verify-generated verify-style verify-proto verify-status verify-library verify-manager verify-watch verify-mobile verify-ops

generate: generate-proto generate-mobile

generate-proto:
	pnpm generate

generate-mobile:
	cd apps/mobile && dart run pigeon --input pigeons/status_bridge.dart
	cd apps/mobile && dart format lib/core/platform/status_bridge.g.dart >/dev/null
	node tools/normalize-text.mjs apps/mobile/android/app/src/main/kotlin/me/realtime/mobile/platform/StatusBridge.g.kt

verify: verify-generated verify-style verify-proto verify-status verify-library verify-manager verify-watch verify-mobile verify-ops

verify-generated: generate
	git diff --exit-code -- gen/go packages/status-contracts-web/src/gen packages/status-contracts-dart/lib/gen packages/library-contracts-web/src/gen services/manager/src/gen packages/manager-contracts-dart/lib/gen apps/mobile/lib/core/platform/status_bridge.g.dart apps/mobile/android/app/src/main/kotlin/me/realtime/mobile/platform/StatusBridge.g.kt

verify-style:
	pnpm check:style

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

verify-watch:
	./gradlew :apps:watch:lintDebug :apps:watch:assembleDebug

verify-mobile:
	cd apps/mobile && flutter analyze
	cd apps/mobile/android && ./gradlew app:lintDebug
	cd apps/mobile && flutter build apk --debug

verify-ops:
	find deploy scripts -type f -name '*.sh' -print0 | xargs -0 -n1 bash -n
	find deploy scripts -type f -name '*.sh' -print0 | xargs -0 shellcheck --severity=warning
	PYTHONPATH=deploy/library/operator python3 -B -c 'import compose_expected, compose_policy, compose_rendered_policy, compose_source_policy'
	python3 -B deploy/library/operator/validate-compose.py source deploy/library/compose.yaml
