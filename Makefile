VERSION=`git describe --tags`
flags=-ldflags="-s -w -X main.gitVersion=${VERSION}"
debug_flags=-ldflags="-X main.gitVersion=${VERSION}"
odir=`cat ${PKG_CONFIG_PATH}/oci8.pc | grep "libdir=" | sed -e "s,libdir=,,"`

all: build

vet:
	go vet .

build:
	go clean; rm -rf pkg sqlshell*; go build ${flags}

build_debug:
	go clean; rm -rf pkg sqlshell*; go build -gcflags=all="-N -l" ${debug_flags}

build_all: build build_osx build_linux build_power8 build_arm64

build_osx:
	go clean; rm -rf pkg sqlshell_osx; GOOS=darwin go build ${flags}
	mv sqlshell sqlshell_osx

build_linux:
	go clean; rm -rf pkg sqlshell_linux; GOOS=linux go build ${flags}
	mv sqlshell sqlshell_linux

build_power8:
	go clean; rm -rf pkg sqlshell_power8; GOARCH=ppc64le GOOS=linux go build ${flags}
	mv sqlshell sqlshell_power8

build_arm64:
	go clean; rm -rf pkg sqlshell_arm64; GOARCH=arm64 GOOS=linux go build ${flags}
	mv sqlshell sqlshell_arm64

install:
	go install

clean:
	go clean; rm -rf pkg

test-github: test-shell

test-shell:
	cd test && LD_LIBRARY_PATH=${odir} DYLD_LIBRARY_PATH=${odir} go test -v

bench:
	cd test
	go test -run Benchmark -bench=.
