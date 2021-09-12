package main

import (
	"github.com/jmoiron/sqlx"
)

type WalletStore struct {
	db DbExecutor
}

func (s *WalletStore) BeginTx() (TxExecutor, error) {
	return s.db.Beginx()
}

func (s *WalletStore) Create(w *Wallet) error {
	stmt, err := s.db.PrepareNamed("INSERT INTO wallets (name) VALUES (:name) RETURNING id")
	if err != nil {
		return err
	}

	if err := stmt.Get(w, w); err != nil {
		return err
	}

	return nil
}

func (s *WalletStore) GetByID(id uint) (*Wallet, error) {
	var w Wallet
	stm := `SELECT * FROM wallets WHERE id=$1 FOR UPDATE`
	if err := s.db.Get(&w, stm, id); err != nil {
		return nil, err
	}

	return &w, nil
}

func (s *WalletStore) LockAndGetByID(id uint, tx TxExecutor) (*Wallet, error) {
	var w Wallet
	fetchWallet := `SELECT * FROM wallets WHERE id=$1 FOR UPDATE`
	if err := tx.Get(&w, fetchWallet, id); err != nil {
		return nil, err
	}

	return &w, nil
}

func (s *WalletStore) UpdateWallet(w *Wallet, tx TxExecutor) error {
	updateWallet, err := tx.PrepareNamed(`UPDATE wallets SET balance=:balance WHERE id=:id RETURNING id`)
	if err != nil {
		return err
	}
	if err := updateWallet.Get(w, w); err != nil {
		return err
	}
	return nil
}

func (s *WalletStore) CreateBalanceChange(bc *BalanceChange, tx TxExecutor) error {
	insertChange, err := tx.PrepareNamed(`INSERT INTO balance_changes
		(wallet_id, operation, amount, balance_before, balance_after, reference)
		VALUES (:wallet_id,:operation,:amount,:balance_before,:balance_after,:reference)
		RETURNING id`,
	)
	if err != nil {
		return err
	}

	err = insertChange.Get(bc, bc)
	if err != nil {
		return err
	}

	return nil
}

func NewWalletStore(db DbExecutor) *WalletStore {
	return &WalletStore{
		db: db,
	}
}

type TxExecutor interface {
	Get(interface{}, string, ...interface{}) error
	PrepareNamed(string) (*sqlx.NamedStmt, error)
	Commit() error
	Rollback() error
}

type DbExecutor interface {
	Get(interface{}, string, ...interface{}) error
	PrepareNamed(string) (*sqlx.NamedStmt, error)
	Beginx() (*sqlx.Tx, error)
}
