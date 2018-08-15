.EXPORT_ALL_VARIABLES:

# kops vars
OWNER?=$(USER)
AWS_PROFILE=arch-$(OWNER)
NAME=k8s.$(OWNER).dev.okro.io
KOPS_STATE_STORE=s3://dev-okro-io

# flags
KOPS_MASTER_TYPE?=t2.medium
KOPS_MASTER_MAX_PRICE?=0.02
KOPS_NODE_TYPE?=t2.medium
KOPS_NODE_MAX_PRICE?=0.02
KOPS_NODES_PROMETHEUS?=false
KOPS_BASTION_TYPE?=t2.micro
KOPS_BASTION_MAX_PRICE?=0.008

#
# terraform
#

.PHONY: tf-init
tf-init:
	cd tf/envs/$(OWNER) ;\
	terraform init

.PHONY: tf-plan
tf-plan:
	cd tf/envs/$(OWNER) ;\
	terraform plan

.PHONY: tf-apply
tf-apply:
	cd tf/envs/$(OWNER) ;\
	terraform apply -auto-approve

.PHONY: tf-kops
tf-kops:
	cd tf/envs/$(OWNER) ;\
	terraform output -json | tf2kops > tf2kops.yaml

.PHONY: tf-destroy
tf-destroy:
	cd tf/envs/$(OWNER) ;\
	terraform destroy -auto-approve

#
# kops
#
.PHONY: kops-clean
kops-clean:
	rm -rf kops/generated

.PHONY: kops-render
kops-render: tf-kops
	mkdir kops/generated
	mkdir kops/generated/instancegroups

	export AESCBC_SECRET=$$(head -c 32 /dev/urandom | base64) ;\
	envsubst '$$AESCBC_SECRET' < kops/templates/encryptionconfig.tmpl.yaml > kops/generated/encryptionconfig.yaml

	export TF2KOPS=$$(cat tf/envs/$(OWNER)/tf2kops.yaml) ;\
	export AUDIT_POLICY=$$(cat kops/audit-policy.yaml | base64 | tr -d "\n=") ;\
	envsubst '$$NAME $$TF2KOPS $$AUDIT_POLICY' < kops/templates/cluster.tmpl.yaml > kops/generated/cluster.yaml ;\

	envsubst '$$NAME $$KOPS_MASTER_TYPE $$KOPS_MASTER_MAX_PRICE' < kops/templates/instancegroups/masters.tmpl.yaml > kops/generated/instancegroups/masters.yaml
	envsubst '$$NAME $$KOPS_NODE_TYPE $$KOPS_NODE_MAX_PRICE' < kops/templates/instancegroups/nodes.tmpl.yaml > kops/generated/instancegroups/nodes.yaml
	envsubst '$$NAME $$KOPS_BASTION_TYPE $$KOPS_BASTION_MAX_PRICE' < kops/templates/instancegroups/bastions.tmpl.yaml > kops/generated/instancegroups/bastions.yaml
ifeq ($(KOPS_NODES_PROMETHEUS),true)
	envsubst '$$NAME $$KOPS_NODE_TYPE $$KOPS_NODE_MAX_PRICE' < kops/templates/instancegroups/nodes.prometheus.tmpl.yaml > kops/generated/instancegroups/nodes.prometheus.yaml
endif

.PHONY: kops-up
kops-up: kops-clean kops-render
	kops create -f kops/generated/cluster.yaml
	find kops/generated/instancegroups -type f | xargs -I {} kops create -f {}
	kops create secret sshpublickey admin -i ~/.ssh/dev-okro-io-admin.pub --name=$(NAME)
	kops create secret encryptionconfig -f kops/generated/encryptionconfig.yaml --name=$(NAME)
	kops update cluster $(NAME) --yes
	@echo -------------------------------------------------------
	@echo -  k8s should be up in about 5 minutes,'                '-
	@echo -  run \'make kops-validate\' to validate the cluster.'   '-
	@echo -  source kops/env.sh to run manual kops commands'      '-
	@echo -------------------------------------------------------

.PHONY: kops-validate
kops-validate:
	kops validate cluster $(NAME)

.PHONY: kops-kubecfg
kops-kubecfg:
	kops export kubecfg $(NAME)

.PHONY: kops-down
kops-down:
	kops delete cluster $(NAME) --yes

.PHONY: kops-nuke
kops-nuke:
	kops get clusters | tail -n+2 | cut -f1 | xargs -I {} kops delete cluster {} --yes

#
# post-install
#

.PHONY: ssh-config
ssh-config:
	aws s3 cp s3://dev-okro-io-secrets/dev-okro-io-admin ~/.ssh/
	aws s3 cp s3://dev-okro-io-secrets/dev-okro-io-admin.pub ~/.ssh/
	chmod 400 ~/.ssh/dev-okro-io-admin
	chmod 400 ~/.ssh/dev-okro-io-admin.pub
	export CIDR=$$(grep 'cidr.*=' tf/envs/$(OWNER)/main.tf | grep -oe '[[:digit:]]\+\.[[:digit:]]\+' | head -1) ;\
	envsubst '$$NAME $$CIDR' < hack/templates/sshconfig.tmpl > ~/.ssh/config

.PHONY: oidc-login
oidc-login:
	export S3_CA_URL="s3://dev-okro-io/$(NAME)/pki/issued/ca" ;\
	export CA_FILE_NAME=$$(aws s3 ls $$S3_CA_URL/ | grep .crt | awk '{print $$NF}') ;\
	export CA_DATA=$$(aws s3 cp $$S3_CA_URL/$$CA_FILE_NAME - | base64 | tr -d '\n') ;\
	konfigurator \
		-c 0oafiqm6by9qgsCE40h7 \
		-u https://dev-405616.oktapreview.com/oauth2/default \
		-a "api.$(NAME)" -s $$CA_DATA \
		> OIDC_KUBE_CONFIG

