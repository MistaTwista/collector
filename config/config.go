package config

import (
	"fmt"
	"log"
	"io"
	"io/ioutil"
	"net/url"
	"time"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type MetricType string
type MetricName string
type JobType string
const (
	JSON JobType = "json"
)
var ValidWorkTypes = [...]JobType{JSON}
const (
	Gauge MetricType = "gauge"
	Counter MetricType = "counter"
)
var ValidMetricTypes = [...]MetricType{Gauge,Counter}

func (m MetricName) Validate(prefix string) error {
	name := string(m)
	if (prefix != "") {
		name = fmt.Sprintf("%s_%s", prefix, name)
	}

	r, err := regexp.Compile(`[^a-zA-Z_:][^a-zA-Z0-9_:]*`)
	if err != nil {
		return fmt.Errorf("cannot compile metric name parser: %w", err)
	}

	match := r.FindAllString(name, -1)
	if len(match) != 0 {
		return fmt.Errorf("metric name is not valid '%s' (%s is not allowed)", name, strings.Join(match, ", "))
	}

	return nil
}

type Task struct {
	Name MetricName
	Req string
	Type MetricType
	Description string
}

func (t *Task) populateDefaults() {
	if t.Type == "" {
		t.Type = Gauge
	}
}

func (t *Task) Validate(prefix string) error {
	t.populateDefaults()
	if err := t.Name.Validate(prefix); err != nil {
		return err
	}

	switch t.Type {
	case Gauge, Counter:
	default:
		return fmt.Errorf("'%s' metrics type is not supported", t.Type)
	}

	return nil
}

type Job struct {
	Type JobType
	Name string
	Namespace string
	Subsystem string
	Url string
	ScrapeInterval time.Duration `yaml:"scrape_interval"`
	ScrapeDelay time.Duration `yaml:"scrape_delay"`
	Tasks []Task
}

func (j *Job) populateDefaults() {
	if j.Type == "" {
		j.Type = JSON
	}

	if j.ScrapeInterval == time.Duration(0) {
		j.ScrapeInterval = 1 * time.Minute
	}
}

func (j *Job) Prefix() string {
	if j.Subsystem == "" {
		return j.Namespace
	}

	return fmt.Sprintf("%s_%s", j.Namespace, j.Subsystem)
}

func (j *Job) Validate() error {
	j.populateDefaults()

	switch j.Type {
	case JSON:
	default:
		return fmt.Errorf("'%s' work type is not supported", j.Type)
	}

	if !strings.HasPrefix(j.Url, "http") && !strings.HasPrefix(j.Url, "https") {
		return fmt.Errorf("url must starts with http or https")
	}
	if _, err := url.ParseRequestURI(j.Url); err != nil {
		return fmt.Errorf("cannot parse url: %w", err)
	}

	if len(j.Tasks) == 0 {
		return fmt.Errorf("tasks are empty")
	}

	dupes := make(map[MetricName]bool)
	for i:=0;i<len(j.Tasks);i++ {
		task := &j.Tasks[i]
		if _, ok := dupes[task.Name]; ok {
			return fmt.Errorf("task '%s' is duplicated in config", task.Name)
		}

		if err := task.Validate(j.Prefix()); err != nil {
			return err
		}
		dupes[task.Name] = true
	}

	return nil
}

type Config struct {
	Jobs []Job
}

func (c *Config) Validate() error {
	if len(c.Jobs) == 0 {
		return fmt.Errorf("config without jobs")
	}

	dupes := make(map[string]bool)
	for i:=0;i<len(c.Jobs);i++ {
		wrk := &c.Jobs[i]

		if _, ok := dupes[wrk.Name]; ok {
			return fmt.Errorf("work '%s' is duplicated in config", wrk.Name)
		}
		if err := wrk.Validate(); err != nil {
			return fmt.Errorf("work '%s' is not valid: %w", wrk.Name, err)
		}

		dupes[wrk.Name] = true
	}


	return nil
}

func Init(r io.Reader) (*Config, error) {
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
