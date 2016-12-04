all: help

# GIT_PORT = 10080
# PARENTDIR = $(shell pwd)/..
# DEPS = webfw goodl-lib goi18n redistore

help:
	@echo ""
	@echo "-- Help Menu"
	@echo ""
	@echo "   please check makefile... :p"

start-redis:
	docker run -d \
		--name goodl_redis \
		--restart=always \
		-v /srv/goodl-alpha/redis:/data \
		redis:latest

build-dev:
	@rm -rf Dockerfile
	@cp ./scripts/Dockerfile.dev ./Dockerfile
	@docker build --tag=odl_go/goodl-dev .

update-deps:
	@echo "... checking out all develop branches...\n"
	cd $(shell pwd)/../webfw && git checkout develop && git pull origin develop
	cd $(shell pwd)/../goodl-lib && git checkout develop && git pull origin develop
	cd $(shell pwd)/../odl-geocoder && git checkout develop && git pull origin develop
	cd $(shell pwd)/../goi18n && git checkout develop && git pull origin develop
	cd $(shell pwd)/../redistore && git checkout develop && git pull origin develop

run-dev:
	@docker run -it --rm \
		--name goodl-dev \
		--link goodl_redis:redis \
		-e ENVIRONMENT=development \
		-e SMTP_HOST=mail.opendriverslog.de \
		-e SMTP_PORT=587 \
		-v "$(shell pwd)/../goodl-databases/databases":/databases \
		-v "$(shell pwd)":/go/src/github.com/OpenDriversLog/goodl \
		-v "$(shell pwd)/../webfw":/go/src/github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/webfw \
		-v "$(shell pwd)/../goodl-lib":/go/src/github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/goodl-lib \
		-v "$(shell pwd)/../goodl-lib/test_integration":/go/src/github.com/OpenDriversLog/goodl-lib/test_integration \
		-v "$(shell pwd)/../goi18n":/go/src/github.com/OpenDriversLog/goodl/vendor/github.com/Compufreak345/go-i18n \
		-v "$(shell pwd)/../redistore":/go/src/github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/redistore \
		-v "$(shell pwd)/../odl-geocoder":/go/src/github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/odl-geocoder \
		-p 4000:4000 \
		-p 6060:6060 \
		odl_go/goodl-dev

dev-run: run-dev

dev-bash:
	@docker exec -it \
		goodl-dev \
		/bin/bash

dev-test:
	@echo "Starting Selenium Grid & goodl dev container with links...."
	@echo "You'll see selenium & goodl logs here..."
	@echo "to start acceptance test suide run: make dev-test-compose-ff\n"
	@echo "  or  \n"
	@echo "make dev-test-compose-chrome\n\n"
	@docker-compose --file scripts/compose-acceptance-dev.yml -p selhub up --force-recreate
# force-recreate needed because of https://github.com/SeleniumHQ/docker-selenium/issues/91
dev-test-compose-ff:
	docker exec -it \
		selhub_goodlserver_1 \
		ginkgo -r /go/src/github.com/OpenDriversLog/goodl/tests/acceptance -- -odl.browser=firefox -odl.lang=de

dev-test-compose-chrome:
	docker exec -it \
		selhub_goodlserver_1 \
		ginkgo -r /go/src/github.com/OpenDriversLog/goodl/tests/acceptance -- -odl.browser=chrome -odl.lang=de

# this is for running integration tests on dev-machine using dev-image
dev-test-integration:
	@docker run -it --rm \
		--name goodl_test_integration \
		--link goodl_redis:redis \
		-e ENVIRONMENT=test \
		-e SMTP_HOST=mail.opendriverslog.de \
		-e SMTP_PORT=587 \
		-v "$(shell pwd)":/go/src/github.com/OpenDriversLog/goodl:ro \
		-v "$(shell pwd)/../webfw":/go/src/github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/webfw \
		-v "$(shell pwd)/../goodl-lib":/go/src/github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/goodl-lib \
		-v "$(shell pwd)/../goi18n":/go/src/github.com/OpenDriversLog/goodl/vendor/github.com/Compufreak345/go-i18n \
		-v "$(shell pwd)/../redistore":/go/src/github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/redistore \
		-p 4004:4000 \
		-v /etc/localtime:/etc/localtime:ro \
		odl_go/goodl-dev \
		ginkgo /go/src/github.com/OpenDriversLog/goodl/tests/integration

dev-test-lib-all: dev-test-lib-datapolish dev-test-lib-addressManager dev-test-lib-dbMan dev-test-lib-tripMan

dev-test-lib-datapolish:
	@docker exec -it \
		goodl-dev \
		CompileDaemon -build="go build -v github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/goodl-lib/datapolish" -directory="/go/src/github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/goodl-lib" -command="ginkgo /go/src/github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/goodl-lib/datapolish"  -color=true -recursive=true

dev-test-lib-addressManager:
	@docker exec -it \
		goodl-dev \
		CompileDaemon -build="go build -v github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/goodl-lib/jsonapi/addressManager" -directory="/go/src/github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/goodl-lib" -command="ginkgo /go/src/github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/goodl-lib/jsonapi/addressManager"  -color=true -recursive=true

dev-test-lib-dbMan:
	@docker exec -it \
		goodl-dev \
		CompileDaemon -build="go build -v github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/goodl-lib/dbMan" -directory="/go/src/github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/goodl-lib" -command="ginkgo /go/src/github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/goodl-lib/dbMan"  -color=true -recursive=true

dev-test-lib-tripMan:
	@docker exec -it \
		goodl-dev \
		CompileDaemon -build="go build -v github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/goodl-lib/jsonapi/tripMan" -directory="/go/src/github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/goodl-lib/jsonapi/tripMan" -command="ginkgo --noisyPendings=false /go/src/github.com/OpenDriversLog/goodl/vendor/github.com/OpenDriversLog/goodl-lib/jsonapi/tripMan"  -color=true -recursive=true

dev-doc:
	@docker exec -it \
		goodl-dev \
		godoc -http=:6060 -v=true -index

dev-remove:
	@docker stop goodl-dev \
		&& docker rm goodl-dev

run-tests: dev-test dev-test-compose-ff
