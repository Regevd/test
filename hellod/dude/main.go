package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	pb "github.com/traiana/okro/hellod/api/hello/v1"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const svcname = "dude"

var log = &logrus.Logger{
	Out:       os.Stderr,
	Formatter: &logrus.JSONFormatter{},
	Hooks:     make(logrus.LevelHooks),
	Level:     logrus.InfoLevel,
}

type config struct {
	Port        int    `mapstructure:"port"`
	MetricsPort int    `mapstructure:"metrics-port"`
	HelloerAddr string `mapstructure:"helloer-addr"`
	EchoerAddr  string `mapstructure:"echoer-addr"`
	JaegerAddr  string `mapstructure:"jaeger-addr"`
}

func loadConfig() (*config, error) {
	pflag.Int("port", 8080, "http port")
	pflag.Int("metrics-port", 8081, "metrics http port")
	pflag.String("helloer-addr", "localhost:5050", "helloer address")
	pflag.String("echoer-addr", "localhost:6060", "echoer address")
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

	grpcOpts := []grpc.DialOption{
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{
			StartOptions: trace.StartOptions{
				Sampler:  trace.AlwaysSample(),
				SpanKind: trace.SpanKindClient,
			},
		}),
		grpc.WithInsecure(),
	}
	hconn, err := grpc.Dial(cfg.HelloerAddr, grpcOpts...)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer hconn.Close()
	hc := pb.NewHelloClient(hconn)

	econn, err := grpc.Dial(cfg.EchoerAddr, grpcOpts...)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer econn.Close()
	ec := pb.NewEchoClient(econn)

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

	r.Get("/hi/{name}", hiHandler(hc, ec))
	r.Get("/bye/{name}", byeHandler(hc))
	r.Get("/echo/{name}", echoHandler(ec))

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
	if err := view.Register(ocgrpc.DefaultClientViews...); err != nil {
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

func hiHandler(hc pb.HelloClient, ec pb.EchoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		var req = &pb.HiRequest{Name: name}
		var res *pb.HiResponse

		g := errgroup.Group{}
		g.Go(func() error {
			hres, err := hc.Hi(r.Context(), req)
			res = hres
			return err
		})
		g.Go(func() error {
			_, err := ec.Echo(r.Context(), req)
			return err
		})
		err := g.Wait()

		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.PlainText(w, r, fmt.Sprintf("failed to call method: %v", err))
			return
		}
		render.JSON(w, r, render.M{"message": res.Message})
	}
}

func byeHandler(hc pb.HelloClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		res, err := hc.Bye(r.Context(), &pb.ByeRequest{Name: name})
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.PlainText(w, r, fmt.Sprintf("failed to call method: %v", err))
			return
		}
		render.JSON(w, r, render.M{"message": res.Message})
	}
}

func echoHandler(hc pb.EchoClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		res, err := hc.Echo(r.Context(), &pb.HiRequest{Name: name})
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.PlainText(w, r, fmt.Sprintf("failed to call method: %v", err))
			return
		}

		lg.RequestLog(r).Infoln("detached echo will run in 3 seconds")
		time.AfterFunc(3*time.Second, func() {
			echoed := strings.Repeat(name, 3)
			_, err := hc.Echo(context.Background(), &pb.HiRequest{Name: echoed})
			if err != nil {
				log.Warnf("detached echo failed: %v\n", err)
			}
		})

		render.JSON(w, r, render.M{"name": res.Name})
	}
}
