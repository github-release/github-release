LAST_TAG := $(shell git describe --abbrev=0 --tags)

USER := aktau
EXECUTABLE := github-release

# only include the amd64 binaries, otherwise the github release will become
# too big
EXECUTABLES := \
	darwin/amd64/$(EXECUTABLE) \
	freebsd/amd64/$(EXECUTABLE) \
	linux/amd64/$(EXECUTABLE) \
	windows/amd64/$(EXECUTABLE)

COMPRESSED_EXECUTABLES=$(EXECUTABLES:%=%.tar.bz2)
COMPRESSED_EXECUTABLE_TARGETS=$(EXECUTABLES:%=bin/%.tar.bz2)

UPLOAD_CMD = bin/tmp/$(EXECUTABLE) upload -u $(USER) -r $(EXECUTABLE) -t $(LAST_TAG) -n $(subst /,-,$(FILE)) -f bin/$(FILE)

# the executable used to perform the upload, dogfooding and all...
bin/tmp/$(EXECUTABLE):
	go build -o "$@"

# arm
bin/linux/arm/5/$(EXECUTABLE):
	GOARM=5 GOARCH=arm GOOS=linux go build -o "$@"
bin/linux/arm/7/$(EXECUTABLE):
	GOARM=7 GOARCH=arm GOOS=linux go build -o "$@"

# 386
bin/darwin/386/$(EXECUTABLE):
	GOARCH=386 GOOS=darwin go build -o "$@"
bin/linux/386/$(EXECUTABLE):
	GOARCH=386 GOOS=linux go build -o "$@"
bin/windows/386/$(EXECUTABLE):
	GOARCH=386 GOOS=windows go build -o "$@"

# amd64
bin/freebsd/amd64/$(EXECUTABLE):
	GOARCH=amd64 GOOS=freebsd go build -o "$@"
bin/darwin/amd64/$(EXECUTABLE):
	GOARCH=amd64 GOOS=darwin go build -o "$@"
bin/linux/amd64/$(EXECUTABLE):
	GOARCH=amd64 GOOS=linux go build -o "$@"
bin/windows/amd64/$(EXECUTABLE):
	GOARCH=amd64 GOOS=windows go build -o "$@"

# compressed artifacts, makes a huge difference (Go executable is ~9MB,
# after compressing ~2MB)
%.tar.bz2: %
	tar -jcvf "$<.tar.bz2" "$<"

# git tag -a v$(RELEASE) -m 'release $(RELEASE)'
release: bin/tmp/$(EXECUTABLE) $(COMPRESSED_EXECUTABLE_TARGETS)
	git push && git push --tags
	bin/tmp/$(EXECUTABLE) release -u $(USER) -r $(EXECUTABLE) \
		-t $(LAST_TAG) -n $(LAST_TAG) || true
	$(foreach FILE,$(COMPRESSED_EXECUTABLES),$(UPLOAD_CMD);)

clean:
	rm -rf bin/

.PHONY: clean release
