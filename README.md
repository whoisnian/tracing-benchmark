# tracing-benchmark

## todo
* [x] api server with gin/gorm/redis
* [x] instrumentation with elastic apm-server
* [x] instrumentation with jaegertracing
* [ ] instrumentation with openzipkin
* [ ] instrumentation with skywalking
* [x] prometheus metrics with grafana
* [ ] custom load generator
* [ ] collect benchmark results

## compose projects
| project                           | objective                                  | endpoint                            | authorization                    |
| --------------------------------- | ------------------------------------------ | ----------------------------------- | -------------------------------- |
| [10-prometheus](#10-prometheus)   | record resource usage of all containers    | grafana http://127.0.0.1:3000       | `admin` `KejTCKmMBIPxBm1m7h4f`   |
| [20-jaeger](#20-jaeger)           | provide `otlp` trace backend               | jaeger-query http://127.0.0.1:16686 | none                             |
| [30-apm](#30-apm)                 | provide `apm` trace backend                | kibana http://127.0.0.1:5601        | `elastic` `DVPMuwCOpH5iOPDFnjd5` |
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

### 90-application
* services:
  * mysql: `8.4.3`
  * redis: `7.4.1`
  * server: `v0.0.2` limits `cpus:4.000 mem_limit:4gb`
* start: `docker compose --env-file ./deploy/90-application/.env.example --file ./deploy/90-application/compose.yaml up -d`
* down: `docker compose --env-file ./deploy/90-application/.env.example --file ./deploy/90-application/compose.yaml down -v`

## bench
| backend | total requests | requests per second | mean | min | max | 50% | 75% | 90% | 95% | 99% |
| ------- | -------------- | ------------------- | ---- | --- | --- | --- | --- | --- | --- | --- |
| none    |                |                     |      |     |     |     |     |     |     |     |
| otlp    |                |                     |      |     |     |     |     |     |     |     |
| apm     |                |                     |      |     |     |     |     |     |     |     |

```sh
# To reduce the performance impact of docker-proxy, do not use http://127.0.0.1:8080 in benchmark.
CONTAINER_ID=$(docker compose --env-file ./deploy/90-application/.env.example --file ./deploy/90-application/compose.yaml ps --quiet server | head -n1)
CONTAINER_IP=$(docker inspect --format '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $CONTAINER_ID | head -n1)

# warm up
ab -n 10000 -c 50 "http://$CONTAINER_IP:8080/ping/GRM" # longest request 67 ms
ab -n 10000 -c 50 "http://$CONTAINER_IP:8080/ping/GRM" # longest request 12 ms

# start
ab -t 60 -n 100000 -c 50 "http://$CONTAINER_IP:8080/ping/GRM"
```
