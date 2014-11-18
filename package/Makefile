.PHONY: package clean

VERSION=$(shell git describe --tags --dirty)

all: package

package:
	docker build -t spotify/syslog-redirector:$(VERSION) .

push:
	docker push spotify/syslog-redirector:$(VERSION)

clean:
