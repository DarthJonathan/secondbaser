package model

import "gorm.io/gorm"

type TransactionParticipant struct {

	gorm.Model

	TransactionId string `json:"transaction_id"`

	ParticipantSystem string `json:"participant_system"`

	ParticipantStatus TransactionStatus `json:"participant_status"`

}
