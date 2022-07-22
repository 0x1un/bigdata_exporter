package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func RetrieveMetricFromYarnProxy(target string, taskID string) (JMXBeans, error) {
	uri := fmt.Sprintf("http://%s/proxy/%s/metrics/json", target, parseAppAttemptStr(taskID))
	resp, err := http.Get(uri)
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
func yarnTaskMetricHandler(w http.ResponseWriter, r *http.Request) {
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
	metrics, ok := cfg.Modules[moduleName]
	if !ok {
		http.Error(w, moduleName+" module not defined", http.StatusBadRequest)
		return
	}

	scrapeYarnMetric(info, metrics.ScrapeMetrics, target, w, r)

}
func scrapeYarnMetric(info JMXBeans, metrics []string, target string, w http.ResponseWriter, r *http.Request) {
	openTasks := make(map[string]int, 0)
	for _, bean := range info["beans"] {
		if name, ok := bean["name"]; !ok {
			continue
		} else {
			if strings.HasPrefix(name.(string), "Hadoop:service=ResourceManager,name=RpcActivityForPort") &&
				bean.Get("tag.serverName") == "ApplicationMasterProtocolService" {
				_ = json.Unmarshal([]byte(bean.Get("ApplicationMasterProtocolService").(string)), &openTasks)
			}
		}
	}
	for task, state := range openTasks {
		if state != 1 {
			continue
		}
		beans, err := RetrieveMetricFromYarnProxy(target, parseAppAttemptStr(task))
		if err != nil {
			log.Errorf("failed to retrieve yarn proxy metrics, %s", err)
			continue
		}
		// TODO 取yarn proxy中的任务gauge参数
		_ = beans
	}
}
func parseAppAttemptStr(s string) string {
	if s == "" {
		return ""
	}
	sp := strings.Split(s, "_")
	if len(sp) == 4 {
		return fmt.Sprintf("application_%s_%s", sp[1], sp[2])
	}
	return ""
}
