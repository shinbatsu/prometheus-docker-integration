package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	labels = []string{"name", "stack", "service"}

	memUsageDesc = prometheus.NewDesc(
		"docker_container_memory_usage_bytes",
		"Total memory usage in bytes",
		labels, nil,
	)

	memReservationDesc = prometheus.NewDesc(
		"docker_container_memory_reservation_bytes",
		"Memory reserved for the container in bytes",
		labels, nil,
	)

	memLimitDesc = prometheus.NewDesc(
		"docker_container_memory_limit_bytes",
		"Memory limit for the container in bytes",
		labels, nil,
	)

	docker *client.Client
)
type dockerCollector struct{}

func (c dockerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- memUsageDesc
	ch <- memReservationDesc
	ch <- memLimitDesc
}

func (c dockerCollector) Collect(ch chan<- prometheus.Metric) {
	cgroupVersion := detectCgroupVersion()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	containers, err := docker.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		log.Printf("failed to list containers: %v", err)
		return
	}

	for _, container := range containers {
		name := cleanContainerName(container.Names)
		stack, service := extractStackService(container.Labels)

		usageBytes, err := containerUsageBytes(container.ID, cgroupVersion)
		if err != nil {
			log.Printf("failed to read usage for container %s: %v", name, err)
			continue
		}

		totalCache, err := containerTotalCacheBytes(container.ID, cgroupVersion)
		if err != nil {
			log.Printf("failed to read cache for container %s: %v", name, err)
			continue
		}

		inspect, err := docker.ContainerInspect(ctx, container.ID)
		if err != nil {
			log.Printf("failed to inspect container %s: %v", name, err)
			continue
		}

		emitMemoryMetrics(ch, name, stack, service, usageBytes-totalCache, inspect.HostConfig.MemoryReservation, inspect.HostConfig.Memory)
	}
}

func emitMemoryMetrics(ch chan<- prometheus.Metric, name, stack, service string, usage, reservation, limit int64) {
	ch <- prometheus.MustNewConstMetric(memUsageDesc, prometheus.GaugeValue, float64(usage), name, stack, service)
	ch <- prometheus.MustNewConstMetric(memReservationDesc, prometheus.GaugeValue, float64(reservation), name, stack, service)
	ch <- prometheus.MustNewConstMetric(memLimitDesc, prometheus.GaugeValue, float64(limit), name, stack, service)
}
func detectCgroupVersion() string {
	if version, ok := os.LookupEnv("DOCKER_CLUSTER_CGROUP_VERSION"); ok {
		return version
	}
	return "v1"
}

func containerUsageBytes(containerID, cgroupVersion string) (int64, error) {
	path := "/host/sys/fs/cgroup/memory/docker/" + containerID + "/memory.usage_in_bytes"
	if cgroupVersion == "v2" {
		path = "/host/docker-" + containerID + ".scope/memory.current"
	}
	return readIntFromFile(path)
}

func containerTotalCacheBytes(containerID, cgroupVersion string) (int64, error) {
	var path string
	if cgroupVersion == "v2" {
		path = "/host/docker-" + containerID + ".scope/memory.stat"
		stat, err := readMapFile(path)
		if err != nil {
			return 0, err
		}
		return stat["inactive_file"], nil
	}
	path = "/host/sys/fs/cgroup/memory/docker/" + containerID + "/memory.stat"
	stat, err := readMapFile(path)
	if err != nil {
		return 0, err
	}
	return stat["total_cache"], nil
}

func readIntFromFile(path string) (int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
}

func readMapFile(path string) (map[string]int64, error) {
	result := make(map[string]int64)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) != 2 {
			continue
		}
		value, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}
		result[parts[0]] = value
	}
	return result, scanner.Err()
}
func cleanContainerName(names []string) string {
	if len(names) == 0 {
		return "-"
	}
	name := names[0]
	if len(name) > 0 && name[0] == '/' {
		return name[1:]
	}
	return name
}

func extractStackService(labels map[string]string) (stack, service string) {
	if val := labels["io.rancher.stack_service.name"]; val != "" {
		parts := strings.SplitN(val, "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	if val := labels["com.docker.swarm.service.name"]; val != "" {
		ns := labels["com.docker.stack.namespace"]
		if ns != "" {
			return ns, strings.TrimPrefix(val, ns+"_")
		}
		parts := strings.SplitN(val, "_", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
		return "-", val
	}
	stack = labels["com.docker.compose.project"]
	service = labels["com.docker.compose.service"]
	if stack == "" {
		stack = "-"
	}
	if service == "" {
		service = "-"
	}
	return stack, service
}

func main() {
	var err error
	docker, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("failed to create Docker client: %v", err)
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(dockerCollector{})

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	port := 9476
	log.Printf("Listening on :%d...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
