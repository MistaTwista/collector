# Collector
Make HTTP requests and prepare data for prometheus scraper.
Currently Counter and Gauge supported.

# Flags
- `-c config.yaml` - config with work description
- `-l :8080` - where http server listen

# Usage
1. Make config ([example](examples/config.example.yaml))
2. Run collector
3. Add [Prometheus](https://prometheus.io/) scrape work to `http://yourserver/metrics`

# Config
Currently only JSON can be used.
I use [GJSON](https://github.com/tidwall/gjson) to get data from received JSON responses, so your config might looks like:
```yaml
works:
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
You can play with json in console:
```bash
$ make build
$ bin/json-play -json examples/data1.example.json -r data.1.nets.#
data.1.nets.# => 2
```

## Gauge
Works same as in usual Prometheus workflow, can increase and decrease

## Counter
Prometheus counters works only as increasing value, so do we:

### Example
You app data have metrics: `20 25 35 APPKILL 10 15 15 15 20`

Where **APPKILL** - when your app stopped for some reason and metric started from 0

When there was no changes in time - we did nothing (like 15,15,15 in example, we count just first 15)

As a reult our final metric after last values check will be: 20+5+10+10+5+5 = 55 (or just 35, KILL, 20)

Of course there can be situations like `20 APPKILL 35` so Collector counts 35 instead of 55, because it will not know if app was killed.
Workflow for such problems can be more frequent requests to app. So we will see something like `15 20 APPKILL 10 15 35`.
In this case everything will be ok. I promise :)

# TODO
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
