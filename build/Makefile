.PHONY: build clean

all: build

build:
	mkdir -p src/
	cp ../*.go src/
	docker build -t syslog-redirector-tmp .
	docker run --name syslog-redirector-build-tmp syslog-redirector-tmp
	docker cp syslog-redirector-build-tmp:/syslog-redirector .
	docker rm syslog-redirector-build-tmp
	docker rmi syslog-redirector-tmp

clean:
	rm -rf src/
