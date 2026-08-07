# Getting started with the Elastic Stack and Docker-Compose

This repo is for study purposes using this one as reference:
- https://github.com/elkninja/elastic-stack-docker-part-one

Blog reference [Getting started with the Elastic Stack and Docker-Compose](https://www.elastic.co/blog/getting-started-with-the-elastic-stack-and-docker-compose).

## Architecture


setup → es01 → kibana → otel-collector → favorite-go → go-auto

```
setup creates certificates
        ↓
es01 init
        ↓
setup configure kibana_system
        ↓
setup ends
```

### Load .env in terminal

```
set -a
source .env
set +a
```

```
docker compose cp \
  es01:/usr/share/elasticsearch/config/certs/ca/ca.crt \
  ./ca.crt

curl \
  --cacert ./ca.crt \
  --user "elastic:${ELASTIC_PASSWORD}" \
  https://localhost:9200

docker inspect \
  es01 \
  --format '{{range .Config.Env}}{{println .}}{{end}}' |
grep -E 'ELASTICSEARCH|ELASTIC_PASSWORD'

curl \
  --silent \
  --cacert ./ca.crt \
  --user "elastic:${ELASTIC_PASSWORD}" \
  'https://localhost:9200/_cluster/health?pretty'
```

https://github.com/open-telemetry/opentelemetry-go-instrumentation/blob/main/docs/getting-started.md