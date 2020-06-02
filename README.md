# Collector
Scrape your app and prepare metrics for Prometheus

# Flags
- `-c config.yaml` - config with work description
- `-l :8080` - where http server listen

# Usage
1. Make config ([example](examples/config.example.yaml))
2. Run collector
3. Add [Prometheus](https://prometheus.io/) scrape work to `http://yourserver/metrics`

# Domain
- **Job** - source of data with 1+ metrics to get from there
- **Task** - prepare data from job to scrape with Prometheus
- **Duration** - [Duration](https://golang.org/pkg/time/#ParseDuration)

# Config
Configuration use yaml

All jobs described in **jobs** section:
```yaml
jobs:
  - name: job1
  - name: job2
```

Job params:
name|type|description|default|required
---|---|---|---|---
**type**|enum{"json"}|scrape job type. Currently only **json** supported|json|*
**name**|string|just a name for the job||*
**namespace**|string|Prometheus namespace (used in final metric name format)|""|
**subsystem**|string|Prometheus subsystem (also used in final metric name format)|""|
**url**|URL string (start with http/https)|Url to get data from||*
**scrape_interval**|duration|How frequently to scrape tagret|1m|
**scrape_delay**|duration|Delay before first run|0s|
**tasks**|Tasks for the job||*

Task params:
name|type|description|default|required
---|---|---|---|---
**name**|string|just task name (used in final metric name format)||*
**type**|enum{"gauge",counter"}|Prometheus metric type|gauge|*
**description**|string|Description of metric||
**req**|GJSON string|Request in json (look [example](#Example) below)||

## Metrics name
Collector forms name of metrics from:
- Job name
- Job namespace
- Job subsystem
- Task name

So for config:
```yaml
jobs:
  - name: users_stats
    namespace: company
    subsystem: project
    tasks:
      - name: reg_count
        type: counter
      - name: cpu_load
        type: gauge
```
Collector will form:
```text
company_project_users_stats_fails_total{name="users_stats|reg_count"} - counter with scrape fails grouped by name
company_project_reg_count - counter
company_project_cpu_load - gauge
```

## Example
Collector use [GJSON](https://github.com/tidwall/gjson) to get data from JSON responses, so your config might looks like:
```yaml
jobs:
  - name: project_stats
    ...
    mapping:
      important_counter:
        req: data.1.count
        ptype: counter
      important_nets_count:
        req: data.1.nets.#
        ptype: gauge
```
for data:
```json
{
  "data": [
    {"count": 44, "nets": ["ig", "fb", "tw"]},
    {"name": "important", "count": 68, "nets": ["fb", "tw"]},
    {"count": 47, "nets": ["ig", "tw"]}
  ]
}
```
it will create Prometheus metrics for **count** (68) and length of elements in nets (2)
There will be also error counter work each of task and the whole work

## CLI
You can play with GJSON in console:
```bash
$ make build
$ bin/json-play -json examples/data1.example.json -r data.1.nets.#
data.1.nets.# => 2
```

# Metrics
## Gauge
Works same as in usual Prometheus workflow, can increase and decrease

## Counter
Prometheus counters works only as increasing value, so do we:

### Example
You app data have metrics: `20 25 35 APPKILL 10 15 15 15 20`

Where **APPKILL** - when your app stopped for some reason and metric started from 0

When there was no changes in time - we did nothing (like 15,15,15 in example, we count just first 15)

As a reult our final metric after last values check will be: 20+5+10+10+5+5 = 55 (or just 35, KILL, 20)

There can be situations like `20 APPKILL 35` so Collector counts 35 instead of 55, because it will not know if app was killed.
Workflow for such problems can be more frequent requests to app. So we will see something like `15 20 APPKILL 10 15 35`.
In this case everything will be ok. I promise :)

# TODO
- [ ] support labels for metrics
- [ ] fix first scrape bug (zero values)
- [x] collect fails with counter (fails_counter{name="job_name"})
- [ ] better examples
- [ ] standard goapp layout
- [ ] histogram support
- [ ] summary support
- Auth support:
  - [ ] basic
  - [ ] token
  - [ ] custom HTTP header
- [ ] custom user agent

# Maybe
- [ ] cron rules for scraper
- [ ] json data aggregation
- [ ] db data scrape
- [ ] take data from local files
