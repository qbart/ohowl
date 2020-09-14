build:
	mkdir -p bin/
	CGO_ENABLED=0 GOOS=linux go build -o bin/owl 

deploy: build
	scp bin/owl root@$(IP):/usr/local/bin


