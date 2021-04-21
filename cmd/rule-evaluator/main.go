// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/gpe-collector/pkg/export"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
	apihttp "google.golang.org/api/transport/http"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/notifier"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/rules"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/strutil"
)

// convertValueToVector converts model.Value type to promql.Vector type.
func convertValueToVector(val model.Value) (promql.Vector, error) {
	results, ok := val.(model.Vector)
	if !ok {
		return nil, errors.Errorf("Expected Prometheus results of type vector. Actual results type: %v\n", results.Type())
	}
	v := make(promql.Vector, len(results))
	for i, result := range results {
		ls := make(labels.Labels, 0, len(result.Metric))
		for name, value := range result.Metric {
			l := labels.Label{
				Name:  string(name),
				Value: string(value),
			}
			ls = append(ls, l)
		}
		s := promql.Sample{
			Point: promql.Point{
				T: int64(result.Timestamp),
				V: float64(result.Value),
			},
			Metric: ls,
		}
		v[i] = s
	}
	return v, nil
}

type flagConfig struct {
	configFile    string
	targetURL     string
	listenAddress string
	notifier      notifier.Options
}

// QueryFunc queries a Prometheus instance and returns a promql.Vector.
func QueryFunc(ctx context.Context, q string, t time.Time, v1api v1.API) (promql.Vector, v1.Warnings, error) {
	results, warnings, err := v1api.Query(ctx, q, time.Now())
	if err != nil {
		return nil, warnings, errors.Errorf("Error querying Prometheus: %v\n", err)
	}
	v, err := convertValueToVector(results)
	return v, warnings, err
}

// sendAlerts returns the rules.NotifyFunc for a Notifier.
func sendAlerts(s *notifier.Manager, externalURL string) rules.NotifyFunc {
	return func(ctx context.Context, expr string, alerts ...*rules.Alert) {
		var res []*notifier.Alert
		for _, alert := range alerts {
			a := &notifier.Alert{
				StartsAt:     alert.FiredAt,
				Labels:       alert.Labels,
				Annotations:  alert.Annotations,
				GeneratorURL: externalURL + strutil.TableLinkForExpression(expr),
			}
			if !alert.ResolvedAt.IsZero() {
				a.EndsAt = alert.ResolvedAt
			} else {
				a.EndsAt = alert.ValidUntil
			}
			res = append(res, a)
		}
		if len(alerts) > 0 {
			s.Send(res...)
		}
	}
}

type reloader struct {
	name     string
	reloader func(*config.Config) error
}

// reloadConfig applies the configuration files.
func reloadConfig(filename string, logger log.Logger, rls ...reloader) (err error) {
	start := time.Now()
	timings := []interface{}{}
	level.Info(logger).Log("msg", "Loading configuration file", "filename", filename)

	conf, err := config.LoadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "couldn't load configuration (--config.file=%q)", filename)
	}

	failed := false
	for _, rl := range rls {
		rstart := time.Now()
		if err := rl.reloader(conf); err != nil {
			level.Error(logger).Log("msg", "Failed to apply configuration", "err", err)
			failed = true
		}
		timings = append(timings, rl.name, time.Since(rstart))
	}
	if failed {
		return errors.Errorf("one or more errors occurred while applying the new configuration (--config.file=%q)", filename)
	}

	l := []interface{}{"msg", "Completed loading of configuration file", "filename", filename, "totalDuration", time.Since(start)}
	level.Info(logger).Log(append(l, timings...)...)
	return nil
}

