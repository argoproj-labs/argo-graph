.PHONY: start
start: graph/staticfiles.go
	go run ./cmd server

ui/node_modules:
	yarn --cwd ui install

ui/dist/app/index.html: ui/node_modules ui/src
	# Build UI
	yarn --cwd ui build
	touch ui/dist/app/index.html

$(HOME)/go/bin/staticfiles:
	# Install the "staticfiles" tool
	go get bou.ke/staticfiles

graph/staticfiles.go: $(HOME)/go/bin/staticfiles ui/dist/app/index.html
	# Pack UI into a Go file.
	staticfiles -o graph/staticfiles.go ui/dist/app
