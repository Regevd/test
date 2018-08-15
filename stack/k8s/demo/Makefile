OWNER ?= $(USER)
NAME = k8s.$(OWNER).dev.okro.io
KOPS_STATE_STORE=s3://dev-okro-io

.PHONY: mtls-test-up
mtls-test-up:
	kubectl apply -f mtls-test-namespaces.yaml
	kubectl apply -n test1 -f ../sleep.yaml
	kubectl apply -n test1-injected -f ../sleep.yaml

.PHONY: mtls-test-down
mtls-test-down:
	kubectl delete -f mtls-test-namespaces.yaml

.PHONY: mtls-permissive
mtls-permissive:
	kubectl apply -f mtls-permissive.yaml

.PHONY: mtls-strict
mtls-strict:
	kubectl apply -f mtls-strict.yaml

.PHONY: traffic-init
traffic-init:
	kubectl apply -f traffic-init.yaml

.PHONY: traffic-mirror
traffic-mirror:
	kubectl apply -f traffic-mirror.yaml

.PHONY: traffic-canary
traffic-canary:
	kubectl apply -f traffic-canary.yaml

.PHONY: traffic-shift
traffic-shift:
	kubectl apply -f traffic-shift.yaml

.PHONY: traffic-fault-injection
traffic-fault-injection:
	kubectl apply -f traffic-fault-injection.yaml

.PHONY: kops-roll
kops-roll:
	export AWS_PROFILE=arch-$(OWNER) ;\
    export KOPS_STATE_STORE=$(KOPS_STATE_STORE) ;\
    kops get cluster $(NAME) -o yaml | sed -e 's/1\.9\../1.9.8/' > kops/kops.yaml ;\
    kops replace -f kops/kops.yaml ;\
    kops update cluster $(NAME) --yes ;\
    kops rolling-update cluster $(NAME) --yes
