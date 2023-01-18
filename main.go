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
