```
favorite-go
  process HTTP
      │
      │ observed via eBPF
      ▼
go-auto
      │
      │ OTLP/HTTP
      ▼
otel-collector
      │
      ▼
Elasticsearch
```

https://github.com/open-telemetry/opentelemetry-go-instrumentation/blob/main/docs/how-it-works.md