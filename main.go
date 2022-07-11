package main

import (
	"bigdata_exporter/config"
	"encoding/json"
	"fmt"
	"github.com/gobeam/stringy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
	"net/http"
	"strings"
)

type Bean map[string]any
type JMXBeans map[string][]Bean

func (b Bean) Get(metricName string) any {
	if value, ok := b[metricName]; ok {
		return value
	}
	return nil
}

func (j JMXBeans) Get(name string) Bean {
	for _, bean := range j["beans"] {
		if _name, ok := bean["name"]; ok {
			if _name == name {
				return bean
			}
		}
	}
	return nil
}

func RetrieveMetricFromTarget(target string) (JMXBeans, error) {
	resp, err := http.Get("http://" + target + "/jmx")
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var out JMXBeans
	return out, json.Unmarshal(data, &out)
}

func metricHandler(w http.ResponseWriter, r *http.Request) {
	moduleName := r.URL.Query().Get("module")
	if moduleName == "" {
		http.Error(w, "module name is missing", http.StatusBadRequest)
		return
	}
	params := r.URL.Query()
	target := params.Get("target")
	if target == "" {
		http.Error(w, "target parameter is missing", http.StatusBadRequest)
		return
	}
	info, err := RetrieveMetricFromTarget(target)
	if err != nil {
		http.Error(w, "retrieve metric failed: "+target+" "+err.Error(), http.StatusBadRequest)
		return
	}
	switch moduleName {
	case "namenode":
		scrapeMetric(info, config.Cfg.Modules.Namenode.ScrapeMetrics, w, r)
	case "yarn":
		scrapeMetric(info, config.Cfg.Modules.Yarn.ScrapeMetrics, w, r)
	case "hbase":
		scrapeMetric(info, config.Cfg.Modules.Hbase.ScrapeMetrics, w, r)
	}
}

func scrapeMetric(info JMXBeans, metrics []string, w http.ResponseWriter, r *http.Request) {
	registry := prometheus.NewRegistry()
	for _, v := range metrics {
		name, metricName, helpMsg, TagName := parseScrapeMetric(v)
		if name == "" {
			continue
		}
		opts := prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_%s", TagName, stringy.New(metricName).SnakeCase("?", "").ToLower()),
			Help: helpMsg,
		}
		value := info.Get(name).Get(metricName).(float64)
		if value != 0 && name == "Hadoop:service=HBase,name=Master,sub=Server" {
			opts.ConstLabels = prometheus.Labels{
				"deadRegionServers": info.Get(name).Get("tag.deadRegionServers").(string),
			}
		}
		gauge := prometheus.NewGauge(opts)
		gauge.Set(value)
		registry.MustRegister(gauge)
	}
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

// name metric_name help_msg tag
func parseScrapeMetric(line string) (string, string, string, string) {
	sp := strings.Split(line, " ")
	if len(sp) == 4 {
		return sp[0], sp[1], sp[2], sp[3]
	}
	return "", "", "", ""
}

func main() {
	if config.Cfg == nil {
		panic("config parse error")
	}
	http.HandleFunc("/metrics", metricHandler)
	http.ListenAndServe(":7070", nil)
}
