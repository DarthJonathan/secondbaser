package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/openzipkin/zipkin-go"
	api "github.com/trakkie-id/secondbaser/api/go_gen"
	"github.com/trakkie-id/secondbaser/config/application"
	"github.com/trakkie-id/secondbaser/model"
	"github.com/trakkie-id/secondbaser/sdk"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"time"
)

type transactionServiceImpl struct {
}

var (
	INIT_TRANSACTION 		= "SECONDBASER_INIT_TRANSACTION"
	COMMIT_TRANSACTION 		= "SECONDBASER_COMMIT_TRANSACTION"
	ROLLBACK_TRANSACTION 	= "SECONDBASER_ROLLBACK_TRANSACTION"
)

func setLogFormat(ctx context.Context) {
	spanID := zipkin.SpanFromContext(ctx).Context().ID.String()
	traceID := zipkin.SpanFromContext(ctx).Context().TraceID.String()
	application.LOGGER.SetFormat("%{time} [%{module}] [%{level}] [" + traceID +  "," + spanID + "]  %{message}")
}

func (t transactionServiceImpl) RegisterParticipant(ctx context.Context, request *api.TransactionRequest) (*api.TransactionResponse, error) {
	setLogFormat(ctx)
	application.LOGGER.Infof("Registered participant with payload [%+v]", request)

	if request.TransactionId == "" {
		application.LOGGER.Errorf("error illegal param, no trx id")
		return nil, errors.New("Fill transaction ID!")
	}

	if request.ParticipantSystem == "" {
		application.LOGGER.Errorf("error illegal param, no trx id")
		return nil, errors.New("Participant system cannot be empty!")
	}

	trx := &model.Transaction{}
	res := application.DB.Model(trx).Where(&model.Transaction{TransactionId: request.TransactionId}).First(trx)

	// check error ErrRecordNotFound
	if errors.Is(res.Error , gorm.ErrRecordNotFound) {
		application.LOGGER.Errorf("error, no such transaction")
		return nil, errors.New("no transaction exist!")
	}else if res.Error != nil {
		application.LOGGER.Errorf("error writing to db, err : ", res.Error)
		return nil, errors.New("Error writing to db!")
	}

	res = application.DB.Create(&model.TransactionParticipant{
		TransactionId:     trx.TransactionId,
		ParticipantSystem: request.ParticipantSystem,
		ParticipantStatus: model.TRX_INIT,
	})

	// check error ErrRecordNotFound
	if errors.Is(res.Error , gorm.ErrRecordNotFound) {
		application.LOGGER.Errorf("error, no such transaction")
		return nil, errors.New("no transaction exist!")
	}else if res.Error != nil {
		application.LOGGER.Errorf("error writing to db, err : ", res.Error)
		return nil, errors.New("Error writing to db!")
	}

	//Start transaction
	res = application.DB.Create(&model.TransactionParticipant{
		TransactionId:     request.TransactionId,
		ParticipantSystem: request.ParticipantSystem,
		ParticipantStatus: model.TRX_INIT,
	})

	if res.Error != nil {
		application.LOGGER.Errorf("error writing to db, err : ", res.Error)
		return nil, errors.New("Error writing to db!")
	}

	application.LOGGER.Infof("Registered participant success")
	return &api.TransactionResponse{
		TransactionStatus: string(trx.TransactionStatus),
		TransactionStart:  timestamppb.New(trx.CreatedAt),
		TransactionId:     request.TransactionId,
		InitSystem:        trx.InitiatorSystem,
	}, nil
}

