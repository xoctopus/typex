PACKAGES=$(shell go list ./... | grep -E -v 'example|proto')
FORMAT_FILES=`find . -type f -name '*.go' | grep -E -v '_generated.go|.pb.go'`


debug:
	@echo ${PACKAGES}

tidy:
	@go mod tidy

cover: tidy vet
	@go test -race -failfast -parallel 1 -gcflags="all=-N -l" ${PACKAGES} -covermode=atomic -coverprofile=cover.out

test: tidy vet
	@go test -race -failfast -parallel 1 -gcflags="all=-N -l" ${PACKAGES}

vet:
	@go vet ${PACKAGES}

report:
	@echo ">>>static checking"
	@go vet ./...
	@echo "done\n"
	@echo ">>>detecting ineffectual assignments"
	@ineffassign ./...
	@echo "done\n"
	@echo ">>>detecting icyclomatic complexities over 10 and average"
	@gocyclo -over 10 -avg -ignore '_test|vendor' . || true
	@echo "done\n"

check: tidy vet test

MOD=$(shell cat go.mod | grep ^module -m 1 | awk '{ print $$2; }' || '')
FILE_LIST=$(shell find . -type f -name '*.go' | grep -E -v '_generated.go|.pb.go')

.PHONY: fmt
fmt:
	@echo ${MOD}
	@for item in ${FILE_LIST} ; \
	do \
		if [ -z ${MOD} ]; then \
		        goimports -d -w $$item ; \
		else \
		        goimports -d -w -local "${MOD}" $$item ; \
		fi \
	done

