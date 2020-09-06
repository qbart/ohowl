# Oh Owl!

Hetzner Cloud utils

## Usage (Hetzner)

```
owl hcloud metadata # => {"id":"12345","hostname":"name","ip":"10.0.1.11","public_ip":"X.Y.Z.W"}
```

## Template rendering (generic)

```
owl tpl render /tmp/consul.json \
    ip=$(owl hcloud metadata | jq -r .ip) \
    node_name=$(owl hcloud metadata | jq -r .hostname) \
    > /opt/consul/config/default.json
```
