SHELL=/bin/bash
.SHELLFLAGS=-euo pipefail -c

KIND_KUBECONFIG:=.kind-kubeconfig
KIND_CLUSTER_NAME=kustomize-plugin-packagedeployment
MODULE:=github.com/erdii/kustomize-plugin-packagedeployment

PACKAGE_OPERATOR_KUSTOMIZATION:=~/projects/github.com/thetechnick/package-operator/config/deploy
CONTROLLER_GEN_VERSION:=v0.5.0

# Build Flags
export CGO_ENABLED:=0
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
SHORT_SHA=$(shell git rev-parse --short HEAD)
VERSION?=$(shell echo ${BRANCH} | tr / -)-${SHORT_SHA}
BUILD_DATE=$(shell date +%s)
LD_FLAGS=-X $(MODULE)/internal/version.Version=$(VERSION) \
                        -X $(MODULE)/internal/version.Branch=$(BRANCH) \
                        -X $(MODULE)/internal/version.Commit=$(SHORT_SHA) \
                        -X $(MODULE)/internal/version.BuildDate=$(BUILD_DATE)
UNAME_OS:=$(shell uname -s)
UNAME_ARCH:=$(shell uname -m)

# PATH/Bin
DEPENDENCIES:=.cache/dependencies/$(UNAME_OS)/$(UNAME_ARCH)
export GOBIN?=$(abspath .cache/dependencies/bin)
export PATH:=$(GOBIN):$(PATH)

clean:
	rm -rf ".cache"
.PHONY: clean

all: \
	.cache/plugins/packages.k8s.erdii.net/v1alpha1/packagetransformer/PackageTransformer

.cache/plugins/packages.k8s.erdii.net/v1alpha1/packagetransformer/PackageTransformer: GOARGS = GOOS=linux GOARCH=amd64
.cache/plugins/packages.k8s.erdii.net/v1alpha1/packagetransformer/PackageTransformer: generate FORCE
	@echo -e -n "compiling cmd/v1alpha1/packagetransformer...\n  "
	$(GOARGS) go build \
		-ldflags "$(LD_FLAGS)" \
		-o .cache/plugins/packages.k8s.erdii.net/v1alpha1/packagetransformer/PackageTransformer \
		cmd/v1alpha1/packagetransformer/main.go
	@echo

FORCE:

# prints the version as used by build commands.
version:
	@echo $(VERSION)
.PHONY: version

# setup controller-gen
CONTROLLER_GEN:=$(DEPENDENCIES)/controller-gen/$(CONTROLLER_GEN_VERSION)
$(CONTROLLER_GEN):
	@echo "installing controller-gen $(CONTROLLER_GEN_VERSION)..."
	$(eval CONTROLLER_GEN_TMP := $(shell mktemp -d))
	@(cd "$(CONTROLLER_GEN_TMP)"; \
		go mod init tmp; \
		go get "sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION)"; \
	) 2>&1 | sed 's/^/  /'
	@rm -rf "$(CONTROLLER_GEN_TMP)" "$(dir $(CONTROLLER_GEN))"
	@mkdir -p "$(dir $(CONTROLLER_GEN))"
	@touch "$(CONTROLLER_GEN)"
	@echo

# Generate api object code
generate: $(CONTROLLER_GEN)
	@echo
	@echo "generating code..."
	@controller-gen object paths=./apis/... 2>&1 | sed 's/^/  /'
	@echo
.PHONY: generate

create-kind-cluster:
	@echo "creating kind cluster $(KIND_CLUSTER_NAME)..."
	@mkdir -p .cache/e2e
	@(source hack/determine-container-runtime.sh; \
		$$KIND_COMMAND create cluster \
			--config="kind.yaml" \
			--kubeconfig=$(KIND_KUBECONFIG) \
			--name="$(KIND_CLUSTER_NAME)"; \
		sudo chown $$USER: $(KIND_KUBECONFIG); \
		echo)
	@(source .envrc; \
		kubectl apply -k $(PACKAGE_OPERATOR_KUSTOMIZATION); \
		echo)
.PHONY: create-kind-cluster

delete-kind-cluster:
	@echo "deleting kind cluster $(KIND_CLUSTER_NAME)..."
	@(source hack/determine-container-runtime.sh; \
		$$KIND_COMMAND delete cluster \
			--kubeconfig="$(KIND_KUBECONFIG)" \
			--name "$(KIND_CLUSTER_NAME)"; \
		rm -rf "$(KIND_KUBECONFIG)"; \
		echo; \
	)
.PHONY: delete-kind-cluster
