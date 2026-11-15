include .env  # Make実行時に.env読み込み
export
# setup
up:
	docker-compose up -d

down:
	docker-compose down

re-build:
	$(MAKE) down
	docker-compose build
	$(MAKE) up
# run
cli:
	go run ./cmd/cli

import_pdf:
	go run ./cmd/import

# pprof
pprof_cpu:
	go tool pprof cpu.prof
pprof_trace:
	go tool trace trace.out

# moq update
generate:
	go generate ./...
	
# テスト雛形の生成（例: make gen-test FILE=./internal/infrastructure/analyzer/kagome.go FUNC=ChunkTextBySemantics）
# もしくは対象関数名上で右クリック、Go: Generate Unit Tests For Function
gen-test:
	gotests -only $(FUNC) -w $(FILE)


# go tool cover -html coverage.out