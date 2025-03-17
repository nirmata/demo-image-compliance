.DEFAULT_GOAL := build

##########
# CONFIG #
##########

ORG                                ?= nirmata
PACKAGE                            ?= github.com/$(ORG)/demo-image-compliance
KIND_IMAGE                         ?= kindest/node:v1.32.0
KIND_NAME                          ?= verify-images
GIT_SHA                            := $(shell git rev-parse HEAD)
GOOS                               ?= $(shell go env GOOS)
GOARCH                             ?= $(shell go env GOARCH)
REGISTRY                           ?= ghcr.io
REPO                               ?= $(ORG)/demo-image-compliance
LOCAL_PLATFORM                     := linux/$(GOARCH)
KO_REGISTRY                        := ko.local
KO_PLATFORMS                       := all
KO_TAGS                            := $(GIT_SHA)
KO_CACHE                           ?= /tmp/ko-cache
CLI_BIN                            := image-compliance

#########
# TOOLS #
#########

TOOLS_DIR                          := $(PWD)/.tools
REFERENCE_DOCS                     := $(TOOLS_DIR)/genref
REFERENCE_DOCS_VERSION             := latest
KIND                               := $(TOOLS_DIR)/kind
KIND_VERSION                       := v0.23.0
HELM                               := $(TOOLS_DIR)/helm
HELM_VERSION                       := v3.12.3
HELM_DOCS                          ?= $(TOOLS_DIR)/helm-docs
HELM_DOCS_VERSION                  ?= v1.11.0
KO                                 := $(TOOLS_DIR)/ko
KO_VERSION                         := v0.17.1
TOOLS                              := $(KIND) $(HELM) $(HELM_DOCS) $(KO)
PIP                                ?= pip
ifeq ($(GOOS), darwin)
SED                                := gsed
else
SED                                := sed
endif
COMMA                              := ,

$(KIND):
	@echo Install kind... >&2
	@GOBIN=$(TOOLS_DIR) go install sigs.k8s.io/kind@$(KIND_VERSION)

$(HELM):
	@echo Install helm... >&2
	@GOBIN=$(TOOLS_DIR) go install helm.sh/helm/v3/cmd/helm@$(HELM_VERSION)

$(HELM_DOCS):
	@echo Install helm-docs... >&2
	@GOBIN=$(TOOLS_DIR) go install github.com/norwoodj/helm-docs/cmd/helm-docs@$(HELM_DOCS_VERSION)

$(KO):
	@echo Install ko... >&2
	@GOBIN=$(TOOLS_DIR) go install github.com/google/ko@$(KO_VERSION)

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

.PHONY: test-all
test-all: test-clean test-unit ## Clean tests cache then run unit tests

.PHONY: test-clean
test-clean: ## Clean tests cache
	@echo Clean test cache... >&2
	@go clean -testcache

.PHONY: test-unit
test-unit: ## Run tests
	@echo Running tests... >&2
	@go test ./... -race -coverprofile=coverage.out -v

# BUILD #
#########

REGISTRY_USERNAME   ?= dummy
LOCAL_PLATFORM := linux/$(GOARCH)
KO_REPO_LOCAL ?= ko.local

ko-login: $(KO)
	@$(KO) login $(REGISTRY) --username $(REGISTRY_USERNAME) --password $(REGISTRY_PASSWORD)

.PHONY: ko-build-local
ko-build-local: $(KO) ## Build image (with ko)
	@echo Build image with ko... >&2
	KO_DOCKER_REPO=$(KO_REPO_LOCAL) $(KO) build --preserve-import-paths --tags=$(GIT_SHA) --platform=$(LOCAL_PLATFORM) ./cmd
	
.PHONY: ko-publish-local
ko-publish-local: $(KO) ## Build and publish the admission controller container image.
	KO_DOCKER_REPO=$(KO_REPO_LOCAL) $(KO) build --bare cmd/main.go

KO_REPO ?= ghcr.io/$(ORG)/demo-image-compliance

.PHONY: ko-build
ko-build: $(KO) ko-login ## Build image (with ko)
	@echo Build image with ko... >&2
	KO_DOCKER_REPO=$(KO_REPO) $(KO) build --preserve-import-paths --tags=$(GIT_SHA) --platform=$(LOCAL_PLATFORM) ./cmd
	
.PHONY: ko-publish
ko-publish: $(KO) ko-login ## Build and publish the admission controller container image.
	KO_DOCKER_REPO=$(KO_REPO) $(KO) build --bare cmd/main.go

.PHONY: publish-policies
publish-policies:
	@echo updating image verification policies... >&2
	docker build -t $(REGISTRY)/nirmata/image-compliance-policies:block-critical-vulnerabilities -f ./policies/Dockerfile --build-arg POLICY_EXPRESSION=critical.yaml ./policies
	docker build -t $(REGISTRY)/nirmata/image-compliance-policies:block-high-and-critical-vulnerabilities -f ./policies/Dockerfile --build-arg POLICY_EXPRESSION=high.yaml ./policies
	docker push $(REGISTRY)/nirmata/image-compliance-policies:block-critical-vulnerabilities
	docker push $(REGISTRY)/nirmata/image-compliance-policies:block-high-and-critical-vulnerabilities

###########
# CODEGEN #
###########

CRDS_PATH                   := ${PWD}/pkg/data/crds

.PHONY: codegen-crds
codegen-crds: ## fetch CRDs
	@echo updating image verification crds... >&2
	@rm -rf $(CRDS_PATH)/*
	@curl https://raw.githubusercontent.com/kyverno/kyverno/refs/heads/main/config/crds/policies.kyverno.io/policies.kyverno.io_imageverificationpolicies.yaml > $(CRDS_PATH)/policies.kyverno.io_imageverificationpolicies.yaml


.PHONY: codegen-helm-docs
codegen-helm-docs: ## Generate helm docs
	@echo Generate helm docs... >&2
	@docker run -v ${PWD}/charts:/work -w /work jnorwood/helm-docs:v1.11.0 -s file

.PHONY: codegen-manifest-install
codegen-manifest-install: ## Generate install manifest
	$(HELM) template nirmata-image-compliance --namespace nirmata charts/image-compliance \
	--set image.tag=latest \
	> config/install.yaml

.PHONY: codegen
codegen: codegen-crds codegen-helm-docs codegen-manifest-install ## Rebuild all generated code and docs

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
kind-load: $(KIND) ko-build-local ## Build image and load in kind cluster
	@echo Load image... >&2
	@$(KIND) load docker-image --name $(KIND_NAME) $(KO_REPO_LOCAL)/$(PACKAGE)/cmd:$(GIT_SHA)

.PHONY: kind-install
kind-install: $(HELM) kind-load ## Build image, load it in kind cluster and deploy helm chart
	@echo Install chart... >&2
	@$(HELM) upgrade --install nirmata-image-compliance --namespace nirmata \
		--create-namespace --wait ./charts/image-compliance \
		--set image.registry=$(KO_REPO_LOCAL) \
		--set image.repository=$(PACKAGE)/cmd \
		--set image.tag=$(GIT_SHA)

########
# HELP #
########

.PHONY: help
help: ## Shows the available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}'
