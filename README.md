# loadbalancer
A generic loadbalancer built on GoLang to help learn go and its concurrency. To register a host, it needs a list of upstream servers, hostname, port, healthcheck and routing algorithm to be used for that host. Multiple hosts can be registered at a time. The healthchecks dynamically update the available collection of upstream servers. 

## Setup
Build and Run: 
```
chmod +x build.sh
./build
```
Run:
```
./bin/loadbalancer
```
Test:
use the docker-compose file to bring up example servers. You can change services and edit config.yaml to your requirement
```
docker-compose -f ./examples/docker-compose.yaml up -d 
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
|   host.timeout	|   Request timeout duration in seconds | int
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
  timeout: 8
  servers:
  - name: localhost:9081
    weight: 5
    health: 
  - name: localhost:9082
    weight: 3
  - name: localhost:9083
    weight: 2
- name: localhost 
  port: 3001
  scheme: random
  health: /health
  interval: 10
  timeout: 8
  servers:
  - name: localhost:9081
    weight: 5
    health: 
  - name: localhost:9082
    weight: 3
  - name: localhost:9083
    weight: 2
```
## Stress Test
**Routing**: Weighted Round Robin

**Summary**:
>Requests: 100000<br> 
Total:	101.7378 secs<br> 
Slowest:	0.3187 secs<br> 
Fastest:	0.0006 secs<br> 
Average:	0.0496 secs<br> 
Requests/sec:	982.9189<br> 
<br> Total data:	17900000 bytes<br> 
Size/request:	179 bytes

**Status code distribution**:
>[200]	100000 responses

**Latency distribution**:
> 10% in 0.0031 secs<br>
25% in 0.0428 secs<br>
50% in 0.0478 secs<br>
75% in 0.0645 secs<br>
90% in 0.0857 secs<br>
95% in 0.1009 secs<br>
99% in 0.1380 secs<br>

**Errors**:
 > 0

## Progress
* ~~Setup Project~~

* ~~Create initializer Sequence~~

* ~~Implement proxy~~

* ~~Implement random, rr and weighted rr routing~~

* ~~[Added Late] Refactor packages without side effects~~

* ~~Add dynamic routing with healthchecks~~

* ~~Multiple concurrent hosts~~

* ~~Unit tests~~

* ~~Graceful exit~~

* ~~Stress test and attach result~~

* ~~Add benchmarks~~

* ~~Proxy Timeouts~~