func main() {
	cfg := flagConfig{}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	a := kingpin.New("rule", "The Prometheus Rule Evaluator")
	exporterOptions := export.NewFlagOptions(a)

	a.Flag("query.target-url", "The address of the Prometheus server query endpoint.").
		Required().StringVar(&cfg.targetURL)

	a.Flag("web.listen-address", "The address to listen on for HTTP requests.").
		Default(":9091").StringVar(&cfg.listenAddress)

	a.Flag("config.file", "Prometheus configuration file path.").
		Default("prometheus.yml").StringVar(&cfg.configFile)

	a.Flag("alertmanager.notification-queue-capacity", "The capacity of the queue for pending Alertmanager notifications.").
		Default("10000").IntVar(&cfg.notifier.QueueCapacity)

	if _, err := a.Parse(os.Args[1:]); err != nil {
		level.Error(logger).Log("msg", "Error parsing commandline arguments", "err", err)
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	if _, err := config.LoadFile(cfg.configFile); err != nil {
		level.Error(logger).Log("msg", fmt.Sprintf("Error loading config (--config.file=%s)", cfg.configFile), "err", err)
		os.Exit(2)
	}

	destination, err := export.NewStorage(logger, nil, *exporterOptions)
	if err != nil {
		level.Error(logger).Log("msg", "Creating a Cloud Monitoring Exporter failed", "err", err)
		os.Exit(1)
	}

	noopQueryable := func(ctx context.Context, mint, maxt int64) (storage.Querier, error) {
		return storage.NoopQuerier(), nil
	}

	ctxRuleManger := context.Background()
	ctxDiscover, cancelDiscover := context.WithCancel(context.Background())

	opts := []option.ClientOption{
		option.WithScopes("https://www.googleapis.com/auth/monitoring.read"),
		option.WithCredentialsFile(exporterOptions.CredentialsFile),
	}
	transport, err := apihttp.NewTransport(ctxRuleManger, http.DefaultTransport, opts...)
	if err != nil {
		level.Error(logger).Log("msg", "Creating proxy HTTP transport failed", "err", err)
		os.Exit(1)
	}
	client, err := api.NewClient(api.Config{
		Address:      cfg.targetURL,
		RoundTripper: transport,
	})
	if err != nil {
		level.Error(logger).Log("msg", "Error creating client", "err", err)
		os.Exit(1)
	}
	v1api := v1.NewAPI(client)

	queryFunc := func(ctx context.Context, q string, t time.Time) (promql.Vector, error) {
		v, warnings, err := QueryFunc(ctx, q, t, v1api)
		if len(warnings) > 0 {
			level.Warn(logger).Log("msg", "Querying Promethues instance returned warnings", "warn", warnings)
		}
		return v, err
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)

	discoveryManager := discovery.NewManager(ctxDiscover, log.With(logger, "component", "discovery manager notify"), discovery.Name("notify"))
	notificationManager := notifier.NewManager(&cfg.notifier, log.With(logger, "component", "notifier"))

	ruleManager := rules.NewManager(&rules.ManagerOptions{
		ExternalURL: &url.URL{},
		QueryFunc:   queryFunc,
		Context:     ctxRuleManger,
		Appendable:  destination,
		Queryable:   storage.QueryableFunc(noopQueryable),
		Logger:      logger,
		NotifyFunc:  sendAlerts(notificationManager, cfg.targetURL),
		Metrics:     rules.NewGroupMetrics(reg),
	})

	reloaders := []reloader{
		{
			name:     "notify",
			reloader: notificationManager.ApplyConfig,
		}, {
			name: "notify_sd",
			reloader: func(cfg *config.Config) error {
				c := make(map[string]discovery.Configs)
				for k, v := range cfg.AlertingConfig.AlertmanagerConfigs.ToMap() {
					c[k] = v.ServiceDiscoveryConfigs
				}
				return discoveryManager.ApplyConfig(c)
			},
		}, {
			name: "rules",
			reloader: func(cfg *config.Config) error {
				// Get all rule files matching the configuration paths.
				var files []string
				for _, pat := range cfg.RuleFiles {
					fs, err := filepath.Glob(pat)
					if fs == nil || err != nil {
						return errors.Errorf("Error retrieving rule file: %s", pat)
					}
					files = append(files, fs...)
				}
				return ruleManager.Update(
					time.Duration(cfg.GlobalConfig.EvaluationInterval),
					files,
					cfg.GlobalConfig.ExternalLabels,
				)
			},
		},
	}

	var g run.Group
	{
		// Termination handler.
		term := make(chan os.Signal, 1)
		cancel := make(chan struct{})
		signal.Notify(term, os.Interrupt, syscall.SIGTERM)
		g.Add(
			func() error {
				select {
				case <-term:
					level.Info(logger).Log("msg", "received SIGTERM, exiting gracefully...")
				case <-cancel:
				}
				return nil
			},
			func(err error) {
				close(cancel)
			},
		)
	}
	{
		// Rule manager.
		g.Add(func() error {
			ruleManager.Run()
			return nil
		}, func(error) {
			ruleManager.Stop()
		})
	}
	{
		// Notifier.
		g.Add(func() error {
			notificationManager.Run(discoveryManager.SyncCh())
			level.Info(logger).Log("msg", "Notification manager stopped")
			return nil
		},
			func(err error) {
				notificationManager.Stop()
			},
		)
	}
	{
		// Notify discovery manager.
		g.Add(
			func() error {
				err := discoveryManager.Run()
				level.Info(logger).Log("msg", "Discovery manager stopped")
				return err
			},
			func(err error) {
				level.Info(logger).Log("msg", "Stopping Discovery manager...")
				cancelDiscover()
			},
		)
	}
	{
		// Storage Processing.
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			err = destination.Run(ctx)
			level.Info(logger).Log("msg", "Background processing of storage stopped")
			return err
		}, func(error) {
			level.Info(logger).Log("msg", "Stopping background storage processing...")
			cancel()
		})
	}
	reloadCh := make(chan chan error)
	{
		// Web Server.
		server := &http.Server{Addr: cfg.listenAddress}
		http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
		http.HandleFunc("/-/reload", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				rc := make(chan error)
				reloadCh <- rc
				if err := <-rc; err != nil {
					http.Error(w, fmt.Sprintf("Failed to reload config: %s", err), http.StatusInternalServerError)
				}
			} else {
				http.Error(w, "Only POST requests allowed.", http.StatusMethodNotAllowed)
			}
		})
		http.HandleFunc("/-/healthy", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		http.HandleFunc("/-/ready", func(w http.ResponseWriter, r *http.Request) {
			_, err := queryFunc(context.Background(), "vector(1)", time.Now()) // Run a test query to check status of rule evaluator.
			if err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprintf(w, "Prometheus is not Ready.\n")
				level.Error(logger).Log("msg", "Error querying Prometheus instance", "err", err)
			} else {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "Prometheus is Ready.\n")
			}
		})
		g.Add(func() error {
			level.Info(logger).Log("msg", "Starting web server", "listen", cfg.listenAddress)
			return server.ListenAndServe()
		}, func(err error) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			server.Shutdown(ctx)
			cancel()
		})
	}
	{
		// Reload handler.
		hup := make(chan os.Signal, 1)
		signal.Notify(hup, syscall.SIGHUP)
		cancel := make(chan struct{})
		g.Add(
			func() error {
				for {
					select {
					case <-hup:
						if err := reloadConfig(cfg.configFile, logger, reloaders...); err != nil {
							level.Error(logger).Log("msg", "Error reloading config", "err", err)
						}
					case rc := <-reloadCh:
						if err := reloadConfig(cfg.configFile, logger, reloaders...); err != nil {
							level.Error(logger).Log("msg", "Error reloading config", "err", err)
							rc <- err
						} else {
							rc <- nil
						}
					case <-cancel:
						return nil
					}
				}
			},
			func(err error) {
				// Wait for any in-progress reloads to complete to avoid
				// reloading things after they have been shutdown.
				cancel <- struct{}{}
			},
		)
	}
	{
		// Initial configuration loading.
		cancel := make(chan struct{})
		g.Add(
			func() error {
				if err := reloadConfig(cfg.configFile, logger, reloaders...); err != nil {
					level.Info(logger).Log("msg", "error loading config file.")
					return errors.Wrapf(err, "error loading config from %q", cfg.configFile)
				}
				<-cancel
				return nil
			},
			func(err error) {
				close(cancel)
			},
		)
	}

	if err := g.Run(); err != nil {
		level.Error(logger).Log("msg", "Running rule evaluator failed", "err", err)
		os.Exit(1)
	}
}