func (t transactionServiceImpl) CommitTransaction(ctx context.Context, request *api.TransactionRequest) (*api.TransactionResponse, error) {
	setLogFormat(ctx)
	application.LOGGER.Infof("Commit transaction with payload [%+v]", request)

	if request.TransactionId == "" {
		application.LOGGER.Errorf("error illegal param, no trx id")
		return nil, errors.New("Fill transaction ID!")
	}

	trx := &model.Transaction{}
	res := application.DB.Model(trx).Where(&model.Transaction{TransactionId: request.TransactionId}).First(trx)

	// check error ErrRecordNotFound
	if errors.Is(res.Error , gorm.ErrRecordNotFound) {
		application.LOGGER.Errorf("error, no such transaction")
		return nil, errors.New("no transaction exist!")
	}else if res.Error != nil {
		application.LOGGER.Errorf("error writing to db, err : ", res.Error)
		return nil, errors.New("Error writing to db!")
	}

	res = application.DB.Where(&model.Transaction{TransactionId: request.TransactionId}).Updates(&model.Transaction{TransactionStatus: model.TRX_COMMIT})

	// check error ErrRecordNotFound
	if errors.Is(res.Error , gorm.ErrRecordNotFound) {
		application.LOGGER.Errorf("error, no such transaction")
		return nil, errors.New("no transaction exist!")
	}else if res.Error != nil {
		application.LOGGER.Errorf("error writing to db, err : ", res.Error)
		return nil, errors.New("Error writing to db!")
	}

	res = application.DB.Where(&model.TransactionParticipant{TransactionId: request.TransactionId}).Updates(&model.TransactionParticipant{ParticipantStatus: model.TRX_COMMIT})

	// check error ErrRecordNotFound
	if errors.Is(res.Error , gorm.ErrRecordNotFound) {
		application.LOGGER.Errorf("error, no such transaction")
		return nil, errors.New("no transaction exist!")
	}else if res.Error != nil {
		application.LOGGER.Errorf("error writing to db, err : ", res.Error)
		return nil, errors.New("Error writing to db!")
	}

	//Publish Trx for client
	topic := sdk.SECONDBASER_PREFIX_TOPIC + trx.BusinessType + "_" + trx.InitiatorSystem

	bizCtx := &sdk.BusinessTransactionContext{
		Initiator:       trx.InitiatorSystem,
		TransactionId:   trx.TransactionId,
		BusinessId:      trx.BusinessId,
		BusinessType:    trx.BusinessType,
		TransactionTime: trx.CreatedAt,
		FinishPhaseTime: time.Now(),
		ActionType:      sdk.ACTION_TYPE_COMMIT,
	}

	bizCtxJson,_ := json.Marshal(bizCtx)
	_ = PublishMessage(ctx, topic, bizCtxJson)

	application.LOGGER.Infof("Commit transaction finished")
	return &api.TransactionResponse{
		TransactionStatus: string(model.TRX_COMMIT),
		TransactionStart:  timestamppb.New(trx.CreatedAt),
		TransactionId:     request.TransactionId,
		InitSystem:        trx.InitiatorSystem,
	}, nil
}

func (t transactionServiceImpl) RollbackTransaction(ctx context.Context, request *api.TransactionRequest) (*api.TransactionResponse, error) {
	setLogFormat(ctx)
	application.LOGGER.Infof("Commit transaction with payload [%+v]", request)

	if request.TransactionId == "" {
		application.LOGGER.Errorf("error illegal param, no trx id")
		return nil, errors.New("Fill transaction ID!")
	}

	trx := &model.Transaction{}
	res := application.DB.Model(trx).Where(&model.Transaction{TransactionId: request.TransactionId}).First(trx)

	// check error ErrRecordNotFound
	if errors.Is(res.Error , gorm.ErrRecordNotFound) {
		application.LOGGER.Errorf("error, no such transaction")
		return nil, errors.New("no transaction exist!")
	}else if res.Error != nil {
		application.LOGGER.Errorf("error writing to db, err : ", res.Error)
		return nil, errors.New("Error writing to db!")
	}

	res = application.DB.Where(&model.Transaction{TransactionId: request.TransactionId}).Updates(&model.Transaction{TransactionStatus: model.TRX_ROLLBACK})

	// check error ErrRecordNotFound
	if errors.Is(res.Error , gorm.ErrRecordNotFound) {
		application.LOGGER.Errorf("error, no such transaction")
		return nil, errors.New("no transaction exist!")
	}else if res.Error != nil {
		application.LOGGER.Errorf("error writing to db, err : ", res.Error)
		return nil, errors.New("Error writing to db!")
	}

	res = application.DB.Where(&model.TransactionParticipant{TransactionId: request.TransactionId}).Updates(&model.TransactionParticipant{ParticipantStatus: model.TRX_ROLLBACK})

	// check error ErrRecordNotFound
	if errors.Is(res.Error , gorm.ErrRecordNotFound) {
		application.LOGGER.Errorf("error, no such transaction")
		return nil, errors.New("no transaction exist!")
	}else if res.Error != nil {
		application.LOGGER.Errorf("error writing to db, err : ", res.Error)
		return nil, errors.New("Error writing to db!")
	}

	//Publish Trx for client
	bizCtx := &sdk.BusinessTransactionContext{
		Initiator:       trx.InitiatorSystem,
		TransactionId:   trx.TransactionId,
		BusinessId:      trx.BusinessId,
		BusinessType:    trx.BusinessType,
		TransactionTime: trx.CreatedAt,
		FinishPhaseTime: time.Now(),
		ActionType:      sdk.ACTION_TYPE_ROLLBACK,
	}

	bizCtxJson,_ := json.Marshal(bizCtx)
	topic := sdk.SECONDBASER_PREFIX_TOPIC + trx.BusinessType + "_" + trx.InitiatorSystem
	_ = PublishMessage(ctx, topic, bizCtxJson)

	application.LOGGER.Infof("Commit transaction finished")
	return &api.TransactionResponse{
		TransactionStatus: string(model.TRX_ROLLBACK),
		TransactionStart:  timestamppb.New(trx.CreatedAt),
		TransactionId:     request.TransactionId,
		InitSystem:        trx.InitiatorSystem,
	}, nil
}

