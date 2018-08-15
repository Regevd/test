VERSION?=v1.0.0
STACK=stack.yaml
REGISTRY=520887410018.dkr.ecr.eu-west-1.amazonaws.com/okro/hellod
OWNER?=$(USER)

nothing:

.PHONY: build-images
build-images:
	docker build -t dude:$(VERSION) --build-arg svc=dude .
	docker build -t helloer:$(VERSION) --build-arg svc=helloer .
	docker build -t echoer:$(VERSION) --build-arg svc=echoer .
	docker build -t watcher:$(VERSION) --build-arg svc=watcher .
	docker build -t reflector:$(VERSION) --build-arg svc=reflector .

.PHONY: ecr-login
ecr-login:
	export AWS_PROFILE=arch-$(OWNER); \
	$$(aws ecr get-login --no-include-email --region eu-west-1)

.PHONY: push-images
push-images:
	docker tag dude:$(VERSION) $(REGISTRY)/dude:$(VERSION)
	docker push $(REGISTRY)/dude:$(VERSION)
	docker tag helloer:$(VERSION) $(REGISTRY)/helloer:$(VERSION)
	docker push $(REGISTRY)/helloer:$(VERSION)
	docker tag echoer:$(VERSION) $(REGISTRY)/echoer:$(VERSION)
	docker push $(REGISTRY)/echoer:$(VERSION)
	docker tag watcher:$(VERSION) $(REGISTRY)/watcher:$(VERSION)
	docker push $(REGISTRY)/watcher:$(VERSION)
	docker tag reflector:$(VERSION) $(REGISTRY)/reflector:$(VERSION)
	docker push $(REGISTRY)/reflector:$(VERSION)

.PHONY: patch-image
patch-image:
	docker build -t $(SVC):$(VERSION) --build-arg svc=$(SVC) .
	docker tag $(SVC):$(VERSION) $(REGISTRY)/$(SVC):$(VERSION)
	docker push $(REGISTRY)/$(SVC):$(VERSION)

.PHONY: deps
deps:
	HELLOD_VERSION=$(VERSION) \
	docker-compose -f $(STACK) up zk kafka jaeger

.PHONY: run
run:
	HELLOD_VERSION=$(VERSION) \
	docker-compose -f $(STACK) up dude helloer echoer watcher reflector

.PHONY: clean
clean:
	HELLOD_VERSION=$(VERSION) \
	docker-compose -f $(STACK) down
