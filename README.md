# Prometheus SonarQube License Exporter

![ci](https://github.com/szevez/prometheus-sonarqube-license-exporter/actions/workflows/ci.yml/badge.svg)

## Requirements

- SonarQube API Token (Requires ‘Administer System’ permission)
- SonarQube URL

## Exported Metrics

| Name                          | Description                               |
|-------------------------------|-------------------------------------------|
| sonarqube_license_expires_at  | The date the SonarQube License expires at |
| sonarqube_license_current_loc | Current lines of code                     |
| sonarqube_license_max_loc     | The maximum lines of code of the license  |

## Docker

```sh
$ docker build -t szevez/prometheus-sonarqube-license-exporter:latest .
$ docker run -p 9191:9191 \
  -e SONARQUBE_TOKEN=YOURTOKEN \
  -e SONARQUBE_URL=https://your-sonarqube \
  szevez/prometheus-sonarqube-license-exporter:latest
```

## Local Development

- Requires go >= 1.18

```sh
$ export SONARQUBE_TOKEN=YOURTOKEN
$ export SONARQUBE_URL=https://your-sonarqube
$ go run main.go
```

Access exporter at `http://localhost:9191/metrics`
