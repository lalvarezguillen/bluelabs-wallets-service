package main

import (
	"time"
)

type Wallet struct {
	ID        uint      `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Name      string    `json:"name"`
	Balance   uint64    `json:"balance"`
}

const (
	AddBalance       string = "ADD"
	SubstractBalance string = "SUBSTRACT"
)

type BalanceChange struct {
	ID            uint      `json:"id" db:"id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	Amount        uint64    `json:"amount"`
	Operation     string    `json:"operation"`
	BalanceBefore uint64    `json:"balance_before" db:"balance_before"`
	BalanceAfter  uint64    `json:"balance_after" db:"balance_after"`
	Reference     string    `json:"reference"`
	Wallet        *Wallet   `json:"-"`
	WalletID      uint      `json:"wallet_id" db:"wallet_id"`
}
