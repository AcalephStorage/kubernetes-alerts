APP_NAME = kube-alerts

all: clean format deps build

clean:
	@echo "--> Cleaning build"
	@rm -rf ./bin

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
	@mv bin/kubernetes-alerts bin/${APP_NAME}
	@tar cf bin/${APP_NAME}-`uname -s | tr [A-Z] [a-z]`-amd64.tar -C bin ${APP_NAME}
	@rm bin/${APP_NAME}
