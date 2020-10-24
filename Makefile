build:
	mkdir -p bin/
	CGO_ENABLED=0 GOOS=linux go build -o bin/owl 

deploy: build
	scp bin/owl root@$(IP):/usr/local/bin

tls:
	go build -o ./bin/owl && ./bin/owl hcloud tls issue token=$(HCLOUD_DNS_TOKEN) email=$(EMAIL) domains=*.$(DOMAIN),$(DOMAIN) cert-path=tls cert-storage=consul account-path=account/tls account-storage=consul debug=true

dev.consul:
	consul agent -config-dir ./.config/dev/consul -data-dir ./tmp

