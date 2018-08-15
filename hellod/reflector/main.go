package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"io"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

const svcname = "reflector"

var log = &logrus.Logger{
	Out:       os.Stderr,
	Formatter: &logrus.JSONFormatter{},
	Hooks:     make(logrus.LevelHooks),
	Level:     logrus.InfoLevel,
}

type config struct {
	Port          int    `mapstructure:"port"`
	MetricsPort   int    `mapstructure:"metrics-port"`
	ReflectorAddr string `mapstructure:"reflector-addr"`
	DudeAddr      string `mapstructure:"dude-addr"`
	JaegerAddr    string `mapstructure:"jaeger-addr"`
}

func loadConfig() (*config, error) {
	pflag.Int("port", 4040, "http port")
	pflag.Int("metrics-port", 4041, "metrics http port")
	pflag.String("reflector-addr", "localhost:4040", "other reflector address")
	pflag.String("dude-addr", "localhost:8080", "dude address")
	pflag.String("jaeger-addr", "http://localhost:14268", "jaeger address")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
	viper.SetConfigFile("/etc/okro/config.yaml")
	viper.ReadInConfig()

	var cfg config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	promExp, err := initMetrics()
	if err != nil {
		log.Fatalf("failed to init metrics: %v", err)
	}

	_, err = initTracing(cfg)
	if err != nil {
		log.Fatalf("failed to init tracing: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(lg.RequestLogger(log))
	lg.RedirectStdlogOutput(log)
	lg.DefaultLogger = log

	client := &http.Client{Transport: &ochttp.Transport{
		Propagation: &tracecontext.HTTPFormat{},
		StartOptions: trace.StartOptions{
			Sampler:  trace.AlwaysSample(),
			SpanKind: trace.SpanKindClient,
		},
	}}

	handler := reflector(cfg, client)
	r.Get("/hi/*", handler)
	r.Get("/bye/*", handler)

	och := &ochttp.Handler{
		Propagation: &tracecontext.HTTPFormat{},
		Handler:     r,
		StartOptions: trace.StartOptions{
			Sampler:  trace.AlwaysSample(),
			SpanKind: trace.SpanKindServer,
		},
	}

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promExp)
		smport := fmt.Sprintf(":%v", cfg.MetricsPort)
		log.Infof("serving metrics at %s", smport)
		log.Fatal(http.ListenAndServe(smport, mux))
	}()

	sport := fmt.Sprintf(":%v", cfg.Port)
	log.Infof("serving http at %s", sport)
	log.Fatal(http.ListenAndServe(sport, och))
}

func initMetrics() (*prometheus.Exporter, error) {
	if err := view.Register(ochttp.DefaultServerViews...); err != nil {
		return nil, err
	}
	if err := view.Register(ochttp.DefaultClientViews...); err != nil {
		return nil, err
	}

	exp, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		return nil, err
	}
	view.RegisterExporter(exp)
	return exp, nil
}

func initTracing(cfg *config) (*jaeger.Exporter, error) {
	exp, err := jaeger.NewExporter(jaeger.Options{
		Endpoint:    cfg.JaegerAddr,
		ServiceName: svcname,
	})
	if err != nil {
		return nil, err
	}
	trace.RegisterExporter(exp)
	return exp, nil
}

func reflector(cfg *config, client *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var target string
		if r.URL.Query().Get("reflect") == "true" {
			lg.RequestLog(r).Infof("reflecting %s", r.URL.Path)
			target = cfg.ReflectorAddr
		} else {
			lg.RequestLog(r).Infof("forwarding %s", r.URL.Path)
			target = cfg.DudeAddr
		}

		targetURL := fmt.Sprintf("http://%s%s", target, r.URL.Path)
		req, err := http.NewRequest("GET", targetURL, nil)
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.PlainText(w, r, fmt.Sprintf("failed to create reflection: %v", err))
			return
		}

		res, err := client.Do(req.WithContext(r.Context()))
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.PlainText(w, r, fmt.Sprintf("failed to reflect: %v", err))
			return
		}

		w.Header().Set("Content-Type", res.Header.Get("Content-Type"))
		w.WriteHeader(res.StatusCode)
		io.Copy(w, res.Body)
		res.Body.Close()
	}
}
