PROTO_REF_DIR=$(CURDIR)/protopgx
PROTO_REF_FILES=$(shell find "$(PROTO_REF_DIR)" -type f -name '*.proto')
compile-proto-ref:
	protoc  --go_out=$(PROTO_REF_DIR) --go_opt=paths=source_relative --proto_path=$(PROTO_REF_DIR) $(PROTO_REF_FILES)

.PHONY: build
build:
	go build ./

.PHONY: build-proto
build-proto:
	protoc --go_out=./protopgx --go_opt=paths=source_relative --proto_path=./protopgx ./protopgx/pgx.proto


ALL_PROTO_FILES = $(shell find './' -type f -name '*.proto')
.PHONY: build-test
build-test: build
	protoc --plugin=./protoc-gen-pgx-orm --pgx-orm_out=./test --pgx-orm_opt=paths=source_relative,sql_file=./test/models.sql,orm_folder=./test/orm --proto_path=. $(ALL_PROTO_FILES)
