.DEFAULT_GOAL := build

##########
# CONFIG #
##########

ORG                                ?= kyverno
PACKAGE                            ?= github.com/$(ORG)/image-verification-service
KIND_IMAGE                         ?= kindest/node:v1.32.0
KIND_NAME                          ?= kind
GIT_SHA                            := $(shell git rev-parse HEAD)
GOOS                               ?= $(shell go env GOOS)
GOARCH                             ?= $(shell go env GOARCH)
REGISTRY                           ?= ghcr.io
REPO                               ?= image-verification-service
LOCAL_PLATFORM                     := linux/$(GOARCH)
KO_REGISTRY                        := ko.local
KO_PLATFORMS                       := all
KO_TAGS                            := $(GIT_SHA)
KO_CACHE                           ?= /tmp/ko-cache
CLI_BIN                            := image-verification-service

#########
# TOOLS #
#########

TOOLS_DIR                          := $(PWD)/.tools
REFERENCE_DOCS                     := $(TOOLS_DIR)/genref
REFERENCE_DOCS_VERSION             := latest
KIND                               := $(TOOLS_DIR)/kind
KIND_VERSION                       := v0.20.0
HELM                               := $(TOOLS_DIR)/helm
HELM_VERSION                       := v3.10.1
KO                                 := $(TOOLS_DIR)/ko
KO_VERSION                         := v0.14.1
TOOLS                              := $(REFERENCE_DOCS) $(KIND) $(HELM) $(KO)
PIP                                ?= pip
ifeq ($(GOOS), darwin)
SED                                := gsed
else
SED                                := sed
endif
COMMA                              := ,

.PHONY: install-tools
install-tools: $(TOOLS) ## Install tools

.PHONY: clean-tools
clean-tools: ## Remove installed tools
	@echo Clean tools... >&2
	@rm -rf $(TOOLS_DIR)

.PHONY: fmt
fmt: ## Run go fmt
	@echo Go fmt... >&2
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo Go vet... >&2
	@go vet ./...

.PHONY: ko-build
ko-build: $(KO) ## Build image (with ko)
	@echo Build image with ko... >&2
	@LDFLAGS=$(LD_FLAGS) KOCACHE=$(KO_CACHE) KO_DOCKER_REPO=$(KO_REGISTRY) \
		$(KO) build . --preserve-import-paths --tags=$(KO_TAGS) --platform=$(LOCAL_PLATFORM)

###########
# CODEGEN #
###########

CRDS_PATH                   := ${PWD}/pkg/data/crds

.PHONY: codegen-crds
codegen-crds: ## fetch CRDs
	@echo updating image verification crds... >&2
	@rm -rf $(CRDS_PATH)/*
	@curl https://raw.githubusercontent.com/kyverno/kyverno/refs/heads/main/config/crds/policies.kyverno.io/policies.kyverno.io_imageverificationpolicies.yaml > $(CRDS_PATH)/policies.kyverno.io_imageverificationpolicies.yaml

.PHONY: codegen-policies
codegen-policies:
	@echo updating image verification crds... >&2
	docker build -t $(REGISTRY)/vishal-chdhry/ivpol:latest -f ./policies/Dockerfile ./policies
	docker push $(REGISTRY)/vishal-chdhry/ivpol:latest

.PHONY: codegen-helm-crds
codegen-helm-crds: codegen-crds ## Generate helm CRDs
	@echo Generate helm crds... >&2
	@cat $(CRDS_PATH)/* \
		| $(SED) -e '1i{{- if .Values.crds.install }}' \
		| $(SED) -e '$$a{{- end }}' \
		| $(SED) -e '/^  annotations:/a \ \ \ \ {{- end }}' \
 		| $(SED) -e '/^  annotations:/a \ \ \ \ {{- toYaml . | nindent 4 }}' \
		| $(SED) -e '/^  annotations:/a \ \ \ \ {{- with .Values.crds.annotations }}' \
 		| $(SED) -e '/^  annotations:/i \ \ labels:' \
		| $(SED) -e '/^  labels:/a \ \ \ \ {{- end }}' \
 		| $(SED) -e '/^  labels:/a \ \ \ \ {{- toYaml . | nindent 4 }}' \
		| $(SED) -e '/^  labels:/a \ \ \ \ {{- with .Values.crds.labels }}' \
		| $(SED) -e '/^  labels:/a \ \ \ \ {{- include "kyverno-json.labels" . | nindent 4 }}' \
 		> ./charts/kyverno-json/templates/crds.yaml

.PHONY: codegen-helm-docs
codegen-helm-docs: ## Generate helm docs
	@echo Generate helm docs... >&2
	@docker run -v ${PWD}/charts:/work -w /work jnorwood/helm-docs:v1.11.0 -s file

.PHONY: codegen
codegen: codegen-crds codegen-helm-crds codegen-helm-docs ## Rebuild all generated code and docs

.PHONY: verify-codegen
verify-codegen: codegen ## Verify all generated code and docs are up to date
	@echo Checking codegen is up to date... >&2
	@git --no-pager diff -- .
	@echo 'If this test fails, it is because the git diff is non-empty after running "make codegen".' >&2
	@echo 'To correct this, locally run "make codegen", commit the changes, and re-run tests.' >&2
	@git diff --quiet --exit-code -- .

########
# KIND #
########

.PHONY: kind-create
kind-create: $(KIND) ## Create kind cluster
	@echo Create kind cluster... >&2
	@$(KIND) create cluster --name $(KIND_NAME) --image $(KIND_IMAGE)

.PHONY: kind-delete
kind-delete: $(KIND) ## Delete kind cluster
	@echo Delete kind cluster... >&2
	@$(KIND) delete cluster --name $(KIND_NAME)

.PHONY: kind-load
kind-load: $(KIND) ko-build ## Build image and load in kind cluster
	@echo Load image... >&2
	@$(KIND) load docker-image --name $(KIND_NAME) $(KO_REGISTRY)/$(PACKAGE):$(GIT_SHA)

.PHONY: kind-install
kind-install: $(HELM) kind-load ## Build image, load it in kind cluster and deploy helm chart
	@echo Install chart... >&2
	@$(HELM) upgrade --install image-verification-service --namespace kyverno-image-verification-service --create-namespace --wait ./charts/ \
		--set image.registry=$(KO_REGISTRY) \
		--set image.repository=$(PACKAGE) \
		--set image.tag=$(GIT_SHA)

########
# HELP #
########

.PHONY: help
help: ## Shows the available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}'
