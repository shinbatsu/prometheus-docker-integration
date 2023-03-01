module prometheus-docker-integration

go 1.24.0

toolchain go1.24.6

require (
	github.com/docker/docker v23.0.2+incompatible
	github.com/prometheus/client_golang v1.17.0
)

replace github.com/distribution/reference => github.com/distribution/reference v0.6.0

require (
	github.com/Microsoft/go-winio v0.4.21 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/distribution/reference v0.0.0-00010101000000-000000000000 // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/go-connections v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.4.1-0.20230718164431-9a2bf3000d16 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gotest.tools/v3 v3.5.2 // indirect
)
