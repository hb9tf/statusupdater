REPO="hb9tf/statusupdater"
TAG="latest"

.PHONY: build

build:
	docker buildx build --platform linux/amd64 --push -t ${REPO}:${TAG} .
