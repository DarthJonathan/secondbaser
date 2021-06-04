package service

import (
	"context"
	"encoding/json"
	"errors"
	api "github.com/DarthJonathan/secondbaser/api/go_gen"
	"github.com/DarthJonathan/secondbaser/config/application"
	"github.com/DarthJonathan/secondbaser/model"
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

func (t transactionServiceImpl) RegisterParticipant(ctx context.Context, request *api.TransactionRequest) (*api.TransactionResponse, error) {
	if request.TransactionId == "" {
		return nil, errors.New("Fill transaction ID!")
	}

	if request.ParticipantSystem == "" {
		return nil, errors.New("Participant system cannot be empty!")
	}

	trx := &model.Transaction{}
	res := application.DB.Model(trx).Where(&model.Transaction{TransactionId: request.TransactionId}).First(trx)

	// check error ErrRecordNotFound
	if errors.Is(res.Error , gorm.ErrRecordNotFound) {
		return nil, errors.New("no transaction exist!")
	}else if res.Error != nil {
		return nil, res.Error
	}


	res = application.DB.Where(&model.TransactionParticipant{TransactionId: request.TransactionId}).Updates(&model.Transaction{TransactionStatus: model.TRX_COMMIT})

	// check error ErrRecordNotFound
	if errors.Is(res.Error , gorm.ErrRecordNotFound) {
		return nil, errors.New("no transaction exist!")
	}else if res.Error != nil {
		return nil, res.Error
	}

	//Start transaction
	res = application.DB.Create(&model.TransactionParticipant{
		TransactionId:     request.TransactionId,
		ParticipantSystem: request.ParticipantSystem,
		ParticipantStatus: model.TRX_INIT,
	})

	if res.Error != nil {
		return nil, errors.New("Error writing to db!")
	}

	return nil, nil
}

func (t transactionServiceImpl) CommitTransaction(ctx context.Context, request *api.TransactionRequest) (*api.TransactionResponse, error) {
	if request.TransactionId == "" {
		return nil, errors.New("Fill transaction ID!")
	}

	trx := &model.Transaction{}
	res := application.DB.Model(trx).Where(&model.Transaction{TransactionId: request.TransactionId}).First(trx)

	// check error ErrRecordNotFound
	if errors.Is(res.Error , gorm.ErrRecordNotFound) {
		return nil, errors.New("no transaction exist!")
	}else if res.Error != nil {
		return nil, res.Error
	}

	res = application.DB.Where(&model.TransactionParticipant{TransactionId: request.TransactionId}).Updates(&model.Transaction{TransactionStatus: model.TRX_COMMIT})

	// check error ErrRecordNotFound
	if errors.Is(res.Error , gorm.ErrRecordNotFound) {
		return nil, errors.New("no transaction exist!")
	}else if res.Error != nil {
		return nil, res.Error
	}

	//Publish Trx for client
	trxInfo, _ := json.Marshal(trx)
	PublishMessage(ctx, COMMIT_TRANSACTION, trxInfo)

	return &api.TransactionResponse{
		TransactionStatus: string(trx.TransactionStatus),
		TransactionStart:  timestamppb.New(trx.CreatedAt),
		TransactionId:     request.TransactionId,
		InitSystem:        trx.InitiatorSystem,
	}, nil
}

func (t transactionServiceImpl) RollbackTransaction(ctx context.Context, request *api.TransactionRequest) (*api.TransactionResponse, error) {
	if request.TransactionId == "" {
		return nil, errors.New("Fill transaction ID!")
	}

	trx := &model.Transaction{}
	res := application.DB.Model(trx).Where(&model.Transaction{TransactionId: request.TransactionId}).First(trx)

	// check error ErrRecordNotFound
	if errors.Is(res.Error , gorm.ErrRecordNotFound) {
		return nil, errors.New("no transaction exist!")
	}else if res.Error != nil {
		return nil, res.Error
	}

	res = application.DB.Where(&model.TransactionParticipant{TransactionId: request.TransactionId}).Updates(&model.Transaction{TransactionStatus: model.TRX_ROLLBACK})

	// check error ErrRecordNotFound
	if errors.Is(res.Error , gorm.ErrRecordNotFound) {
		return nil, errors.New("no transaction exist!")
	}else if res.Error != nil {
		return nil, res.Error
	}

	//Publish Trx for client
	trxInfo, _ := json.Marshal(trx)
	PublishMessage(ctx, COMMIT_TRANSACTION, trxInfo)

	return &api.TransactionResponse{
		TransactionStatus: string(trx.TransactionStatus),
		TransactionStart:  timestamppb.New(trx.CreatedAt),
		TransactionId:     request.TransactionId,
		InitSystem:        trx.InitiatorSystem,
	}, nil
}

func (t transactionServiceImpl) StartTransaction(ctx context.Context, request *api.TransactionRequest) (*api.TransactionResponse, error) {
	if request.TransactionId == "" {
		return nil, errors.New("Fill transaction ID!")
	}

	if request.InitSystem == "" {
		return nil, errors.New("Init system cannot be empty!")
	}

	//Start transaction
	trx := &model.Transaction{
		TransactionId:     request.TransactionId,
		InitiatorSystem:   request.InitSystem,
		TransactionStatus: model.TRX_INIT,
	}
	res := application.DB.Create(trx)

	if res.Error != nil {
		return nil, errors.New("Error writing to db!")
	}

	//Publish Trx for client
	trxInfo, _ := json.Marshal(trx)
	PublishMessage(ctx, INIT_TRANSACTION, trxInfo)

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
