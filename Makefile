.PHONY: start
start: graph/staticfiles.go
	go run ./cmd server

ui/node_modules:
	yarn --cwd ui install

ui/dist/app: ui/node_modules ui/src
	# Build UI
	yarn --cwd ui build
	touch ui/dist/app

$(HOME)/go/bin/staticfiles:
	# Install the "staticfiles" tool
	go get bou.ke/staticfiles

graph/staticfiles.go: $(HOME)/go/bin/staticfiles ui/dist/app
	# Pack UI into a Go file.
	staticfiles -o graph/staticfiles.go ui/dist/app

.PHONY: clean
clean:
	rm -Rf dist graph/staticfiles.go ui/dist

.PHONY: lint
lint:
	golangci-lint -v run --fix --skip-files graph/staticfiles.go
	yarn --cwd ui lint