VERSION ?= "0.0.0-dev"
LDFLAGS=-ldflags "-s -w -X github.com/axllent/myback/cmd.Version=${VERSION}"
BINARY=myback
DOCKERIMG=axllent/myback

build = echo "\n\nBuilding $(1)-$(2)" && GO386=softfloat CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build ${LDFLAGS} -o dist/${BINARY}$(3) \
	&& if [ $(1) = "windows" ]; then \
		zip -9jq dist/${BINARY}_$(1)_$(2).zip dist/${BINARY}$(3) LICENSE && rm -f dist/${BINARY}$(3); \
	else \
		tar --owner=root --group=root -C dist/ -zcf dist/${BINARY}_$(1)_$(2).tar.gz ${BINARY} ../LICENSE && rm -f dist/${BINARY}; \
	fi

build: *.go go.*
	CGO_ENABLED=0 go build ${LDFLAGS} -o ${BINARY}
	rm -rf /tmp/go-*

compress: build
	upx -9 ${BINARY}

clean:
	rm -f ${BINARY}

release:
	mkdir -p dist
	rm -f dist/${BINARY}_*
	$(call build,linux,amd64,)
	$(call build,linux,386,)
	$(call build,linux,arm64,)
	$(call build,linux,arm,)
	$(call build,darwin,amd64,)
	$(call build,darwin,arm64,)
	$(call build,windows,386,.exe)
	$(call build,windows,amd64,.exe)

docker:
	docker build --build-arg VERSION=${VERSION} -t ${DOCKERIMG} -f contrib/Dockerfile .
	docker tag ${DOCKERIMG} ${DOCKERIMG}:${VERSION}

docker-release: docker
	docker push ${DOCKERIMG}:${VERSION}
	docker push ${DOCKERIMG}
