package main

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	pb "github.com/traiana/okro/hellod/api/hello/v1"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var kafkaTraceOptions = []trace.StartOption{
	trace.WithSampler(trace.AlwaysSample()),
	trace.WithSpanKind(trace.SpanKindClient),
}

type server struct {
	producer   sarama.SyncProducer
	errorTopic string
}

func NewServer(kp sarama.SyncProducer, errorTopic string) *server {
	return &server{
		producer:   kp,
		errorTopic: errorTopic,
	}
}

func (s *server) Echo(ctx context.Context, in *pb.HiRequest) (*pb.HiRequest, error) {
	if in.Name != "666" {
		log.Infof("echoing %s", in.Name)
		return in, nil
	}

	log.Warnln("trap card activated, returning 400")
	_, span := trace.StartSpan(ctx, fmt.Sprintf("kafka/%s", s.errorTopic), kafkaTraceOptions...)
	msg := &sarama.ProducerMessage{
		Topic: s.errorTopic,
		Key:   strEnc(in.Name),
		Value: protoEnc(in),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("trace-bin"),
				Value: propagation.Binary(span.SpanContext()),
			},
		},
	}
	_, _, err := s.producer.SendMessage(msg)
	if err != nil {
		errmsg := fmt.Sprintf("failed to produce to kafka: %v", err)
		log.Errorln(errmsg)
		span.SetStatus(trace.Status{
			Code:    trace.StatusCodeCancelled,
			Message: errmsg,
		})
	}
	span.End()
	return nil, status.Errorf(codes.InvalidArgument, "oh no, it's the trap card")
}

func strEnc(s string) sarama.Encoder {
	return sarama.StringEncoder(s)
}

func protoEnc(m proto.Message) sarama.Encoder {
	bbs, _ := proto.Marshal(m) // yolo
	return sarama.ByteEncoder(bbs)
}
