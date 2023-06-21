VGO=go # Set to vgo if building in Go 1.10
BUILD_VERSION = 0.2
BINARY_NAME = icon
OUTPUT_PATH = ./build
DEVELOP_TMP_DIR = ./vault_temp
VAULT_VERSION = 1.11.2
ARCH := $(shell arch)
UNAME := $(shell uname)
LOWER_UNAME := `echo $(UNAME) | tr A-Z a-z`
DEPLOY_SERVER = 100.106.142.90

GO_OS=$(shell go env GOOS)
GO_ARCH=$(shell go env GOARCH)

ifeq ("$(ARCH)", "x86_64")
	ARCH = amd64
endif

VAULT_BIN := vault_$(GO_OS)_$(ARCH)

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    ECHO_OPTION = "-e"
    SED_OPTION =
    SHASUM_CMD = sha256sum
endif
ifeq ($(UNAME_S),Darwin)
    ECHO_OPTION = ""
	SED_OPTION = ''
	SHASUM_CMD = shasum -a 256
endif

define colorecho
      @tput setaf 6
      @echo $1
      @tput sgr0
endef


# define standard colors
ifneq (,$(findstring xterm,${TERM}))
	BLACK        := $(shell tput -Txterm setaf 0)
	RED          := $(shell tput -Txterm setaf 1)
	GREEN        := $(shell tput -Txterm setaf 2)
	YELLOW       := $(shell tput -Txterm setaf 3)
	LIGHTPURPLE  := $(shell tput -Txterm setaf 4)
	PURPLE       := $(shell tput -Txterm setaf 5)
	BLUE         := $(shell tput -Txterm setaf 6)
	WHITE        := $(shell tput -Txterm setaf 7)
	RESET := $(shell tput -Txterm sgr0)
else
	BLACK        := ""
	RED          := ""
	GREEN        := ""
	YELLOW       := ""
	LIGHTPURPLE  := ""
	PURPLE       := ""
	BLUE         := ""
	WHITE        := ""
	RESET        := ""
endif


SHA3SUM := $(shell find . -maxdepth 1  -name '$(BINARY_NAME)*' -exec $(SHASUM_CMD) {} \;)
SRC_GOFILES := $(shell find . -name '*.go' -print)
.DELETE_ON_ERROR:

word-hyphen = $(word $2,$(subst -, ,$1))

all: build test shasum

test: deps
		$(VGO) test  ./... -cover -coverprofile=coverage.txt -covermode=atomic

iconsign: ${SRC_GOFILES}
		@echo " ${BLUE}[BUILD] ${GO_OS}_${GO_ARCH} ${RESET}"
		CGO_ENABLED=0 $(VGO) build -o ${OUTPUT_PATH}/${BINARY_NAME} -ldflags "-X main.buildDate=`date -u +\"%Y-%m-%dT%H:%M:%SZ\"` -X main.buildVersion=$(BUILD_VERSION)" -tags=prod -v

build:  build-$(GO_OS)-$(GO_ARCH)

all-build: test build-darwin-arm64 build-darwin-amd64 build-linux-arm64 build-linux-amd64

build-%:
# it's not work in linux, but osx works well
#override os=$(call word-hyphen,$*,1)
#override arch=$(call word-hyphen,$*,2)

	$(eval os := $(call word-hyphen,$*,1))
	$(eval arch := $(call word-hyphen,$*,2))
	@echo "${os}, ${arch}"
	@$(call colorecho, "[${@}]")
	@\
	if [ "$(arch)" == "" ]; then\
		 export arch="${ARCH}";\
	fi ;\
	export BIN_FILE="${BINARY_NAME}_${os}_${arch}" ;\
	echo "${BLUE}[BUILD] ${@}  / OS_ARCH=${ARCH}, os=${os} arch=${arch}, BIN=$${BIN_FILE} ${RESET}" ;\
	CGO_ENABLED=0 GOARCH=$${arch} GOOS=${os} $(VGO) build -o ${OUTPUT_PATH}/$${BIN_FILE} \
		-ldflags "-X main.buildDate=`date -u +\"%Y-%m-%dT%H:%M:%SZ\"` -X main.buildVersion=$(BUILD_VERSION)" -tags=prod -v || exit 1;\
	$(MAKE) shasum-$${BIN_FILE}


shasum:
	@find ${OUTPUT_PATH} -maxdepth 1  -name '$(BINARY_NAME)*' -exec ls -lT {}\; -exec $(SHASUM_CMD) {} \;

shasum-%:
	@ \
	export OUTPUT_FILE=`find $(OUTPUT_PATH) -maxdepth 1  -name '${patsubst shasum-%,%,$(@)}' -exec $(SHASUM_CMD) {} \;` ;\
	echo "${BLUE}[BUILD] $${OUTPUT_FILE} ${RESET}" ;\

	@echo ""
	export SHASUM=$(shell $(SHASUM_CMD) '${OUTPUT_PATH}/${patsubst shasum-%,%,$(@)}' | cut -d " " -f1)

clean:
		$(VGO) clean
		rm -f $(OUTPUT_PATH)/$(BINARY_NAME)*
		rm -f ${BINARY_NAME}
		rm -rf ${DEVELOP_TMP_DIR}/plugin
deps:
		$(VGO) get

#deploy_dev: build-linux-amd64 test

deploy_dev: build-linux-amd64 test deploy_file

deploy_file: shasum
	scp build/icon_linux_amd64 root@$(DEPLOY_SERVER):/app/vault/vault_server/build/;
	#ssh root@$(DEPLOY_SERVER) 'docker exec -it vault-local vault plugin deregister iconsign';



docker_dev: build-linux-${ARCH} test
	cd docker && docker-compose -f docker-compose-local.dev.yml up -d;
	docker exec -e PLUGIN_NAME=$(PLUGIN_NAME) -it vault-dev /script/enable_plugin_docker.sh

dev: setup_development_environment clean
	cd $(DEVELOP_TMP_DIR);
	VAULT_DIR=$(DEVELOP_TMP_DIR)  VAULT_BIN="./$(VAULT_BIN)" ./script/start_vault.sh


reload:
	./script/start_vault.sh

setup_development_environment:
	@ echo "* Set up environment for development"
	@if test ! -d $(DEVELOP_TMP_DIR) ; \
		then mkdir -p $(DEVELOP_TMP_DIR)/plugin ; \
	fi
	@echo "OS: $(UNAME) $(ARCH) GO_ARCH: $(GO_OS) $(GO_ARCH), vault_bin = $(DEVELOP_TMP_DIR)/$(VAULT_BIN)"
#	@test ! -d $(DEVELOP_TMP_DIR) || mkdir $(DEVELOP_TMP_DIR)

	@if test ! -f $(DEVELOP_TMP_DIR)/$(VAULT_BIN) ; \
		then echo "* Download vault binary " ; \
	 	curl -o $(DEVELOP_TMP_DIR)/$(VAULT_BIN).zip https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_$(GO_OS)_$(ARCH).zip ;\
	 	unzip -o $(DEVELOP_TMP_DIR)/$(VAULT_BIN).zip -d $(DEVELOP_TMP_DIR)/ ;\
		mv $(DEVELOP_TMP_DIR)/vault $(DEVELOP_TMP_DIR)/$(VAULT_BIN)  ;\
	fi


enable_plugin:
	VAULT_DIR=$(DEVELOP_TMP_DIR)  VAULT_BIN="./$(VAULT_BIN)" ./script/enable_plugin.sh
