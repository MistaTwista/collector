package main

import (
	"flag"
	"log"
	"net/http"
	"time"
	"os"
	"sync"
	"strconv"

	"collector/config"
	httpClient "collector/internal/http"
	"collector/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tidwall/gjson"
)

func main() {
	cfgPath := flag.String("c", "config.yaml", "config file path")
	addr := flag.String("l", ":8080", "Listen address")
	flag.Parse()

	cdata, err := os.Open(*cfgPath)
	if err != nil {
		log.Fatal("cannot read config", err)
	}
	defer cdata.Close()

	c, err := config.Load(cdata)
	if err != nil {
		log.Printf("cannot load config: %s", err)
		return
	}

	var wg sync.WaitGroup
	for _, w := range c.Works {
		log.Printf("Work: %s, %s from %s every %s with delay %s\n", w.Name, w.Method, w.Url, w.Every, w.Delay)
		wg.Add(1)
		go func(wrk config.Work) {
			defer wg.Done()
			failCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
					Namespace: wrk.Namespace,
					Subsystem: wrk.Subsystem,
					Name: "fails_total",
					Help: "Scrape fail counter",
				},
				[]string{"name"},
			)
			prometheus.MustRegister(failCounter)

			log.Print("prepare metrics collectors")
			metricMap := metrics.NewMetrics(wrk)

			var once sync.Once
			for {
				once.Do(func() {
					time.Sleep(wrk.Delay)
				})
				log.Printf("Run work %s", wrk.Name)

				data, err := httpClient.GetData(wrk.Url)
				if err != nil {
					log.Print(err)
					time.Sleep(5 * time.Second)
					failCounter.WithLabelValues(wrk.Name).Inc()
					continue
				}

				for _, m := range wrk.Mapping {
					val := gjson.GetBytes(data, m.Req)
					f, err := strconv.ParseFloat(val.String(), 64)
					if err != nil {
						log.Println("cannot parse value", err)
						failCounter.WithLabelValues(string(m.Name)).Inc()
						continue
					}

					metricMap[m.Name].Set(f)
				}

				time.Sleep(wrk.Every)
			}
		}(w)
	}

	log.Printf("Run metrics server at %s", *addr)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))

	wg.Wait()
}
