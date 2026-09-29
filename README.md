# fortiweb_exporter

A minimal Prometheus exporter for FortiWeb, architected like
[fortigate_exporter](https://github.com/mlueckert/fortigate_exporter). It
implements the multi-target `/probe` pattern (à la the Prometheus
[blackbox_exporter](https://github.com/prometheus/blackbox_exporter)):
one exporter instance can scrape any number of FortiWeb devices, with
per-target authentication and IP-based access restriction.

The following probes are implemented:

| Probe | FortiWeb monitor API endpoint |
| --- | --- |
| `System/Resource` | `api/v2.0/system/status.systemresource` |
| `System/Status` | `api/v2.0/system/status.systemstatus` |
| `System/HAStatus` | `api/v2.0/system/status.hastatus` |
| `System/Interface` | `api/v2.0/system/status.systemoperation` |
| `Policy/Status` | `api/v2.0/policy/policystatus` |

The endpoint schemas are documented in [docs/](docs).

## Metrics

| Metric | Description |
| --- | --- |
| `probe_success` | Whether the last scrape of a target succeeded (1) or failed (0) |
| `probe_duration_seconds` | How long the scrape of a target took |
| `fortiweb_cpu_usage_ratio` | Current CPU usage ratio (0-1) |
| `fortiweb_memory_usage_ratio` | Current memory usage ratio (0-1) |
| `fortiweb_disk_usage_ratio` | Current disk usage ratio (0-1) |
| `fortiweb_current_sessions` | Current amount of sessions |
| `fortiweb_connections_per_second` | Current amount of new connections per second |
| `fortiweb_log_disk_info{status}` | Status of the log disk (info metric) |
| `fortiweb_db_status_info{status}` | Status of the local database (info metric) |
| `fortiweb_system_info{serial,firmware_version,cluster_name,cluster_role}` | System information (info metric); `cluster_role` is the role of this device in the cluster member list |
| `fortiweb_uptime_seconds` | Time since the system was started (minute resolution) |
| `fortiweb_ha_cfg_sync_state_info{state}` | HA configuration synchronization state, e.g. `In sync` (info metric) |
| `fortiweb_interface_info{interface,alias,ip_netmask,speed_duplex}` | Network interface information (info metric) |
| `fortiweb_interface_link_up{interface}` | Whether the interface link is up (1) or not (0) |
| `fortiweb_interface_transmit_packets_total{interface}` | Packets transmitted on the interface |
| `fortiweb_interface_receive_packets_total{interface}` | Packets received on the interface |
| `fortiweb_interface_transmit_bytes_total{interface}` | Bytes transmitted on the interface |
| `fortiweb_interface_receive_bytes_total{interface}` | Bytes received on the interface |
| `fortiweb_policy_sessions{policy,name,status,protocol,http_port,https_port,mode}` | Current amount of sessions per policy |
| `fortiweb_policy_connections_per_second{...}` | Current amount of new connections per second per policy |
| `fortiweb_policy_client_rtt{...}` | Client round trip time per policy, as reported by FortiWeb |
| `fortiweb_policy_server_rtt{...}` | Server round trip time per policy, as reported by FortiWeb |
| `fortiweb_policy_app_response_time{...}` | Application response time per policy, as reported by FortiWeb |
| `fortiweb_exporter_build_info{version,revision}` | Exporter build information |

## Configuration

### Authentication map / profiles

FortiWeb authenticates with an `Authorization: <token>` header, where
`<token>` is the base64 encoding of
`{"username":"...","password":"...","vdom":"..."}`. The exporter builds
this token from `username`, `password` and `vdom` (default `root`).

Each `/probe` request carries `target` and (optionally) `profile` query
parameters. The `profile` names an entry in the auth-file (see
[fortiweb-key.yaml.example](fortiweb-key.yaml.example)) which can hold
`username`, `password`, `vdom` (or a pre-encoded `token`) and selects which
probes should run for that target, via `probes.include` and/or
`probes.exclude` lists of probe name prefixes (`System/Resource`,
`System/Status`, `System/HAStatus`, `System/Interface`, `Policy/Status`). If a profile is omitted or has no probe
selection, all probes run.

The credentials can also be passed (or overridden) per request with the
`username`, `password`, `vdom` query parameters, or a pre-encoded `token`
query parameter. Request parameters take precedence over the profile; a
`token` parameter takes precedence over credentials.

Point the exporter at your auth-file with `-auth-file` (default
`fortiweb-key.yaml`).

### IP restriction

Both `/metrics` and `/probe` are protected by an IP allow-list, configured
with `-allowed-subnets` (comma-separated list of IP prefixes; default
`0.0.0.0`, which allows any IP). Requests from IPs that don't match any
configured prefix receive `403 Forbidden`.

### Flags

* `-auth-file`: file containing the authentication/profile map (default `fortiweb-key.yaml`).
* `-listen`: address to listen on (default `:9723`).
* `-scrape-timeout`: max seconds to allow a scrape to take (default `30`).
* `-https-timeout`: TLS handshake timeout in seconds (default `10`).
* `-insecure`: allow insecure (e.g. self-signed) TLS certificates on FortiWeb devices (default `false`).
* `-extra-ca-certs`: comma-separated files containing extra PEMs to trust for TLS connections.
* `-allowed-subnets`: comma-separated list of allowed IPs or subnet prefixes (default `0.0.0.0`, meaning any IP).

## Usage

Build and run:

```sh
go build -o fortiweb_exporter .
cp fortiweb-key.yaml.example fortiweb-key.yaml
./fortiweb_exporter
```

Scrape a FortiWeb device via `/probe`:

```sh
curl 'http://localhost:9723/probe?target=https://fortiweb.example.com&profile=default'

# or with credentials passed per request
curl 'http://localhost:9723/probe?target=https://fortiweb.example.com&username=admin&password=<password>&vdom=root'
```

Example Prometheus scrape config:

```yaml
scrape_configs:
  - job_name: fortiweb
    metrics_path: /probe
    static_configs:
      - targets: ['https://fortiweb.example.com']
    params:
      profile: ['default']
      vdom: ['root']
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9723
```

## Tests

```sh
go test ./...
```

or, including format check and `go vet`:

```sh
make test
```

## Secret scanning

A [betterleaks](https://github.com/betterleaks/betterleaks) pre-commit hook
blocks commits that contain secrets. Enable it once per clone:

```sh
make hooks
```

It uses a `betterleaks` binary from `PATH` if present, otherwise
`go run` (the first run takes about a minute to compile). To scan the whole
git history instead, run `make secrets-check`.

## Releases

Releases are created automatically by
[semantic-release](https://semantic-release.gitbook.io/) on every push to
`main` (see [.github/workflows/release.yml](.github/workflows/release.yml)).
The next version is derived from the
[Conventional Commits](https://www.conventionalcommits.org/) messages since
the last release:

| Commit message | Release |
| --- | --- |
| `fix: ...` | patch (`1.0.0` → `1.0.1`) |
| `feat: ...` | minor (`1.0.0` → `1.1.0`) |
| `feat!: ...` or a `BREAKING CHANGE:` footer | major (`1.0.0` → `2.0.0`) |
| `chore:`, `docs:`, `test:`, ... | no release |

Binaries for Linux, Windows and macOS (amd64/arm64) are attached to each
GitHub release. They can be built locally with `make build-release`.
