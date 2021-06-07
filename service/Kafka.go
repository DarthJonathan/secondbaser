package service

import (
	"context"
	"github.com/DarthJonathan/secondbaser/config/application"
	"github.com/openzipkin/zipkin-go"
	"github.com/segmentio/kafka-go"
)

func PublishMessage(context context.Context, topic string, payload []byte) error {
	//Try create topic first
	_, err := kafka.DialLeader(context, "tcp", "localhost:9092", topic, 0)
	if err != nil {
		application.LOGGER.Errorf("failed to create topic: %s", err)
		return err
	}


	w := &kafka.Writer{
		Addr:     kafka.TCP(application.KafkaBroker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	//Put tracing data
	spanID := zipkin.SpanFromContext(context).Context().ID.String()
	traceID := zipkin.SpanFromContext(context).Context().TraceID.String()
	headers := []kafka.Header{
		{
			Key:   "X-B3-TraceId",
			Value: []byte(traceID),
		},
		{
			Key:   "X-B3-SpanId",
			Value: []byte(spanID),
		},
		{
			Key:   "X-B3-Sampled",
			Value: []byte("1"),
		},
	}

	err = w.WriteMessages(
		context,
		kafka.Message{
			Value: payload,
			Headers: headers,
		},
	)
	if err != nil {
		application.LOGGER.Errorf("failed to write messages: %s", err)
		return err
	}

	application.LOGGER.Infof("[KAFKA] Message Sent! Topic : %s Payload: %s", topic, string(payload))
	return nil
}
