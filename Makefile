BASEDIR=$(CURDIR)
TOOLDIR=$(BASEDIR)/script

BINARY=dmk
SOURCES := $(shell find $(BASEDIR) -name '*.go')
TESTRESOURCES := $(shell find '$(BASEDIR)/res' -type f)
TESTED=.tested
VERSIONIN=VERSION
VERSIONOUT=version.go

build: $(BINARY)
$(BINARY): $(SOURCES) $(TESTED) $(VERSIONOUT)
	go build

install: build
	go install

version: $(VERSIONOUT)
$(VERSIONOUT): $(VERSIONIN)
	$(TOOLDIR)/versiongen
	go fmt $(VERSIONOUT)

clean:
	rm -f $(BINARY) debug debug.test cover.out $(TESTED) $(VERSIONOUT)
	$(TOOLDIR)/versiongen

test: $(TESTED) $(TESTRESOURCES)
$(TESTED): $(SOURCES) $(VERSIONOUT)
	$(TOOLDIR)/test

testv: clean $(VERSIONOUT)
	$(TOOLDIR)/test -v

cover: $(SOURCES) $(VERSIONOUT)
	$(TOOLDIR)/cover

update: clean
	$(TOOLDIR)/update

dist: build
	$(TOOLDIR)/dist

release: build dist
	$(TOOLDIR)/release

.PHONY: clean test testv cover build run update version release
