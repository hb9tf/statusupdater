REPO="hb9tf/statusupdater"
TAG="latest"

.PHONY: build, push

build_image:
	docker build -t ${REPO}:${TAG} .

push:
	docker push ${REPO}:${TAG}

build: build_image, push
