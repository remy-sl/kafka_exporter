package kafka_exporter

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func ListenAndServe(metricsPath, addr string) error {
	http.HandleFunc("/", rootHandler(metricsPath))
	http.Handle(metricsPath, promhttp.Handler())
	return http.ListenAndServe(addr, nil)
}

func rootHandler(metricsPath string) http.HandlerFunc {
	if !strings.HasPrefix(metricsPath, "/") {
		metricsPath = "/" + metricsPath
	}
	return func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
	<head>
		<title>Kafka Collector</title>
	</head>
	<body>
		<p>
			<a href="` + metricsPath + `">Kafka Metrics</a>
		</p>
	</body>
</html>`,
		))
	}
}
