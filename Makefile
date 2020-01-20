.PHONY: build build-http docker-http docker-aws deploy-api clean

build: build-http docker-http

build-http:
	GOOS=linux GOARCH=amd64 go build -o ./docker/api/bin/service-deployer cmd/api/*.go

docker-http:
	docker build -t grid/service-deployer docker/api

docker-aws:
	docker tag 
	docker push 598240822331

deploy-service:
	kubectl create -f k8s/srvc-aws.yaml --namespace ehernandez

deploy-pod:
	kubectl create -f k8s/pod-aws.yml --namespace ehernandez

remove:
	kubectl --namespace=ehernandez delete -f k8s/pod-aws.yml || true
	kubectl --namespace=ehernandez delete -f k8s/srvc-aws.yaml || true

clean:
	docker rmi grid/service-deployer || true
	docker rmi service-deployer || true

run:
	go run cmd/api/*.go

delete:
	kubectl delete -f k8s/pod-aws.yml --namespace ehernandez

send:
	go run sender/main.go

test-environment:
	kubectl create -f testFiles/author-deployment.yml --namespace ehernandez
	kubectl create -f testFiles/publish-deployment.yml --namespace ehernandez

delete-test-environment:
	kubectl delete -f testFiles/publish-deployment.yml --namespace ehernandez
	kubectl delete -f testFiles/author-deployment.yml --namespace ehernandez

new-tag:
	git tag -a v"$(VERSION)" -m "version v$(VERSION)"
	git push --tags
