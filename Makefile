build:
	mkdir -p bin/
	CGO_ENABLED=0 GOOS=linux go build -o bin/owl 

deploy: build
	scp bin/owl root@$(IP):/usr/local/bin

agent:
	go build -o ./bin/owl && ./bin/owl agent acltoken=secret debug=true account-path=ops/owl cert-path=owl

test.issue:
	http PUT :1914/tf/v1/certificate OhOwl-api-token:secret domains:='["owl.hashira.cloud"]'

dev.hashi:
	consul agent -dev&
	vault server -dev&
