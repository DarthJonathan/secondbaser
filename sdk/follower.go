package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/openzipkin/zipkin-go"
	zipkinModel "github.com/openzipkin/zipkin-go/model"
	"github.com/segmentio/kafka-go"
	"github.com/trakkie-id/secondbaser/model"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
)

var span zipkin.Span

func FollowTransactionTemplate(ctx context.Context, process func() error, rollback func(bizContext BusinessTransactionContext) error, forward func(bizContext BusinessTransactionContext) error) error {
	span, _ = TRACER.StartSpanFromContext(ctx, "Start SECONDBASER Follower First Stage")
	span.Tag("SECONDBASER", "First Stage Follower")
	SetLogFormat(ctx)

	//Load business context from context
	metaDataCtx, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errors.New("unable to get metadata from context")
	}

	trxContextGroup := metaDataCtx["secondbaser-biz-trx-context"]
	if trxContextGroup == nil || len(trxContextGroup) < 1 || trxContextGroup[0] == "" {
		LOGGER.Debugf("Metadata payload : [%+v]", metaDataCtx)
		return errors.New("unable to get business transaction context from context")
	}
	trxContext := trxContextGroup[0]
	businessTrxContext := &BusinessTransactionContext{}
	err := json.Unmarshal([]byte(trxContext), &businessTrxContext)

	if err != nil {
		LOGGER.Errorf("Error in parsing business transaction context data, err : %+v", err)
		return errors.New("unable to get business transaction context from context")
	}

	//Save to db
	trxFollowerDO := &model.TransactionParticipant{
		TransactionId:     businessTrxContext.TransactionId,
		ParticipantSystem: AppName,
		ParticipantStatus: model.TRX_INIT,
	}
	resErr := DB.Create(trxFollowerDO)

	if resErr.Error != nil && !errors.Is(resErr.Error, gorm.ErrRecordNotFound) {
		LOGGER.Errorf("Unable to store transaction, err : %+v", resErr.Error)
	}

	processErr := process()

	if processErr != nil {
		span.Tag(string(zipkin.TagError), fmt.Sprint(processErr))
		return processErr
	}

	//Finish 1st Span
	span.Finish()

	//Load kafka
	topic := SECONDBASER_PREFIX_TOPIC + businessTrxContext.BusinessType + businessTrxContext.Initiator
	go listenToKafkaMsg(topic, businessTrxContext.TransactionId, rollback, forward)

	return err
}

func listenToKafkaMsg(topic string, trxId string, rollback func(bizContext BusinessTransactionContext) error, forward func(bizContext BusinessTransactionContext) error) {
	LOGGER.Debugf("[KAFKA] Waiting for final phase [Topic : %s]", topic)

	// make a new reader that consumes from topic-A
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{KafkaAddress},
		GroupID:  KafkaGroupId,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			LOGGER.Errorf("[KAFKA] Unable start reader, err : %+v", err)
			break
		}

		bizContext := &BusinessTransactionContext{}
		err = json.Unmarshal(m.Value, bizContext)

		if err != nil {
			LOGGER.Errorf("[KAFKA] Unable to parse payload, err : %+v", err)
		}

		//Validate same trx id
		if trxId != bizContext.TransactionId {
			LOGGER.Debugf("[KAFKA] Skipping message, trx id not matched! trx id : %v", bizContext.TransactionId)
			continue
		}

		LOGGER.Infof("[KAFKA] Received SECONDBASER Phase Two Message [Topic : %s, Payload: %+v]", m.Topic, string(m.Value))

		traceId := ""
		spanId := ""

		for _, header := range m.Headers {
			if header.Key == "X-B3-TraceId" {
				traceId = string(header.Value)
			} else if header.Key == "X-B3-SpanId" {
				spanId = string(header.Value)
			}
		}

		unitSpanId, _ := strconv.ParseUint(spanId, 0, 64)
		traceIdModel, _ := zipkinModel.TraceIDFromHex(traceId)

		spanContext := zipkinModel.SpanContext{
			TraceID: traceIdModel,
			ID:      zipkinModel.ID(unitSpanId),
		}

		span = TRACER.StartSpan("SECONDBASER Phase 2", zipkin.Parent(spanContext))
		LOGGER.SetFormat("%{time} [%{module}] [%{level}] [" + traceId + "," + spanId + "]  %{message}")
		LOGGER.Infof("SECONDBASER Phase Two Message Parse Result [%+v]", m.Topic, string(m.Value))

		trxFollowerDO := &model.TransactionParticipant{
			TransactionId:     bizContext.TransactionId,
			ParticipantSystem: AppName,
			ParticipantStatus: model.TRX_INIT,
		}

		if bizContext.ActionType == ACTION_TYPE_COMMIT {
			//Update to db
			resErr := DB.Model(trxFollowerDO).Updates(model.TransactionParticipant{
				ParticipantStatus: model.TRX_COMMIT,
			})

			if resErr.Error != nil && !errors.Is(resErr.Error, gorm.ErrRecordNotFound) {
				LOGGER.Errorf("Unable to store transaction, err : %+v", resErr.Error)
			}

			err = forward(*bizContext)
		} else {
			//Update to db
			resErr := DB.Model(trxFollowerDO).Updates(model.TransactionParticipant{
				ParticipantStatus: model.TRX_ROLLBACK,
			})

			if resErr.Error != nil && !errors.Is(resErr.Error, gorm.ErrRecordNotFound) {
				LOGGER.Errorf("Unable to store transaction, err : %+v", resErr.Error)
			}

			err = rollback(*bizContext)
		}

		if err != nil {
			span.Tag(string(zipkin.TagError), fmt.Sprint(err))
		}

		//Finish 2nd Span
		LOGGER.Infof("SECONDBASER Phase two finished with final status %v, and transaction ID : %s", bizContext.ActionType, bizContext.TransactionId)
		span.Finish()
	}

	if err := r.Close(); err != nil {
		LOGGER.Errorf("Unable to close reader, err : %+v", err)
	}
}
