package sdk

import "time"

type BusinessTransactionContext struct {

	Initiator string `json:"initiator"`

	TransactionId string `json:"transaction_id"`

	BusinessId string `json:"business_id"`

	BusinessType string `json:"business_type"`

	TransactionTime time.Time `json:"transaction_time"`

	FinishPhaseTime time.Time `json:"finish_phase_time"`

	ActionType ActionType `json:"action_type"`

}