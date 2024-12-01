# tracing-benchmark

## todo
* [x] api server with gin/gorm/redis
* [x] instrumentation with elastic apm-server
* [x] instrumentation with jaegertracing
* [x] instrumentation with openzipkin
* [x] instrumentation with skywalking
* [x] prometheus metrics with grafana
* [ ] custom load generator
* [ ] collect benchmark results

## compose projects
| project                           | objective                                  | endpoint                            | authorization                    |
| --------------------------------- | ------------------------------------------ | ----------------------------------- | -------------------------------- |
| [10-prometheus](#10-prometheus)   | record resource usage of all containers    | grafana http://127.0.0.1:3000       | `admin` `KejTCKmMBIPxBm1m7h4f`   |
| [20-jaeger](#20-jaeger)           | provide `otlp` trace backend               | jaeger-query http://127.0.0.1:16686 | none                             |
| [30-apm](#30-apm)                 | provide `apm` trace backend                | kibana http://127.0.0.1:5601        | `elastic` `DVPMuwCOpH5iOPDFnjd5` |
| [40-zipkin](#40-zipkin)           | provide `zipkin` trace backend             | zipkin-lens http://127.0.0.1:9411   | none                             |
| [50-skywalking](#50-skywalking)   | provide `skywalking` trace backend         | skywalking-ui http://127.0.0.1:8090 | none                             |
| [90-application](#90-application) | instrument api server using gin/gorm/redis | api http://127.0.0.1:8080           | none                             |

### 10-prometheus
* services:
  * prometheus: `v2.53.2`
  * grafana: `11.3.0`
  * cadvisor: `v0.49.1`
* start: `docker compose --env-file ./deploy/10-prometheus/.env.example --file ./deploy/10-prometheus/compose.yaml up -d`
* down: `docker compose --env-file ./deploy/10-prometheus/.env.example --file ./deploy/10-prometheus/compose.yaml down -v`

### 20-jaeger
* services:
  * elasticsearch: `8.15.3` limits `cpus:2.000 mem_limit:4gb`
  * jaeger-collector: `1.62.0` limits `cpus:2.000 mem_limit:4gb`
  * jaeger-query: `1.62.0`
* start: `docker compose --env-file ./deploy/20-jaeger/.env.example --file ./deploy/20-jaeger/compose.yaml up -d`
* down: `docker compose --env-file ./deploy/20-jaeger/.env.example --file ./deploy/20-jaeger/compose.yaml down -v`

### 30-apm
* services:
  * elasticsearch: `8.15.3` limits `cpus:2.000 mem_limit:4gb`
  * kibana: `8.15.3`
  * apm-server: `8.15.3` limits `cpus:2.000 mem_limit:4gb`
* start: `docker compose --env-file ./deploy/30-apm/.env.example --file ./deploy/30-apm/compose.yaml up -d`
* down: `docker compose --env-file ./deploy/30-apm/.env.example --file ./deploy/30-apm/compose.yaml down -v`

### 40-zipkin
* services:
  * elasticsearch: `8.15.3` limits `cpus:2.000 mem_limit:4gb`
  * zipkin-slim: `3.4.2` limits `cpus:2.000 mem_limit:4gb`
* start: `docker compose --env-file ./deploy/40-zipkin/.env.example --file ./deploy/40-zipkin/compose.yaml up -d`
* down: `docker compose --env-file ./deploy/40-zipkin/.env.example --file ./deploy/40-zipkin/compose.yaml down -v`

### 50-skywalking
* services:
  * elasticsearch: `8.15.3` limits `cpus:2.000 mem_limit:4gb`
  * skywalking-oap-server: `10.1.0` limits `cpus:2.000 mem_limit:4gb`
  * skywalking-ui: `10.1.0`
* start: `docker compose --env-file ./deploy/50-skywalking/.env.example --file ./deploy/50-skywalking/compose.yaml up -d`
* down: `docker compose --env-file ./deploy/50-skywalking/.env.example --file ./deploy/50-skywalking/compose.yaml down -v`

### 90-application
* services:
  * mysql: `8.4.3`
  * redis: `7.4.1`
  * server: `v0.0.4` limits `cpus:4.000 mem_limit:4gb`
* start default: `docker compose --env-file ./deploy/90-application/.env.example --file ./deploy/90-application/compose.yaml --profile default up -d`
* start skywalking: `docker compose --env-file ./deploy/90-application/.env.example --file ./deploy/90-application/compose.yaml --profile skywalking up -d`
* down all: `docker compose --env-file ./deploy/90-application/.env.example --file ./deploy/90-application/compose.yaml --profile default --profile skywalking down -v`

## bench
| backend    | total requests | requests per second | mean  | min | max | 50% | 75% | 90% | 95% | 99% |
| ---------- | -------------- | ------------------- | ----- | --- | --- | --- | --- | --- | --- | --- |
| none       | 868239         | 14470.44            | 3.455 | 1   | 65  | 3   | 4   | 4   | 4   | 4   |
| otlp*      | 467025         | 7778.65             | 6.428 | 0   | 86  | 4   | 4   | 7   | 16  | 63  |
| apm*       | 602956         | 10044.54            | 4.978 | 0   | 85  | 4   | 4   | 5   | 6   | 53  |
| zipkin     |                |                     |       |     |     |     |     |     |     |     |
| skywalking |                |                     |       |     |     |     |     |     |     |     |

* otlp(server drop): total spans 1401186, saved spans 634837, dropped 766349 54.69%
* apm(client drop): total spans 1839045, saved spans 1385113, dropped 453932 24.68%

```sh
# To reduce the performance impact of docker-proxy, do not use http://127.0.0.1:8080 in benchmark.
CONTAINER_ID=$(docker compose --env-file ./deploy/90-application/.env.example --file ./deploy/90-application/compose.yaml ps --quiet server | head -n1)
CONTAINER_IP=$(docker inspect --format '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $CONTAINER_ID | head -n1)

# warm up
ab -n 10000 -c 50 "http://$CONTAINER_IP:8080/ping/GRM" # longest request 80 ms
ab -n 10000 -c 50 "http://$CONTAINER_IP:8080/ping/GRM" # longest request 24 ms

# start
ab -t 60 -n 1000000 -c 50 "http://$CONTAINER_IP:8080/ping/GRM"
# Concurrency Level:      50
# Time taken for tests:   60.039 seconds
# Complete requests:      467025
# Failed requests:        0
# Total transferred:      56043000 bytes
# HTML transferred:       1868100 bytes
# Requests per second:    7778.65 [#/sec] (mean)
# Time per request:       6.428 [ms] (mean)
# Time per request:       0.129 [ms] (mean, across all concurrent requests)
# Transfer rate:          911.56 [Kbytes/sec] received

# Connection Times (ms)
#               min  mean[+/-sd] median   max
# Connect:        0    1   0.4      1       6
# Processing:     0    5  10.7      2      86
# Waiting:        0    4  10.5      2      84
# Total:          0    6  10.7      4      86

# Percentage of the requests served within a certain time (ms)
#   50%      4
#   66%      4
#   75%      4
#   80%      4
#   90%      7
#   95%     16
#   98%     58
#   99%     63
#  100%     86 (longest request)
```
