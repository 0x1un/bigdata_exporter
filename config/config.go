package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var (
	Cfg = parse()
)

type MetricConfig struct {
	Modules struct {
		Namenode struct {
			ScrapeMetrics []string `yaml:"scrape_metrics"`
		} `yaml:"namenode"`
		Yarn struct {
			ScrapeMetrics []string `yaml:"scrape_metrics"`
		} `yaml:"yarn"`
		Hbase struct {
			ScrapeMetrics []string `yaml:"scrape_metrics"`
		} `yaml:"hbase"`
	} `yaml:"modules"`
}

func parse() *MetricConfig {
	data, _ := ioutil.ReadFile("config/config.yml")
	var mc MetricConfig
	if err := yaml.Unmarshal(data, &mc); err != nil {
		fmt.Println(err)
		return nil
	}
	return &mc
}
