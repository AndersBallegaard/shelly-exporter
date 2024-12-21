package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func slash(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("Go to /probe?ip=x.y.z.w to get the metrics for the IP x.y.z.w\n, or /metrics to get the metrics for the exporter itself\n"))
}

func probeEndpointFactory(hitCounter prometheus.Counter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reg := prometheus.NewRegistry()
		hitCounter.Inc()
		shelly_ip := r.URL.Query().Get("target")

		resp, err := http.Get("http://" + shelly_ip + "/status")
		if err != nil {
			http.Error(w, "Failed to get status from Shelly device", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, "Non-OK HTTP status: "+resp.Status, http.StatusInternalServerError)
			return
		}

		type ShellyStatusWifi struct {
			Connected bool   `json:"connected"`
			SSID      string `json:"ssid"`
			IP        string `json:"ip"`
			Rssi      int    `json:"rssi"`
		}

		type ShellyStatusRelay struct {
			IsOn bool `json:"ison"`
		}
		type ShellyStatusMeter struct {
			Power float64 `json:"power"`
		}

		var ShellyStatus struct {
			WifiSta     ShellyStatusWifi    `json:"wifi_sta"`
			Relay       []ShellyStatusRelay `json:"relays"`
			Meters      []ShellyStatusMeter `json:"meters"`
			Uptime      int                 `json:"uptime"`
			Temperature float64             `json:"temperature"`
		}
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal([]byte(body), &ShellyStatus); err != nil {
			http.Error(w, "Failed to unmarshal JSON response", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		wifiRSSIMetrics := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "shelly_wifi_rssi",
				Help: "RSSI of the Shelly device's WiFi connection",
			},
			[]string{"ip", "ssid"},
		)

		reg.MustRegister(wifiRSSIMetrics)

		wifiRSSIMetrics.With(prometheus.Labels{
			"ip":   ShellyStatus.WifiSta.IP,
			"ssid": ShellyStatus.WifiSta.SSID,
		}).Set(float64(ShellyStatus.WifiSta.Rssi))

		uptimeMetrics := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "shelly_uptime_seconds",
				Help: "Uptime of the Shelly device in seconds",
			},
			[]string{"ip"},
		)

		tempMetrics := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "shelly_temperature_celsius",
				Help: "Temperature of the Shelly device in Celsius",
			},
			[]string{"ip"},
		)

		reg.MustRegister(uptimeMetrics)
		reg.MustRegister(tempMetrics)

		uptimeMetrics.With(prometheus.Labels{
			"ip": ShellyStatus.WifiSta.IP,
		}).Set(float64(ShellyStatus.Uptime))

		tempMetrics.With(prometheus.Labels{
			"ip": ShellyStatus.WifiSta.IP,
		}).Set(ShellyStatus.Temperature)

		relayMetrics := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "shelly_relay_status",
				Help: "Status of the Shelly device's relays (1 for on, 0 for off)",
			},
			[]string{"ip", "relay"},
		)

		meterMetrics := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "shelly_meter_power_watts",
				Help: "Power consumption of the Shelly device's meters in watts",
			},
			[]string{"ip", "meter"},
		)

		reg.MustRegister(relayMetrics)
		reg.MustRegister(meterMetrics)

		for i, relay := range ShellyStatus.Relay {
			relayMetrics.With(prometheus.Labels{
				"ip":    ShellyStatus.WifiSta.IP,
				"relay": strconv.Itoa(i),
			}).Set(boolToFloat64(relay.IsOn))
		}

		for i, meter := range ShellyStatus.Meters {
			meterMetrics.With(prometheus.Labels{
				"ip":    ShellyStatus.WifiSta.IP,
				"meter": strconv.Itoa(i),
			}).Set(meter.Power)
		}

		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)

	}
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func main() {
	// Create a new registry.
	reg := prometheus.NewRegistry()

	// Create a new counter metric.
	hitCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hit_counter",
		Help: "Hits to the /probe endpoint since application start",
	})

	// Register the counter with the registry.
	reg.MustRegister(hitCounter)

	// Create a new HTTP handler for the metrics endpoint.
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.HandleFunc("/probe", probeEndpointFactory(hitCounter))
	http.HandleFunc("/", slash)
	// Start the HTTP server.
	log.Println("Starting server on :9118")
	log.Fatal(http.ListenAndServe(":9118", nil))
}