func (t transactionServiceImpl) StartTransaction(ctx context.Context, request *api.TransactionRequest) (*api.TransactionResponse, error) {
	setLogFormat(ctx)
	application.LOGGER.Infof("Commit transaction with payload [%+v]", request)

	if request.TransactionId == "" {
		application.LOGGER.Errorf("error illegal param, no trx id")
		return nil, errors.New("Fill transaction ID!")
	}

	if request.InitSystem == "" {
		application.LOGGER.Errorf("error illegal param, no init system")
		return nil, errors.New("Init system cannot be empty!")
	}

	if request.BizId == "" {
		application.LOGGER.Errorf("error illegal param, no biz id")
		return nil, errors.New("Biz id cannot be empty!")
	}

	if request.BizType == "" {
		application.LOGGER.Errorf("error illegal param, no biz type")
		return nil, errors.New("Biz type cannot be empty!")
	}

	//Start transaction
	trx := &model.Transaction{
		TransactionId:     request.TransactionId,
		InitiatorSystem:   request.InitSystem,
		TransactionStatus: model.TRX_INIT,
		BusinessId:        request.BizId,
		BusinessType:      request.BizType,
	}

	res := application.DB.Create(trx)

	// check error ErrRecordNotFound
	 if res.Error != nil {
		application.LOGGER.Errorf("error writing to db, err : ", res.Error)
		return nil, errors.New("Error writing to db!")
	}

	application.LOGGER.Infof("Start transaction")
	return &api.TransactionResponse{
		TransactionStatus: string(model.TRX_INIT),
		TransactionStart:  timestamppb.New(time.Now()),
		TransactionId:     request.TransactionId,
		InitSystem:        request.InitSystem,
	}, nil
}

func (t transactionServiceImpl) QueryTransactionStatus(ctx context.Context, request *api.TransactionRequest) (*api.TransactionResponse, error) {
	if request.TransactionId == "" {
		return nil, errors.New("Fill transaction ID!")
	}

	//Load from db
	trx := &model.Transaction{}
	res := application.DB.Model(trx).Where(&model.Transaction{TransactionId: request.TransactionId}).First(trx)

	// check error ErrRecordNotFound
	if errors.Is(res.Error , gorm.ErrRecordNotFound) {
		return nil, errors.New("no transaction exist!")
	}else if res.Error != nil {
		return nil, res.Error
	}

	return &api.TransactionResponse{
		TransactionStatus: string(trx.TransactionStatus),
		TransactionStart:  timestamppb.New(trx.CreatedAt),
		TransactionId:     request.TransactionId,
		InitSystem:        trx.InitiatorSystem,
	}, nil
}

func TransactionServiceImpl() api.TransactionalRequestServer {
	return &transactionServiceImpl{}
}
