package models

import (
	"github.com/google/uuid"
)

type Card struct {
	ID          uuid.UUID `json:"card_id" validate:"required"`
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	AccountID   uuid.UUID `json:"account_id" validate:"required"`
	CardNumber  string    `json:"card_number" validate:"required"`
	PhoneNumber string    `json:"phone_number" validate:"required"`
	Currency    string    `json:"currency" validate:"required"`
	Balance     uint64    `json:"balance" validate:"required"`
}
