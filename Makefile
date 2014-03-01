NAME = $(shell basename `pwd` | cut -d '-' -f 1)
VERSION = 0.1

HOST = lightswitchrave.net

EXE = $(shell basename `pwd`)
PKGNAME = $(EXE)
RPMSHORT = $(PKGNAME)-$(VERSION)-1.fc$(shell head -n1 /etc/issue | cut -d ' ' -f 3).x86_64.rpm
RPM = package/RPMS/x86_64/$(RPMSHORT)

all: $(EXE)

rpm: $(RPM)

$(EXE):
	go test
	go build
	mkdir -p site
	cp -a $(EXE) site/
	cp -a add-user site/
	cp -a config site/
	cp -a err static tmpl site/

put: $(RPM)
	rsync -az $(RPM) $(HOST):.
	ssh $(HOST) -t sudo rpm --force -fvi ./$(RPMSHORT)

$(RPM): $(EXE)
	cp -a site $(PKGNAME)-$(VERSION)
	mkdir -p package/{RPMS,BUILD,SOURCES,BUILDROOT}
	tar -czf package/SOURCES/$(PKGNAME)-$(VERSION).tar.gz $(PKGNAME)-$(VERSION)
	rm -rf $(PKGNAME)-$(VERSION)
	cat server.service.in | sed "s/%NAME%/$(NAME)/g" >package/SOURCES/server.service
	cat server.spec.in | sed "s/%NAME%/$(NAME)/g" | sed "s/%VERSION%/$(VERSION)/g" >server.spec
	rpmbuild --define "_topdir $(PWD)/package" -ba server.spec
	rm -rf package/{BUILD,BUILDROOT}

clean:
	rm -f server.spec

.PHONY: $(EXE) rpm clean
