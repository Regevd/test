VERSION?=v1.2.2
IMAGE=fluentd
REGISTRY=520887410018.dkr.ecr.eu-west-1.amazonaws.com/okro/fluentd
OWNER?=$(USER)

nothing:

.PHONY: build-image
build-image:
	docker build -t $(IMAGE):$(VERSION) .

.PHONY: ecr-login
ecr-login:
	export AWS_PROFILE=arch-$(OWNER); \
	$$(aws ecr get-login --no-include-email --region eu-west-1)

.PHONY: push-image
push-image:
	docker tag $(IMAGE):$(VERSION) $(REGISTRY):$(VERSION)
	docker push $(REGISTRY):$(VERSION)

