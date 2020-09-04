build:
	mkdir -p bin/
	go build -o bin/owl

deploy: build
	scp bin/owl root@$(IP):/usr/local/bin


