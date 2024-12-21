# Shelly exporter
A prometheus exporter for shelly devices

## Usage example
Get the status for device 172.16.13.111
```bash
curl http://localhost:9118/probe?ip=172.16.13.111
# HELP shelly_meter_power_watts Power consumption of the Shelly device's meters in watts
# TYPE shelly_meter_power_watts gauge
shelly_meter_power_watts{ip="172.16.13.111",meter="0"} 11.39
# HELP shelly_relay_status Status of the Shelly device's relays (1 for on, 0 for off)
# TYPE shelly_relay_status gauge
shelly_relay_status{ip="172.16.13.111",relay="0"} 1
# HELP shelly_temperature_celsius Temperature of the Shelly device in Celsius
# TYPE shelly_temperature_celsius gauge
shelly_temperature_celsius{ip="172.16.13.111"} 23.43
# HELP shelly_uptime_seconds Uptime of the Shelly device in seconds
# TYPE shelly_uptime_seconds gauge
shelly_uptime_seconds{ip="172.16.13.111"} 5873
# HELP shelly_wifi_rssi RSSI of the Shelly device's WiFi connection
# TYPE shelly_wifi_rssi gauge
shelly_wifi_rssi{ip="172.16.13.111",ssid="R4"} -59
```