PROTO_SRC_DIR=$(CURDIR)/proto
PROTO_REF_FILES=$(shell find "$(PROTO_SRC_DIR)" -type f -name '*.proto')
PROTO_REF_DIR=$(CURDIR)/protopgx
compile-proto-ref:
	rm -rf $(PROTO_REF_DIR) && mkdir -p $(PROTO_REF_DIR)
	protoc  --go_out=$(PROTO_REF_DIR) --go_opt=paths=source_relative --proto_path=$(PROTO_SRC_DIR) $(PROTO_REF_FILES)


build:
	go build ./