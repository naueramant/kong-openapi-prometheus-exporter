# Kong OpenAPI prometheus exporter

This prometheus takes a swagger file and uses the kong request log to generate prometheus metrics, grouping by the swagger paths and methods.

## Setup guide

### 1. Deploy the exporter

First of all we need to deploy the exporter. This can be done with the provided helm chart.

```bash

```

### 2. Add global kong HTTP log plugin

Next we need to add a global kong HTTP log plugin to the kong gateway. This plugin will log all requests to the exporter.

```bash

```

## Configuration

| **Variable**                   | **Default Value** | **Description**                                                                  |
| ------------------------------ | ----------------- | -------------------------------------------------------------------------------- |
| `log.level`                    | `info`            | The level of logging detail. Common values are `debug`, `info`, `warn`, `error`. |
| `log.format`                   | `json`            | The format of the log output. Common formats are `text` and `json`.              |
| `prometheus.path`              | `/metrics`        | The URL path where metrics are exposed.                                          |
| `prometheus.port`              | `9090`            | The port on which the Prometheus metrics endpoint listens.                       |
| `openapi.url`                  |                   | The URL of the OpenAPI specification.                                            |
| `openapi.file`                 |                   | The path to the OpenAPI specification file.                                      |
| `openapi.reload`               | `24h`             | The interval at which the OpenAPI documentation is reloaded.                     |
| `metrics.include_operation_id` | `false`           | Include the operation ID of endpoints in the metrics.                            |
| `metrics.headers`              | `[]`              | List of HTTP headers to be included in the metrics.                              |

**Warning**:

Including headers in the metrics can lead to a high cardinality of metrics, which can lead to performance issues in Prometheus. Use this feature with caution.

Don't include sensitive information in the headers, as they will be exposed in the metrics.
