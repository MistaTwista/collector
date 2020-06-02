package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
	"os"
	"io"
	"strconv"

	"collector/config"
	httpClient "collector/internal/http"
	"collector/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tidwall/gjson"
)

func startWorker(job config.Job, ctx context.Context, w io.Writer) {
	log.SetOutput(w)
	failCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: job.Namespace,
			Subsystem: job.Subsystem,
			Name: fmt.Sprintf("%s_fails_total", job.Name),
			Help: "Scrape fail counter",
		},
		[]string{"name"},
	)
	prometheus.MustRegister(failCounter)

	log.Printf("prepare metrics collectors for %s", job.Name)
	metricMap := metrics.NewMetrics(job)
	delay := job.ScrapeDelay

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.Tick(delay):
			log.Printf("Run work %s\n", job.Name)
			delay = job.ScrapeInterval

			data, err := httpClient.GetData(job.Url)
			if err != nil {
				log.Print(err)
				time.Sleep(5 * time.Second)
				failCounter.WithLabelValues(job.Name).Inc()
				continue
			}

			log.Println("data ready, parse")
			for _, t := range job.Tasks {
				val := gjson.GetBytes(data, t.Req)
				f, err := strconv.ParseFloat(val.String(), 64)
				if err != nil {
					log.Println("cannot parse value", err)
					failCounter.WithLabelValues(string(t.Name)).Inc()
					continue
				}

				metricMap[t.Name].Set(f)
			}
		}
	}
}

func main() {
	cfgPath := flag.String("c", "config.yaml", "config file path")
	addr := flag.String("l", ":8080", "Listen address")
	flag.Parse()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()

	cdata, err := os.Open(*cfgPath)
	if err != nil {
		log.Fatalf("cannot read config: %s", err)
	}
	defer cdata.Close()

	c, err := config.Init(cdata)
	if err != nil {
		log.Fatalf("cannot init config: %s", err)
	}

	for _, j := range c.Jobs {
		log.Printf("Work: %s, from %s every %s, with delay: %s\n", j.Name, j.Url, j.ScrapeInterval, j.ScrapeDelay)
		go startWorker(j, ctx, os.Stdout)
	}

	log.Printf("Run metrics server at %s", *addr)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
