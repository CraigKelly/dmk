BASEDIR=$(CURDIR)
TOOLDIR=$(BASEDIR)/script

BINARY=dmk
SOURCES := $(shell find $(BASEDIR) -name '*.go')
TESTRESOURCES := $(shell find '$(BASEDIR)/res' -type f)
TESTED=.tested

build: $(BINARY)
$(BINARY): $(SOURCES) $(TESTED)
	$(TOOLDIR)/build

clean:
	rm -f $(BINARY) debug debug.test cover.out $(TESTED)

test: $(TESTED) $(TESTRESOURCES)
$(TESTED): $(SOURCES)
	$(TOOLDIR)/test

testv: clean
	$(TOOLDIR)/test -v

cover: $(SOURCES)
	$(TOOLDIR)/cover

update: clean
	$(TOOLDIR)/update

.PHONY: clean test testv cover build run update
