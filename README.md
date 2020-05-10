# Collector
Make HTTP requests and prepare data for prometheus scraper.
Currently Counter and Gauge supported.

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
