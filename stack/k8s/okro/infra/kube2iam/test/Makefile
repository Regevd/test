NS ?= kube2iam-test
ROLE_NAME ?= kube2iam-${USER}-test
POD_NAME ?= kube2iam-test-job

.PHONY: nothing
nothing:

.PHONY: init
init:
	cat assume-role-policy-document.tmpl.json | envsubst > assume-policy.json
	-aws iam create-role --role-name ${ROLE_NAME} --assume-role-policy-document file://assume-policy.json
	-kubectl create namespace ${NS}
	rm assume-policy.json

.PHONY: run-test
run-test:
	./test-kube2iam.sh

.PHONY: clean
clean:
	-aws iam delete-role --role-name ${ROLE_NAME}
	kubectl delete namespace ${NS}
