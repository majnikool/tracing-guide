IMG_BASE = mariomac/ds-example
REGISTRY_BASE = dev-docker-registry.xx.com/trace-test

.PHONY: build-backend build-frontend build-worker build-loadgen build-all
build-backend:
	docker build -t $(IMG_BASE):backend backend/.

build-frontend:
	docker build -t $(IMG_BASE):frontend frontend/.

build-worker:
	docker build -t $(IMG_BASE):worker worker/.

build-loadgen:
	docker build -t $(IMG_BASE):loadgen loadgen/.

build-all: build-backend build-frontend build-worker build-loadgen

.PHONY: tag-push-backend tag-push-frontend tag-push-worker tag-push-loadgen tag-push-all
tag-push-backend:
	docker tag $(IMG_BASE):backend $(REGISTRY_BASE)/backend
	docker push $(REGISTRY_BASE)/backend

tag-push-frontend:
	docker tag $(IMG_BASE):frontend $(REGISTRY_BASE)/frontend
	docker push $(REGISTRY_BASE)/frontend

tag-push-worker:
	docker tag $(IMG_BASE):worker $(REGISTRY_BASE)/worker
	docker push $(REGISTRY_BASE)/worker

tag-push-loadgen:
	docker tag $(IMG_BASE):loadgen $(REGISTRY_BASE)/loadgen
	docker push $(REGISTRY_BASE)/loadgen

tag-push-all: tag-push-backend tag-push-frontend tag-push-worker tag-push-loadgen
