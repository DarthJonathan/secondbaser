package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-uuid"
	"github.com/openzipkin/zipkin-go"
	api "github.com/trakkie-id/secondbaser/api/go_gen"
	"github.com/trakkie-id/secondbaser/model"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func TransactionInitTemplate(ctx context.Context, t *zipkin.Tracer, businessType string, bizId string, process func(injectedContext context.Context) error) error {
	span, _ := t.StartSpanFromContext(ctx, "Start Secondbaser Initiator First Stage")
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

	bizCtxJson,_ := json.Marshal(bizCtx)
	bizAddedCtx := metadata.AppendToOutgoingContext(
		ctx, "secondbaser-biz-trx-context", string(bizCtxJson),
	)

	//Store to DB
	go func() {
		transactionParent := &model.Transaction{
			TransactionId:     trxId,
			InitiatorSystem:   AppName,
			TransactionStatus: model.TRX_INIT,
			BusinessId:        bizId,
			BusinessType:      businessType,
		}

		DB.Create(transactionParent)
	}()

	notifyServerStart(bizAddedCtx, *bizCtx)
	processErr := process(bizAddedCtx)
	bizCtx.FinishPhaseTime = time.Now()

	//Generate new context for asynchronous processing
	asyncContext := metadata.AppendToOutgoingContext(
		context.Background(), "secondbaser-biz-trx-context", string(bizCtxJson),
	)
	asyncContext = zipkin.NewContext(asyncContext, zipkin.SpanFromContext(ctx))

	if processErr != nil {
		span.Tag(string(zipkin.TagError), fmt.Sprint(processErr))

		//Do rollback to clients async
		go func() {
			bizCtx.ActionType = ACTION_TYPE_ROLLBACK
			notifyServerFinal(asyncContext, *bizCtx)

			//Update database
			DB.Where(&model.Transaction{TransactionId: trxId}).Updates(
				&model.Transaction{
					TransactionStatus: model.TRX_ROLLBACK,
				})
		}()
		return processErr
	}

	//Do commit to clients async
	go func() {
		bizCtx.ActionType = ACTION_TYPE_COMMIT
		notifyServerFinal(asyncContext, *bizCtx)

		//Update database
		DB.Where(&model.Transaction{TransactionId: trxId}).Updates(
			&model.Transaction{
				TransactionStatus: model.TRX_COMMIT,
			})
	}()

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
		BizId: 			   transactionContext.BusinessId,
		BizType: 		   transactionContext.BusinessType,
	}

	client := api.NewTransactionalRequestClient(grpcCon)
	_, err = client.StartTransaction(ctx, requestParsed)

	if err != nil {
		LOGGER.Errorf("[SERVER] failed to send transaction start message: %s", err)
	}

	LOGGER.Infof("[SERVER] notified manager for ongoing transaction, trx id : %s", transactionContext.TransactionId)
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

	LOGGER.Infof("[SERVER] notified manager for ongoing transaction, status : %s, trx id : %s", transactionContext.ActionType, transactionContext.TransactionId)
	CloseConn(grpcCon)
}