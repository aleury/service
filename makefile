# Check to see if we can use ash, in Alpine images, or default to BASH
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)

# ==============================================================================
# Kind
# 	For full Kind v0.20 release notes: https://github.com/kubernetes-sigs/kind/releases/tag/v0.20.0
#
# RSA Keys
#   To generate a private/public key PEM file.
#   $ openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
#   $ openssl rsa -pubout -in private.pem -out public.pem
#
# OPA Playground
#   https://play.openpolicyagent.org/
#   https://academy.styra.com/
#   https://www.openpolicyagent.org/docs/latest/policy-reference/

# ==============================================================================
# Define dependencies

GOLANG          := golang:1.21
ALPINE          := alpine:3.18
KIND            := kindest/node:v1.27.3
POSTGRES        := postgres:15.4
VAULT           := hashicorp/vault:1.14
GRAFANA         := grafana/grafana:9.5.3
PROMETHEUS      := prom/prometheus:v2.45.0
TEMPO           := grafana/tempo:2.2.0
LOKI            := grafana/loki:2.8.3
PROMTAIL        := grafana/promtail:2.8.3
TELEPRESENCE    := datawire/ambassador-telepresence-manager:2.14.4

KIND_CLUSTER    := ardan-starter-cluster
NAMESPACE       := sales-system
APP             := sales
BASE_IMAGE_NAME := ardanlabs/service
SERVICE_NAME    := sales-api
VERSION         := 0.0.1
SERVICE_IMAGE   := $(BASE_IMAGE_NAME)/$(SERVICE_NAME):$(VERSION)

# VERSION  		:= "0.0.1-$(shell git rev-parse --short HEAD)"

run-scratch:
	go run app/scratch/main.go

run-local:
	go run app/services/sales-api/main.go \
		| go run app/tooling/logfmt/main.go -service=$(SERVICE_NAME)

run-local-help:
	go run app/services/sales-api/main.go --help

tidy:
	go mod tidy
	go mod vendor

metrics-view:
	expvarmon -ports="$(SERVICE_NAME).$(NAMESPACE).svc.cluster.local:4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

metrics-view-local:
	expvarmon -ports="localhost:4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

test-endpoint:
	curl -il $(SERVICE_NAME).$(NAMESPACE).svc.cluster.local:3000/test

test-endpoint-local:
	curl -il localhost:3000/test

test-auth-endpoint:
	curl -il -H "Authorization: Bearer ${TOKEN}" \
		$(SERVICE_NAME).$(NAMESPACE).svc.cluster.local:3000/test/auth

test-auth-endpoint-local:
	curl -il -H "Authorization: Bearer ${TOKEN}" localhost:3000/test/auth

readiness:
	curl -il $(SERVICE_NAME).$(NAMESPACE).svc.cluster.local:4000/debug/readiness

readiness-local:
	curl -il localhost:4000/debug/readiness

liveness:
	curl -il $(SERVICE_NAME).$(NAMESPACE).svc.cluster.local:4000/debug/liveness

liveness-local:
	curl -il localhost:4000/debug/liveness

pgcli:
	pgcli postgresql://postgres:postgres@database-service.$(NAMESPACE).svc.cluster.local

pgcli-local:
	pgcli postgresql://postgres:postgres@localhost

migrate:
	go run app/tooling/admin/main.go

query-users:
	@curl -s "$(SERVICE_NAME).$(NAMESPACE).svc.cluster.local:3000/users?page=1&rows=2&orderBy=name,ASC"

query-users-local:
	@curl -s localhost:3000/users?page=1&rows=2

# ==============================================================================
# Building containers

all: service

service:
	docker build \
		-f zarf/docker/dockerfile.service \
		-t $(SERVICE_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ==============================================================================
# Running from within k8s/kind

dev-adam:
	kind load docker-image $(POSTGRES) --name $(KIND_CLUSTER)

dev-tele-up:
	telepresence --context=kind-$(KIND_CLUSTER) connect

dev-up-local:
	kind create cluster \
		--image $(KIND) \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/dev/kind-config.yaml

	kubectl wait --timeout=120s --namespace=local-path-storage --for=condition=Available deployment/local-path-provisioner

	kind load docker-image $(TELEPRESENCE) --name $(KIND_CLUSTER)
	kind load docker-image $(POSTGRES) --name $(KIND_CLUSTER)

dev-up: dev-up-local
	telepresence --context=kind-$(KIND_CLUSTER) helm install
	telepresence --context=kind-$(KIND_CLUSTER) connect

dev-down-local:
	kind delete cluster --name $(KIND_CLUSTER)

dev-down:
	telepresence quit -s
	kind delete cluster --name $(KIND_CLUSTER)

dev-load:
	kind load docker-image $(SERVICE_IMAGE) --name $(KIND_CLUSTER)

dev-apply:
	kustomize build zarf/k8s/dev/database | kubectl apply -f -
	kubectl rollout status --namespace=$(NAMESPACE) --watch --timeout=120s sts/database

	kustomize build zarf/k8s/dev/sales | kubectl apply -f -
	kubectl wait pods --namespace=$(NAMESPACE) --selector app=$(APP) --for=condition=Ready

# ------------------------------------------------------------------------------

dev-logs:
	kubectl logs --namespace=$(NAMESPACE) -l app=$(APP) --all-containers=true -f --tail=100 --max-log-requests=6 \
		| go run app/tooling/logfmt/main.go -service=$(SERVICE_NAME)

dev-logs-init:
	kubectl logs --namespace=$(NAMESPACE) -l app=$(APP) -f --tail=100 -c init-migrate

dev-describe-deployment:
	kubectl describe deployment $(APP) --namespace=$(NAMESPACE)

dev-describe-sales:
	kubectl describe pod -l app=$(APP) --namespace=$(NAMESPACE)

dev-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

dev-restart:
	kubectl rollout restart deployment $(APP) --namespace=$(NAMESPACE)

# Run on code change
dev-update: all dev-load dev-restart

# Run on k8s config change
dev-update-apply: all dev-load dev-apply

# ==============================================================================
# Running tests within the local dev environment.

test-race:
	CGO_ENABLED=1 go test -race -count=1 ./...

test:
	CGO_ENABLED=0 go test -count=1 -v ./...

lint:
	CGO_ENABLED=0 go vet ./...
	staticcheck -checks=all ./...

vuln-check:
	govulncheck ./...

test-all: test lint vuln-check

test-all-race: test-race lint vuln-check
