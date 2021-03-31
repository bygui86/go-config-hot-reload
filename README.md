
# go-config-hot-reload

Explore configuration hot reloading options in Golang

## run

### file hot reload

```bash
# run application
go run ./file-hot-reload/main.go

# wait some seconds to have the application up and running

# edit config.yaml and change values

# pay attention to logs and notice that configurations are different now! 
```

### consul

```bash
# start consul
docker run -d --rm --name=consul -e CONSUL_BIND_INTERFACE=eth0 -p 8500:8500 consul

# insert configs in consul
curl \
    --request PUT \
    --data @./consul-reload/config-1.json \
    http://localhost:8500/v1/kv/samples/app

# run application
go run ./consul-reload/main.go

# wait some seconds to have the application up and running

# change a config
curl \
    --request PUT \
    --data @./consul-reload/config-2.json \
    http://localhost:8500/v1/kv/samples/app

# pay attention to logs and notice that configurations are different now!
```

## links

- https://openmymind.net/Golang-Hot-Configuration-Reload/
- https://medium.com/golangspec/sync-rwmutex-ca6c6c3208a0

### consul

- https://www.consul.io/docs/connect/native/go
- https://github.com/hashicorp/consul/
- https://pkg.go.dev/github.com/hashicorp/consul
- https://hub.docker.com/_/consul
- https://www.consul.io/api-docs/kv
- https://medium.com/@pinkudebnath/putting-config-into-consul-using-curl-54c702db602

### harvester

- https://github.com/beatlabs/harvester
- https://build.thebeat.co/harvester-b3bfa19f16e
