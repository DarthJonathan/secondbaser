package service

import (
	"context"
	"github.com/segmentio/kafka-go"
	kafka_zipkin_interceptor "github.com/trakkie-id/kafka-zipkin-interceptor"
	"github.com/trakkie-id/secondbaser/config/application"
)

func PublishMessage(context context.Context, topic string, payload []byte) error {
	//Try create topic first
	_, err := kafka.DialLeader(context, "tcp", application.KafkaBroker, topic, 0)
	if err != nil {
		application.LOGGER.Errorf("failed to create topic: %s", err)
		return err
	}

	//Put tracing data
	headers, ctx := kafka_zipkin_interceptor.WrapInSpan("", topic, "","", context, application.TRACER)

	w := &kafka.Writer{
		Addr:     kafka.TCP(application.KafkaBroker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	err = w.WriteMessages(
		ctx,
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
