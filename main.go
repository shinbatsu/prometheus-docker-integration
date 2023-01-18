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
