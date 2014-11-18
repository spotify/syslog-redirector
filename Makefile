.PHONY: build package

all: build package

build:
	make -C build all
	cp build/syslog-redirector .

package:
	cp syslog-redirector package/
	make -C package all

clean:
	rm -f package/syslog-redirector
	rm -f syslog-redirector
	make -C build clean
	make -C package clean

push:
	make -C package push