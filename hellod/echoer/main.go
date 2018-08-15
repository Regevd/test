package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	pb "github.com/traiana/okro/hellod/api/hello/v1"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const svcname = "echoer"

var log = &logrus.Logger{
	Out:       os.Stderr,
	Formatter: &logrus.JSONFormatter{},
	Hooks:     make(logrus.LevelHooks),
	Level:     logrus.InfoLevel,
}

type config struct {
	Port         int      `mapstructure:"port"`
	MetricsPort  int      `mapstructure:"metrics-port"`
	KafkaBrokers []string `mapstructure:"kafka-brokers"`
	ErrorTopic   string   `mapstructure:"error-topic"`
	JaegerAddr   string   `mapstructure:"jaeger-addr"`
}

func loadConfig() (*config, error) {
	pflag.Int("port", 6060, "grpc port")
	pflag.Int("metrics-port", 6061, "metrics http port")
	pflag.StringSlice("kafka-brokers", []string{"localhost:9092"}, "comma-delimited kafka brokers")
	pflag.String("error-topic", "hellod.failures", "kafka error topic")
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

	kcfg := sarama.NewConfig()
	kcfg.Version = sarama.V1_1_0_0
	kcfg.ClientID = svcname
	kcfg.Producer.RequiredAcks = sarama.WaitForAll
	kcfg.Producer.Return.Errors = true
	kcfg.Producer.Return.Successes = true

	kp, err := sarama.NewSyncProducer(cfg.KafkaBrokers, kcfg)
	if err != nil {
		log.Fatalf("failed to init kafka producer: %v", err)
	}
	defer kp.Close()

	gs := grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{
		StartOptions: trace.StartOptions{
			Sampler:  trace.AlwaysSample(),
			SpanKind: trace.SpanKindServer,
		},
	}))

	impl := NewServer(kp, cfg.ErrorTopic)
	pb.RegisterEchoServer(gs, impl)

	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("hello.v1.Echo", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(gs, healthSrv)

	reflection.Register(gs)

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promExp)
		mports := fmt.Sprintf(":%v", cfg.MetricsPort)
		log.Infof("serving metrics at %s", mports)
		log.Fatal(http.ListenAndServe(mports, mux))
	}()

	sport := fmt.Sprintf(":%v", cfg.Port)
	lis, err := net.Listen("tcp", sport)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Infof("serving grpc at %s", sport)
	log.Fatal(gs.Serve(lis))
}

func initMetrics() (*prometheus.Exporter, error) {
	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
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
