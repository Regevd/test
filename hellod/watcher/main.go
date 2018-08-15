package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	pb "github.com/traiana/okro/hellod/api/hello/v1"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

const svcname = "watcher"

var log = &logrus.Logger{
	Out:       os.Stderr,
	Formatter: &logrus.JSONFormatter{},
	Hooks:     make(logrus.LevelHooks),
	Level:     logrus.InfoLevel,
}

type config struct {
	MetricsPort  int      `mapstructure:"metrics-port"`
	KafkaBrokers []string `mapstructure:"kafka-brokers"`
	ErrorTopic   string   `mapstructure:"error-topic"`
	JaegerAddr   string   `mapstructure:"jaeger-addr"`
}

func loadConfig() (*config, error) {
	pflag.Int("metrics-port", 7071, "metrics http port")
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

var (
	failuresReceived = stats.Int64(
		"io.okro.watcher/failures_received",
		"Failure instance received.",
		stats.UnitDimensionless,
	)

	failuresReceivedView = &view.View{
		Name:        "io.okro.watcher/failures_received",
		Description: "Count of failures received.",
		Measure:     failuresReceived,
		Aggregation: view.Count(),
	}
)

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

	kafkaTraceOptions := []trace.StartOption{
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithSpanKind(trace.SpanKindServer),
	}

	kcfg := sarama.NewConfig()
	kcfg.Version = sarama.V1_1_0_0
	kcfg.ClientID = svcname
	kcfg.Consumer.Offsets.CommitInterval = time.Minute

	kc, err := sarama.NewConsumer(cfg.KafkaBrokers, kcfg)
	if err != nil {
		log.Fatalf("failed to init kafka consumer: %v", err)
	}
	defer kc.Close()
	kpc, err := kc.ConsumePartition(cfg.ErrorTopic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Fatalf("failed to init kafka partition consumer: %v", err)
	}
	defer kpc.Close()

	go func() {
		ctx := context.TODO()
		for msg := range kpc.Messages() {
			req := &pb.HiRequest{}
			if err := proto.Unmarshal(msg.Value, req); err != nil {
				log.Errorf("failed to unmarshal proto: %v", err)
				continue
			}
			stats.Record(ctx, failuresReceived.M(1))
			var th *sarama.RecordHeader
			for _, h := range msg.Headers {
				if string(h.Key) == "trace-bin" {
					th = h
					break
				}
			}
			if th == nil {
				log.Warnln("failed to extract trace header")
				continue
			}

			sc, ok := propagation.FromBinary(th.Value)
			if !ok {
				log.Warnf("failed to parse trace header: %s", string(th.Value))
				continue
			}

			_, span := trace.StartSpanWithRemoteParent(ctx, fmt.Sprintf("kafka/%s", cfg.ErrorTopic), sc, kafkaTraceOptions...)
			log.Infof("consumed message: name=%s", req.Name)
			span.End()
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promExp)
	mports := fmt.Sprintf(":%v", cfg.MetricsPort)
	log.Infof("serving metrics at %s", mports)
	log.Fatal(http.ListenAndServe(mports, mux))
}

func initMetrics() (*prometheus.Exporter, error) {
	if err := view.Register(failuresReceivedView); err != nil {
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
