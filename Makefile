EXE=$(shell basename `pwd`)

all: $(EXE)

$(EXE):
	go test
	go install
	mkdir -p site
	cp $(GOPATH)/bin/$(EXE) site/
	rsync -avz add-user config err static tmpl site

put: $(EXE)
	rsync -cavz site/ boosd.org:site-$(EXE)

clean:
	rm -rf site

.PHONY: $(EXE) clean
