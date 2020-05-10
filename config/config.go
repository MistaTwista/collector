package config

import (
	"fmt"
	"log"
	"io"
	"io/ioutil"
	"time"
	"regexp"

	"gopkg.in/yaml.v2"
)

type MetricType string
type MetricName string
const (
	Gauge MetricType = "gauge"
	Counter MetricType = "counter"
)

func (m MetricName) Validate() error {
	r, err := regexp.Compile("[^a-zA-Z0-9:_]")
	if err != nil {
		return fmt.Errorf("cannot compile metric name parser: %w", err)
	}
	match := r.FindString(string(m))
	if match != "" {
		return fmt.Errorf("metric name is not valid '%s' (%s is not allowed)", m, match)
	}

	return nil
}

type DataMap struct {
	Name MetricName
	Req string
	Ptype MetricType
	Description string
}

func (d *DataMap) Validate() error {
	if err := d.Name.Validate(); err != nil {
		return err
	}

	switch d.Ptype {
	case Gauge, Counter:
	default:
		return fmt.Errorf("'%s' metrics type is not supported", d.Ptype)
	}

	return nil
}

type Work struct {
	Name string
	Namespace string
	Subsystem string
	Url string
	Method string
	Every time.Duration
	Delay time.Duration
	Mapping []DataMap
}

func (w *Work) Validate() error {
	for _, data := range w.Mapping {
		if err := data.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type Config struct {
	Works []Work
}

func (c *Config) Validate() error {
	for _, wrk := range c.Works {
		if err := wrk.Validate(); err != nil {
			return fmt.Errorf("work '%s' is not valid: %w", wrk.Name, err)
		}
	}

	return nil
}

func Load(r io.Reader) (*Config, error) {
	cbdata, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("cannot read config data %w", err)
	}

	c := Config{}
	err = yaml.Unmarshal(cbdata, &c)
	if err != nil {
		log.Fatal(err)
	}

	err = c.Validate()
	if err != nil {
		return nil, fmt.Errorf("config is not valid: %w", err)
	}

	return &c, nil
}
