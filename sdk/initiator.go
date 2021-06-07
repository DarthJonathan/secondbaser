package sdk

import (
	"context"
	"fmt"
	api "github.com/DarthJonathan/secondbaser/api/go_gen"
	"github.com/hashicorp/go-uuid"
	"github.com/openzipkin/zipkin-go"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func TransactionInitTemplate(ctx context.Context, t *zipkin.Tracer, businessType string, bizId string, process func(injectedContext context.Context) error) error {
	span, _ := t.StartSpanFromContext(ctx, "Start Secondbaser Follower First Stage")
	defer span.Finish()
	span.Tag("SECONDBASER", "First Stage Initiator")

	uniqId,_ := uuid.GenerateUUID()

	trxId := businessType + "-" + AppName + "-" + uniqId

	bizCtx := &BusinessTransactionContext{
		Initiator: 		 AppName,
		TransactionId:   trxId,
		BusinessId:      bizId,
		BusinessType:    businessType,
		TransactionTime: time.Now(),
		FinishPhaseTime: time.Time{},
		ActionType:      "INIT",
	}

	bizAddedCtx := context.WithValue(ctx, "SECONDBASER-BIZ-TRX-CONTEXT", bizCtx)

	notifyServerStart(bizAddedCtx, *bizCtx)
	processErr := process(bizAddedCtx)
	bizCtx.FinishPhaseTime = time.Now()

	if processErr != nil {
		span.Tag(string(zipkin.TagError), fmt.Sprint(processErr))

		//Do rollback to clients
		bizCtx.ActionType = ACTION_TYPE_ROLLBACK
		notifyServerFinal(bizAddedCtx, *bizCtx)
		return processErr
	}

	//Do commit to clients
	bizCtx.ActionType = ACTION_TYPE_COMMIT
	notifyServerFinal(bizAddedCtx, *bizCtx)

	return nil
}

func notifyServerStart(ctx context.Context, transactionContext BusinessTransactionContext) {
	grpcCon, err := GetConn()

	if err != nil {
		LOGGER.Errorf("[SERVER] failed to connect to server: %s", err)
	}

	requestParsed := &api.TransactionRequest{
		TransactionId:     transactionContext.TransactionId,
		InitSystem:        transactionContext.Initiator,
		TransactionStart:  timestamppb.New(time.Now()),
		Success:           false,
	}

	client := api.NewTransactionalRequestClient(grpcCon)
	_, err = client.StartTransaction(ctx, requestParsed)

	if err != nil {
		LOGGER.Errorf("[SERVER] failed to send transaction start message: %s", err)
	}

	LOGGER.Infof("[SERVER] notified manager for ongoing transaction: %s", err)
	CloseConn(grpcCon)
}

func notifyServerFinal(ctx context.Context, transactionContext BusinessTransactionContext) {
	grpcCon, err := GetConn()

	if err != nil {
		LOGGER.Errorf("[SERVER] failed to connect to server: %s", err)
	}

	requestParsed := &api.TransactionRequest{
		TransactionId:     transactionContext.TransactionId,
		InitSystem:        transactionContext.Initiator,
	}
	client := api.NewTransactionalRequestClient(grpcCon)

	if transactionContext.ActionType == ACTION_TYPE_COMMIT {
		requestParsed.Success = true
		_, err = client.CommitTransaction(ctx, requestParsed)
	}else {
		requestParsed.Success = false
		_, err = client.RollbackTransaction(ctx, requestParsed)
	}

	if err != nil {
		LOGGER.Errorf("[SERVER] failed to send transaction final message: %s", err)
	}

	LOGGER.Infof("[SERVER] notified manager for finishing transaction: %s", err)
	CloseConn(grpcCon)
}