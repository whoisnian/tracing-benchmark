# tracing-benchmark

## todo
* [x] api server with gin/gorm/redis
* [x] instrumentation with elastic apm-server
* [x] instrumentation with jaegertracing
* [x] instrumentation with openzipkin
* [x] instrumentation with skywalking
* [x] prometheus metrics with grafana
* [x] collect benchmark results

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
  * jaeger: `2.1.0` limits `GOMAXPROCS:2 cpus:2.000 mem_limit:4gb`
* start: `docker compose --env-file ./deploy/20-jaeger/.env.example --file ./deploy/20-jaeger/compose.yaml up -d`
* down: `docker compose --env-file ./deploy/20-jaeger/.env.example --file ./deploy/20-jaeger/compose.yaml down -v`

### 30-apm
* services:
  * elasticsearch: `8.15.3` limits `cpus:2.000 mem_limit:4gb`
  * kibana: `8.15.3`
  * apm-server: `8.15.3` limits `GOMAXPROCS:2 cpus:2.000 mem_limit:4gb`
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
  * server: `v0.0.4` limits `GOMAXPROCS:2 cpus:2.000 mem_limit:4gb`
* start default: `docker compose --env-file ./deploy/90-application/.env.example --file ./deploy/90-application/compose.yaml --profile default up -d`
* start skywalking: `docker compose --env-file ./deploy/90-application/.env.example --file ./deploy/90-application/compose.yaml --profile skywalking up -d`
* down all: `docker compose --env-file ./deploy/90-application/.env.example --file ./deploy/90-application/compose.yaml --profile default --profile skywalking down -v`

## bench
| backend    | total requests | requests per second | avg   | max   | 50%   | 75%   | 90%   | 99%   |
| ---------- | -------------- | ------------------- | ----- | ----- | ----- | ----- | ----- | ----- |
| none       | 873465         | 29018.83            | 0.276 | 4.97  | 0.261 | 0.316 | 0.380 | 0.631 |
| otlp       | 529052         | 17576.42            | 0.489 | 8.68  | 0.405 | 0.535 | 0.764 | 2.13  |
| apm        | 597651         | 19917.30            | 0.449 | 14.14 | 0.370 | 0.466 | 0.601 | 2.49  |
| zipkin     | 384299         | 12806.59            | 0.770 | 18.61 | 0.517 | 0.810 | 1.38  | 4.84  |
| skywalking | 491645         | 16334.06            | 0.542 | 10.67 | 0.440 | 0.573 | 0.801 | 2.79  |

| backend    | total spans | saved spans | dropped spans | dropped percent | collector cpu | storage cpu |
| ---------- | ----------- | ----------- | ------------- | --------------- | ------------- | ----------- |
| otlp       | 529052*3    | 505345      | 1081811       | 68.16%          | 60%           | 200%        |
| apm        | 597651*3    | 774151      | 1018802       | 56.82%          | 150%          | 200%        |
| zipkin     | 384299*3    | 436442      | 716455        | 62.14%          | 180%          | 200%        |
| skywalking | 491645*3    | 227041*3    | 793812        | 53.82%          | 170%          | 60%         |

```sh
# To reduce the performance impact of docker-proxy, do not use http://127.0.0.1:8080 in benchmark.
CONTAINER_ID=$(docker compose --env-file ./deploy/90-application/.env.example --file ./deploy/90-application/compose.yaml ps --quiet server server-sw | head -n1)
CONTAINER_IP=$(docker inspect --format '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $CONTAINER_ID | head -n1)

# To avoid the bottleneck caused by ApacheBench's single thread, use https://github.com/wg/wrk/ instead.
wrk -c8 -t8 -d10 --latency "http://$CONTAINER_IP:8080/ping/GRM" # warm up
wrk -c8 -t8 -d30 --latency "http://$CONTAINER_IP:8080/ping/GRM" # actual benchmark
# Running 30s test @ http://172.18.0.4:8080/ping/GRM
#   8 threads and 8 connections
#   Thread Stats   Avg      Stdev     Max   +/- Stdev
#     Latency   276.50us   99.04us   4.97ms   80.39%
#     Req/Sec     3.65k   120.72     4.14k    69.52%
#   Latency Distribution
#      50%  261.00us
#      75%  316.00us
#      90%  380.00us
#      99%  631.00us
#   873465 requests in 30.10s, 99.96MB read
# Requests/sec:  29018.83
# Transfer/sec:      3.32MB
```
