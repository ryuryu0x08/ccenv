# ccenv build & install
# 自动检测当前平台 (Linux/macOS/Windows) 编译并安装到 PATH 目录

BINARY := ccenv

# --- 平台检测 ---------------------------------------------------------------
ifeq ($(OS),Windows_NT)
	GOOS    := windows
	BIN     := $(BINARY).exe
	# Windows 优先 GOBIN，其次 %USERPROFILE%\go\bin
	INSTALL_DIR := $(if $(GOBIN),$(GOBIN),$(USERPROFILE)\go\bin)
else
	UNAME_S := $(shell uname -s)
	BIN     := $(BINARY)
	ifeq ($(UNAME_S),Darwin)
		GOOS := darwin
	else
		GOOS := linux
	endif
	# 类 Unix 优先 GOBIN，其次 $GOPATH/bin，最后 ~/go/bin
	INSTALL_DIR := $(if $(GOBIN),$(GOBIN),$(if $(shell go env GOPATH),$(shell go env GOPATH)/bin,$(HOME)/go/bin))
endif

LDFLAGS := -s -w

# go.mod 可能要求高于本地版本的 Go,触发工具链自动下载。
# 下载工具链需校验数据库,若用户全局 GOSUMDB=off 会失败,故此处显式覆盖。
GOENV := GOTOOLCHAIN=auto GOSUMDB=sum.golang.org GOPROXY=https://proxy.golang.org,direct

.PHONY: all build install clean

all: install

build:
	@echo ">> building $(BIN) for $(GOOS)"
	@$(GOENV) GOOS=$(GOOS) go build -ldflags "$(LDFLAGS)" -o $(BIN) .

install: build
	@echo ">> installing $(BIN) -> $(INSTALL_DIR)"
	@mkdir -p "$(INSTALL_DIR)"
	@cp -f "$(BIN)" "$(INSTALL_DIR)/$(BIN)"
	@echo ">> done. run '$(BINARY)' (ensure $(INSTALL_DIR) is on your PATH)"

clean:
	@rm -f $(BINARY) $(BINARY).exe
	@echo ">> cleaned"
