package model

import "gorm.io/gorm"

type Transaction struct {

	gorm.Model

	TransactionId string `gorm:"transaction_id;index:idx_trx_id"`

	InitiatorSystem string `gorm:"initiator_system"`

	TransactionStatus TransactionStatus `gorm:"trx_status"`

	BusinessId string `json:"business_id"`

	BusinessType string `json:"business_type"`

}

type TransactionStatus string

var (
	TRX_INIT 		TransactionStatus = "INIT"
	TRX_COMMIT 		TransactionStatus = "COMMIT"
	TRX_ROLLBACK 	TransactionStatus = "ROLLBACK"
)