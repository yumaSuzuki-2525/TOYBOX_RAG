include .env  # Make実行時に.env読み込み
export

up:
	docker-compose up -d

down:
	docker-compose down

re-build:
	$(MAKE) down
	docker-compose build
	$(MAKE) up

cli:
	go run ./cmd/cli

import_pdf:
	go run ./cmd/import

# pprof
pprof_cpu:
	go tool pprof cpu.prof
pprof_trace:
	go tool trace trace.out

# go tool cover -html coverage.out