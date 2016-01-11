APP_NAME = kube-alerts

all: clean format deps build

clean:
	@echo "--> Cleaning build"
	@rm -rf ./bin ./tar ./pkg

format:
	@echo "--> Formatting source code"
	@go fmt ./...

deps:
	@echo "--> Getting dependencies"
	@gb vendor restore

# test: format
# 	@echo "--> Testing application"
# 	@gb test ...

build: format
	@echo "--> Building all application"
	@gb build ...
	@mkdir -p bin/`go env GOOS`/`go env GOARCH`
	@mkdir -p tar
	@mv bin/kubernetes-alerts bin/`go env GOOS`/`go env GOARCH`/${APP_NAME}
	@tar cfz tar/${APP_NAME}-`go env GOOS`-`go env GOARCH`.tgz -C bin/`go env GOOS`/`go env GOARCH` ${APP_NAME}