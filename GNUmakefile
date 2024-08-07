SHELL := /bin/bash
GO_FILES?=$(find . -name '*.go' |grep -v vendor)

export GOBIN := ${PWD}/bin
export PATH := $(GOBIN):$(PATH)

# To provide the version use 'make release VERSION=1.1.1 GPGKEY=<example@efficientip.com>'
ifdef VERSION
	RELEASE := $(VERSION)
else
	RELEASE := 99999.9
endif

ifdef GPGKEY
	GPGKEYOPTION := -u $(GPGKEY)
else
	GPGKEYOPTION :=
endif

# Terraform 13 local registry handler
PKG_NAME := solidserver
OS_ARCH := linux_amd64
TERRAFORM_PLUGINS_DIRECTORY := ~/.terraform.d/plugins/terraform.efficientip.com/efficientip/${PKG_NAME}/${RELEASE}/${OS_ARCH}

default: build

update:
	go get -u ./...

build:
	go get -v ./...
	go mod tidy
	go mod vendor
	if ! [ -d ${TERRAFORM_PLUGINS_DIRECTORY} ]; then mkdir -p ${TERRAFORM_PLUGINS_DIRECTORY}; fi
	env CGO_ENABLED=0 go build -o ${TERRAFORM_PLUGINS_DIRECTORY}/terraform-provider-${PKG_NAME}
	if [ -d ./_tests ]; then cd _tests && rm -rf .terraform* && cd ..; fi

release:
  #The binary name format is terraform-provider-{NAME}_v{VERSION}
  #The archive name format is terraform-provider-{NAME}_{VERSION}_{OS}_{ARCH}.zip
	go get -v ./...
	go mod tidy
	go mod vendor
	if ! [ -d './_releases' ]; then mkdir './_release'; fi
	if ! [ -d "./_releases/$(RELEASE)" ]; then mkdir "./_releases/$(RELEASE)"; else rm -rf ./_releases/$(RELEASE)/*; fi

	if ! [ -d "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_arm" ]; then mkdir "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_arm"; fi
	if ! [ -d "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_arm64" ]; then mkdir "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_arm64"; fi
	if ! [ -d "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_386" ]; then mkdir "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_386"; fi
	if ! [ -d "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_amd64" ]; then mkdir "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_amd64"; fi
	if ! [ -d "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_freebsd_amd64" ]; then mkdir "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_freebsd_amd64"; fi
	if ! [ -d "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_windows_386" ]; then mkdir "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_windows_386"; fi
	if ! [ -d "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_windows_amd64" ]; then mkdir "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_windows_amd64"; fi
	if ! [ -d "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_darwin_arm64" ]; then mkdir "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_darwin_arm64"; fi
	if ! [ -d "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_darwin_amd64" ]; then mkdir "./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_darwin_amd64"; fi

	cp -r ./README.md ./LICENSE ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_arm/
	cp -r ./README.md ./LICENSE ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_arm64/
	cp -r ./README.md ./LICENSE ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_386/
	cp -r ./README.md ./LICENSE ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_amd64/
	cp -r ./README.md ./LICENSE ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_freebsd_amd64/
	cp -r ./README.md ./LICENSE ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_windows_386/
	cp -r ./README.md ./LICENSE ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_windows_amd64/
	cp -r ./README.md ./LICENSE ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_darwin_arm64/
	cp -r ./README.md ./LICENSE ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_darwin_amd64/

	env GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -o ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_arm/terraform-provider-solidserver_v$(RELEASE)
	env GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_arm64/terraform-provider-solidserver_v$(RELEASE)
	env GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_386/terraform-provider-solidserver_v$(RELEASE)
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_amd64/terraform-provider-solidserver_v$(RELEASE)
	env GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 go build -o ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_freebsd_amd64/terraform-provider-solidserver_v$(RELEASE)
	env GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -o ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_windows_386/terraform-provider-solidserver_v$(RELEASE)
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_windows_amd64/terraform-provider-solidserver_v$(RELEASE)
	env GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_darwin_arm64/terraform-provider-solidserver_v$(RELEASE)
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_darwin_amd64/terraform-provider-solidserver_v$(RELEASE)

	zip -j -r ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_arm.zip ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_arm64
	zip -j -r ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_arm64.zip ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_arm64
	zip -j -r ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_386.zip ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_386
	zip -j -r ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_amd64.zip ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_amd64
	zip -j -r ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_freebsd_amd64.zip ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_freebsd_amd64
	zip -j -r ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_windows_386.zip ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_windows_386
	zip -j -r ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_windows_amd64.zip ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_windows_amd64
	zip -j -r ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_darwin_arm64.zip ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_darwin_arm64
	zip -j -r ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_darwin_amd64.zip ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_darwin_amd64

	rm -rf ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_arm
	rm -rf ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_arm64
	rm -rf ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_386
	rm -rf ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_linux_amd64
	rm -rf ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_freebsd_amd64
	rm -rf ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_windows_386
	rm -rf ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_windows_amd64
	rm -rf ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_darwin_arm64
	rm -rf ./_releases/$(RELEASE)/terraform-provider-solidserver_$(RELEASE)_darwin_amd64

	cd ./_releases/$(RELEASE) && shasum -a 256 *.zip > ./terraform-provider-solidserver_$(RELEASE)_SHA256SUMS && cd ../..
	cd ./_releases/$(RELEASE) && gpg $(GPGKEYOPTION) --detach-sign ./terraform-provider-solidserver_$(RELEASE)_SHA256SUMS && cd ../..

test: fmtcheck vet
	go test -v ./... || exit 1

doc: tools
	@sh -c "'$(CURDIR)/scripts/gendoc.sh'"

fmt:
	gofmt -s -w ./*.go
	gofmt -s -w ./solidserver/*.go

vet:
	go vet -all ./solidserver

tools:
	@echo "Installing development tools"
	go generate -tags tools tools/tools.go

fmtcheck:
	./scripts/gofmtcheck.sh

.PHONY: tools