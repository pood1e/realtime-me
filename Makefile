SHELL := /bin/bash

.PHONY: generate generate-proto generate-mobile verify verify-generated verify-style verify-proto verify-status verify-library verify-manager verify-console verify-site verify-watch verify-mobile verify-ops

generate: generate-proto generate-mobile

generate-proto:
	pnpm generate

generate-mobile:
	cd apps/mobile && dart run pigeon --input pigeons/status_bridge.dart
	cd apps/mobile && dart format lib/core/platform/status_bridge.g.dart >/dev/null
	node tools/normalize-text.mjs apps/mobile/android/app/src/main/kotlin/me/realtime/mobile/platform/StatusBridge.g.kt
	node tools/generate-android-connect-procedures.mjs

verify: verify-generated verify-style verify-proto verify-status verify-library verify-manager verify-console verify-site verify-watch verify-mobile verify-ops

verify-generated: generate
	git diff --exit-code -- gen/go packages/auth-contracts-web/src/gen packages/status-contracts-web/src/gen packages/status-contracts-dart/lib/gen packages/library-contracts-web/src/gen packages/manager-contracts-web/src/gen services/manager/src/gen packages/manager-contracts-dart/lib/gen apps/mobile/lib/core/platform/status_bridge.g.dart apps/mobile/android/app/src/main/kotlin/me/realtime/mobile/platform/StatusBridge.g.kt apps/mobile/android/app/src/main/kotlin/me/realtime/mobile/status/StatusGatewayProcedures.kt

verify-style:
	pnpm check:style

verify-proto:
	pnpm check:proto

verify-status:
	test -z "$$(find services/status -type f -name '*.go' -print0 | xargs -0 gofmt -l)"
	go vet ./services/status/... ./libs/go/authn
	go build ./services/status/...
	pnpm --filter @realtime-me/status-web typecheck

verify-library:
	test -z "$$(find services/library -path services/library/vendor -prune -o -type f -name '*.go' -print0 | xargs -0 gofmt -l)"
	go vet ./services/library/...
	go build ./services/library/...
	pnpm --filter @realtime-me/library-web typecheck

verify-manager:
	pnpm --filter @realtime-me/manager check
	pnpm --filter @realtime-me/manager build

verify-console:
	test -z "$$(find services/console libs/go/authn -type f -name '*.go' -print0 | xargs -0 gofmt -l)"
	go vet ./services/console/...
	go build ./services/console/...
	pnpm --filter @realtime-me/console check
	pnpm --filter @realtime-me/console build

verify-site:
	pnpm --filter @realtime-me/site check
	pnpm --filter @realtime-me/site build

verify-watch:
	./gradlew :apps:watch:lintDebug :apps:watch:assembleDebug

verify-mobile:
	cd apps/mobile && flutter analyze
	cd apps/mobile/android && ./gradlew app:lintDebug
	cd apps/mobile && flutter build apk --debug

verify-ops:
	find deploy scripts -type f -name '*.sh' -print0 | xargs -0 -n1 bash -n
	find deploy scripts -type f -name '*.sh' -print0 | xargs -0 shellcheck --severity=warning
	python3 -m compileall -q scripts
	cd scripts/probe && diff -u <(find realtime_probe -type f -name '*.py' | LC_ALL=C sort) <(LC_ALL=C sort manifest.txt)
	PYTHONPATH=deploy/library/operator python3 -B -c 'import compose_expected, compose_policy, compose_rendered_policy, compose_source_policy'
	python3 -B deploy/library/operator/validate-compose.py source deploy/library/compose.yaml
