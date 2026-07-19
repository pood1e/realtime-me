SHELL := /bin/bash

.PHONY: generate generate-proto generate-mobile generate-probe-integrity verify verify-generated verify-style verify-proto verify-go-shared verify-status verify-library verify-manager verify-console verify-site verify-watch verify-mobile verify-ops

generate: generate-proto generate-mobile generate-probe-integrity

generate-proto:
	pnpm generate

generate-mobile:
	cd apps/mobile && dart run pigeon --input pigeons/status_bridge.dart
	cd apps/mobile && dart format lib/core/platform/status_bridge.g.dart >/dev/null
	node tools/normalize-text.mjs apps/mobile/android/app/src/main/kotlin/me/realtime/mobile/platform/StatusBridge.g.kt
	node tools/generate-android-connect-procedures.mjs

generate-probe-integrity:
	python3 scripts/probe/generate-integrity.py

verify: verify-generated verify-style verify-proto verify-status verify-library verify-manager verify-console verify-site verify-watch verify-mobile verify-ops

verify-generated: generate
	git diff --exit-code -- gen/go packages/auth-contracts-web/src/gen packages/status-contracts-web/src/gen packages/status-contracts-dart/lib/gen packages/library-contracts-web/src/gen packages/manager-contracts-web/src/gen services/manager/src/gen packages/manager-contracts-dart/lib/gen apps/mobile/lib/core/platform/status_bridge.g.dart apps/mobile/android/app/src/main/kotlin/me/realtime/mobile/platform/StatusBridge.g.kt apps/mobile/android/app/src/main/kotlin/me/realtime/mobile/status/StatusGatewayProcedures.kt scripts/install-probe.py scripts/probe/integrity.json

verify-style:
	pnpm check:style

verify-proto:
	pnpm check:proto

verify-go-shared:
	test -z "$$(find libs/go/authn libs/go/serviceauth -type f -name '*.go' -print0 | xargs -0 gofmt -l)"
	go vet ./libs/go/authn ./libs/go/serviceauth

verify-status: verify-go-shared
	test -z "$$(find services/status -type f -name '*.go' -print0 | xargs -0 gofmt -l)"
	go vet ./services/status/...
	go build ./services/status/...
	pnpm --filter @realtime-me/status-web typecheck

verify-library: verify-go-shared
	test -z "$$(find services/library -path services/library/vendor -prune -o -type f -name '*.go' -print0 | xargs -0 gofmt -l)"
	go vet ./services/library/...
	go build ./services/library/...
	pnpm --filter @realtime-me/library-web typecheck

verify-manager:
	pnpm --filter @realtime-me/manager check
	pnpm --filter @realtime-me/manager build

verify-console: verify-go-shared
	test -z "$$(find services/console -type f -name '*.go' -print0 | xargs -0 gofmt -l)"
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
	cmp gradlew apps/mobile/android/gradlew
	cmp gradlew.bat apps/mobile/android/gradlew.bat
	cmp gradle/wrapper/gradle-wrapper.jar apps/mobile/android/gradle/wrapper/gradle-wrapper.jar
	cmp gradle/wrapper/gradle-wrapper.properties apps/mobile/android/gradle/wrapper/gradle-wrapper.properties
	cd apps/mobile && flutter pub get
	cd apps/mobile && flutter analyze
	cd apps/mobile/android && ./gradlew app:lintDebug
	cd apps/mobile && flutter build apk --debug

verify-ops:
	find deploy scripts -type f -name '*.sh' -print0 | xargs -0 -n1 bash -n
	find deploy scripts -type f -name '*.sh' -print0 | xargs -0 shellcheck --severity=warning
	python3 -m compileall -q scripts
	python3 scripts/probe/generate-integrity.py --check
	PYTHONPATH=deploy/library/operator python3 -B -c 'import compose_policy, compose_rendered_policy, compose_source_policy'
	python3 -B deploy/library/operator/validate-compose.py source deploy/library/compose.yaml
