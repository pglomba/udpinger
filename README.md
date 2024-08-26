<h1 style="text-align: center">
  <br>
  UDPinger
  <br>
</h1>

<h4 align="center">Simple ping based on UDP</h4>

<p style="text-align: center">
	<a href="https://github.com/pglomba/udpinger/actions/workflows/ci.yaml"><img src="https://github.com/pglomba/udpinger/actions/workflows/ci.yaml/badge.svg"></a>
</p>

### Features
* Client/server service for measuring UDP based round trip time. 
* Config file based service configuration.
* Metrics can be exposed in Prometheus format on a specified address or send to stdout.

### Config
Use `--config` flag to specify path to the YAML config file.

Example `config.yaml`:
```yaml
listen: 127.0.0.1:1051
interval: 2
count: 5
timeout: 2
targets:
  - target1:1052
  - target2:1053
  - target3:1054
unit: ms
exporter_type: prometheus
prometheus_address: 127.0.0.1:2112
```
### Run
```bash
git clone "https://github.com/pglomba/udpinger.git"
make build
./udpinger --config /path/to/config.yaml 
```


