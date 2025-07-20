PROTO_DIR := proto
GO_OUT := .


UNAME_S := $(shell uname -s)

ifeq ($(findstring MINGW,$(UNAME_S)),MINGW)
    # Windows (Git Bash)
    PROTO_SRC := $(shell find $(PROTO_DIR) -name "*.proto")
else
    # Linux / Mac / WSL
    PROTO_SRC := $(wildcard $(PROTO_DIR)/*.proto)
endif

.PHONY: generate-proto
generate-proto:
	@echo "PROTO_SRC = $(PROTO_SRC)"
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(GO_OUT) \
		--go-grpc_out=$(GO_OUT) \
		$(PROTO_SRC)