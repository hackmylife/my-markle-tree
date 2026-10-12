SKELETON  := ./skeleton
WORKSPACE := ./workspace

.PHONY: init test test-all vet reset check-workspace

init:
	@if [ -d $(WORKSPACE) ]; then \
		echo "$(WORKSPACE) は既にあります。作り直すときは make reset CONFIRM=yes を実行してください"; \
		exit 1; \
	fi
	cp -R $(SKELETON) $(WORKSPACE)
	@echo "$(WORKSPACE) を作成しました。guide/00-overview.md から始めてください"

check-workspace:
	@if [ ! -d $(WORKSPACE) ]; then \
		echo "$(WORKSPACE) がありません。先に make init を実行してください"; \
		exit 1; \
	fi

test: check-workspace
	@if [ -z "$(STEP)" ]; then \
		echo "STEP を指定してください（例: make test STEP=1）"; \
		exit 1; \
	fi
	go test ./tests -run '^TestStep$(STEP)_' -v

test-all: check-workspace
	go test ./tests

vet: check-workspace
	go vet ./workspace/... ./tests/...

reset:
	@if [ "$(CONFIRM)" != "yes" ]; then \
		echo "$(WORKSPACE) の実装がすべて消えます。実行するときは make reset CONFIRM=yes"; \
		exit 1; \
	fi
	rm -rf $(WORKSPACE)
	cp -R $(SKELETON) $(WORKSPACE)
	@echo "$(WORKSPACE) を雛形に戻しました"
