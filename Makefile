WMR_WORKDIR = $(shell pwd)
WMR_EXAMPLES_PATH = $(WMR_WORKDIR)/examples
WMR_EXAMPLE ?= "example-01"
WMR_TUTORIAL ?= "1"

build:
	GOOS=linux go build -o dist/main.out cmd/main/*.go

run:
	make build
	$(WMR_WORKDIR)/dist/main.out --data_dir=$(WMR_EXAMPLES_PATH)/$(WMR_EXAMPLE) $(WMR_RUN_OPTIONS)

run-tutorial:
	make build
	wat2wasm examples/tutorial/_input$(WMR_TUTORIAL).wat -o examples/tutorial/input$(WMR_TUTORIAL).wasm &&  WMR__EXAMPLE=tutorial WMR_RUN_OPTIONS="--allow_empty --in_module='input$(WMR_TUTORIAL).wasm' --in_transform='input$(WMR_TUTORIAL).yml' --out_module='output$(WMR_TUTORIAL).wasm' --verbose --data_dir='examples/tutorial'"  make run && wasm2wat examples/tutorial/output$(WMR_TUTORIAL).wasm -o examples/tutorial/output$(WMR_TUTORIAL).wat --generate-names --fold-exprs

lint:
	golangci-lint run --disable govet

docker-build:
	make build
	docker build -t wasm-manipulator .
	rm -r dist/

docker-run:
	make docker-build
	docker run --rm -it -v $(WMR_EXAMPLES_PATH)/$(WMR_EXAMPLE):/data wasm-manipulator $(WMR_PATH_OPTIONS)

sonar:
	sonar-scanner \
      -Dsonar.projectKey=wasm:manipulator \
      -Dsonar.sources=. \
      -Dsonar.host.url=http://localhost:9000 \
      -Dsonar.login=6814983eb08e2e071635160e8dbd190218474768