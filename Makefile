.PHONY: build deps gofmt install lint test test_deps

BUILD_SUFFIX?=-dev

VERSION=${BUILD_VERSION}-$(shell git rev-parse --short HEAD)${BUILD_SUFFIX}
FLAGS=-ldflags "-X bitbucket.org/bitbucket/conker.Version=${VERSION}"

TIMESTAMP=$(shell date +"%Y-%m-%d_%H-%M-%S")

# all: test install

# deps:
# 	go get github.com/Masterminds/glide
# 	# This is a hack to fix builds while glide master is broken with submodules
# 	cd ${GOPATH}/src/github.com/Masterminds/glide && git checkout v0.12.3 && go install
# 	glide install

# test_deps:
# 	go get github.com/alecthomas/gometalinter
# 	gometalinter --install

#build:
	#mkdir -p ./build/bin
	#go build ${FLAGS} -o build/bin/conker bitbucket.org/bitbucket/conker/cmd/conker
	#go build ${FLAGS} -o build/bin/conker-keycheck bitbucket.org/bitbucket/conker/cmd/conker-keycheck
	#go build ${FLAGS} -o build/bin/conq bitbucket.org/bitbucket/conker/cmd/conq
	#go build ${FLAGS} -o build/bin/conqauth bitbucket.org/bitbucket/conker/cmd/conqauth

# install:
# 	go install ${FLAGS} bitbucket.org/bitbucket/conker/cmd/conker
# 	go install ${FLAGS} bitbucket.org/bitbucket/conker/cmd/conker-keycheck
# 	go install ${FLAGS} bitbucket.org/bitbucket/conker/cmd/conq
# 	go install ${FLAGS} bitbucket.org/bitbucket/conker/cmd/conqauth

# release:
# 	mkdir -p ./build/rpm/7
# 	fpm -t rpm -v ${BUILD_VERSION} -p ./build/rpm/7 --after-install=./rpm/post_install --before-remove=./rpm/pre_uninstall ./rpm/conker.service=/usr/lib/systemd/system/conker.service ./build/bin/=/usr/local/bin

# test:
# 	go test -v -race $(shell glide novendor)

# lint: install
# 	gometalinter ./... --fast --vendor -D dupl -D gas --cyclo-over=11

# gofmt:
# 	gofmt -l -w *.go
# 	glide novendor -x | grep -F './' | xargs gofmt -l -w

####
build:
	go build .

bench:
	go test -bench=. -benchmem

test:
	go test ./...

fmt:
	gofmt -l -w *.go

race:
	go test -race ./...

memprof:
	go test -memprofile mem-${TIMESTAMP}.prof

cpuprof:
	go test -cpuprofile cpu-${TIMESTAMP}.prof