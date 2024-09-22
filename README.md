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

## run
```sh
# 10-prometheus
docker compose --env-file ./deploy/10-prometheus/.env.example --file ./deploy/10-prometheus/compose.yaml up -d
# grafana: visit http://127.0.0.1:3000

# 20-jaeger
docker compose --env-file ./deploy/20-jaeger/.env.example --file ./deploy/20-jaeger/compose.yaml up -d
# query-ui: visit http://127.0.0.1:16686

# 30-apm
docker compose --env-file ./deploy/30-apm/.env.example --file ./deploy/30-apm/compose.yaml up -d
# kibana: visit http://127.0.0.1:5601

# 90-application
docker compose --env-file ./deploy/90-application/.env.example --file ./deploy/90-application/compose.yaml up -d
# request: curl http://127.0.0.1:8080/ping/GRM
```
