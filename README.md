# Docker metrics in Prometheus

Export Docker containe metrics (memory usage, limits, and reservations) to Prometheus.

## Features

- Collcts memory metrics per container.
- Supports Docker Swarm, Rancher, and Compose labels for stack/service identification.
- Compatible with cgroups v1 and v2.
- Exposes `/metrics` endpoint for Prometheus scraping.

## Usage

### Running via Docker

You can run the exporter as a Docker container. Example:

```yaml
services:
  prometheus-integration:
    image: your-docker-username/prometheus-integration:latest
    privileged: true
    environment:
      DOCKER_CLUSTER_CGROUP_VERSION: v2
    network_mode: host
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /sys:/host:ro
    pid: host
    ports:
      - "9476:9476"
```

- The exporter reads container cgroups from `/host/sys`.  
- If you use cgroups v2, use `DOCKER_CLUSTER_CGROUP_VERSION=v2`.  
- The `/var/run/docker.sock` volume allows the exporter to request Docker container info.

### Running Locally

1. Build binary with Go 1.21:
```bash
go build -o prometheus-integration main.go
```
2. Run exporter:
```bash
./prometheus-integration
```
3. Metrics are available at `http://localhost:9476/metrics`.

### Running via Docker Compose

```yaml
version: '3.9'
services:
  prometheus-integration:
    build: .
    container_name: prometheus-docker-integration
    privileged: true
    pid: host
    network_mode: host
    environment:
      DOCKER_CLUSTER_CGROUP_VERSION: v2
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /sys:/host:ro
    ports:
      - "9476:9476"
    restart: unless-stopped
```