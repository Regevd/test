OWNER ?= $(USER)

.PHONY: nothing
nothing:

.PHONY: init
init:
	find **/namespace.yaml | xargs -I {} kubectl apply -f {}
	find **/users/**.yaml | xargs -I {} kubectl apply -f {}

.PHONY: istio
istio:
	export AWS_PROFILE=arch-$(OWNER) ;\
	aws s3 cp s3://dev-okro-io-secrets/cacerts/ca-cert.pem .  ;\
	aws s3 cp s3://dev-okro-io-secrets/cacerts/ca-key.pem .  ;\
	aws s3 cp s3://dev-okro-io-secrets/cacerts/root-cert.pem .  ;\
	aws s3 cp s3://dev-okro-io-secrets/cacerts/cert-chain.pem .
	kubectl create -n istio-system secret generic cacerts \
		--from-file=ca-cert.pem --from-file=ca-key.pem \
		--from-file=cert-chain.pem --from-file=root-cert.pem
	kubectl create -n istio-system secret tls istio-ingressgateway-certs \
		--key ca-key.pem --cert ca-cert.pem
	rm -f *.pem
	kubectl apply -f istio/istio-all-in-one.yaml

.PHONY: istio-rbac
istio-rbac:
	kubectl apply -f istio/istio-rbac.yaml

.PHONY: istio-ns
istio-ns:
	find **/istio-ns/**/*.yaml | xargs -I {} kubectl apply -f {}

.PHONY: infra
infra:
	find **/infra/**/*.yaml | xargs -I {} kubectl apply -f {}

.PHONY: services
services:
	find **/services/**/*.yaml | xargs -I {} kubectl apply -f {}

.PHONY: services-down
services-down:
	find **/services/**/*.yaml | xargs -I {} kubectl delete -f {}

.PHONY: logging
logging:
	export AWS_PROFILE=arch-$(OWNER) ;\
	aws s3 cp s3://dev-okro-io-secrets/logz.io/general-token general-token ;\
	aws s3 cp s3://dev-okro-io-secrets/logz.io/team1-token team1-token
	kubectl create secret -n okro-system generic logzio-tokens \
		--from-file=general-token --from-file=team1-token
	rm -f *-token
	kubectl apply -f okro/addons/logging

.PHONY: monitoring
monitoring:
	export AWS_PROFILE=arch-$(OWNER) ;\
	aws s3 cp s3://dev-okro-io-secrets/alertmanager-config.yaml alertmanager.yaml
	kubectl create secret -n okro-system generic alertmanager-main \
		--from-file=alertmanager.yaml
	kubectl create secret -n team1-system generic alertmanager-main \
		--from-file=alertmanager.yaml
	#kubectl create secret -n team2-system generic alertmanager-main \
	#	--from-file=alertmanager.yaml
	rm -f alertmanager.yaml
	kubectl apply -f monitoring/operator
	kubectl apply -f monitoring/okro-system
	kubectl apply -f monitoring/team1
	#kubectl apply -f monitoring/team2

.PHONY: hpa
hpa:
	okro/addons/hpa/create-adapter-cert.sh
	kubectl apply -f okro/addons/hpa/custom-metrics-apiserver.yaml
	kubectl apply -f okro/addons/hpa/hpa-v2.yaml
