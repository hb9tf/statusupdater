REPO="hb9tf/statusupdater"
TAG="latest"

.PHONY: build

build:
	docker build -t ${REPO}:${TAG} .
	docker push ${REPO}:${TAG}
