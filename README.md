# loadbalancer
A generic loadbalancer built on GoLang to help learn go and its concurrency. To register a host, it needs a list of upstream servers, hostname, port, healthcheck and routing algorithm to be used for that host. Multiple hosts can be registered at a time. The healthchecks dynamically update the available collection of upstream servers. 

## Setup
Build: 
```
go vet && go build -v -o $PWD/bin 
```
Run:
```
./bin/loadbalancer
```
## Features
This load-balancer features three routing algorithms:
1. Random
2. Round Robin
3. Weighted Round Robin

## Configurations
The loadbalancer will read yaml configurations from the root directory. You can specify filename in main package when you build. Default is **config.yaml**

| Attributes 	|  Values 	| type
|---	|---	|---
|  env 	|   	dev, prod | string
|  logLevel 	|   info, debug |	string
|   host.name	|   FQDN for host|   string
|   host.scheme	|   random, roundrobin, weightedroundrobin | string
|   host.health	|   healthcheck for upstream | string
|   host.interval	|   healthcheck interval | int
|   servers.name	| FQDN for upstream | string
|   servers.weight	| weight for weightedRR | int


A sample config file
```
env: dev
logLevel: info
hosts: 
- name: localhost 
  port: 3000
  scheme: roundrobin
  health: /health
  interval: 10
  servers:
  - name: www.google.com
    weight: 5
    health: 
  - name: 1.1.1.1
    weight: 3
  - name: www.youtube.com
    weight: 2
- name: localhost 
  port: 3001
  scheme: random
  health: /health
  interval: 10
  servers:
  - name: www.google.com
    weight: 5
    health: 
  - name: 1.1.1.1
    weight: 3
  - name: www.youtube.com
    weight: 2

```
## Progress
* ~~Setup Project~~

* ~~Create initializer Sequence~~

* ~~Implement proxy~~

* ~~Implement random, rr and weighted rr routing~~

* ~~[Added Late] Refactor packages without side effects~~

* ~~Add dynamic routing with healthchecks~~

* ~~Multiple concurrent hosts~~

* Unit tests

* Add metrics and profiling
