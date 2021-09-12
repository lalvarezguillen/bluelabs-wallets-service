package main

import (
	"database/sql"
	"errors"
)

type WalletService struct {
	store WalletStorer
}

func (s *WalletService) Create(w *Wallet) error {
	return s.store.Create(w)
}

func (s *WalletService) GetByID(id uint) (*Wallet, error) {
	w, err := s.store.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &ErrNotFound{Inner: err}
		}
		return nil, err
	}
	return w, nil
}

func (s *WalletService) ChangeBalance(wID uint, c *BalanceChange) error {
	tx, err := s.store.BeginTx()
	if err != nil {
		return err
	}

	w, err := s.store.LockAndGetByID(wID, tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		if errors.Is(err, sql.ErrNoRows) {
			return &ErrNotFound{Inner: err}
		}
		return err
	}

	c.BalanceBefore = w.Balance

	if c.Operation == SubstractBalance {
		if c.Amount > w.Balance {
			if rbErr := tx.Rollback(); rbErr != nil {
				return rbErr
			}
			return &ErrInsufficientBalance{}
		}
		w.Balance -= c.Amount
	} else {
		w.Balance += c.Amount
	}

	if err := s.store.UpdateWallet(w, tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}

	c.WalletID = w.ID
	c.Wallet = w
	c.BalanceAfter = w.Balance

	if err := s.store.CreateBalanceChange(c, tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func NewWalletService(store WalletStorer) *WalletService {
	return &WalletService{
		store: store,
	}
}

type WalletStorer interface {
	BeginTx() (TxExecutor, error)
	Create(*Wallet) error
	GetByID(uint) (*Wallet, error)
	LockAndGetByID(uint, TxExecutor) (*Wallet, error)
	UpdateWallet(*Wallet, TxExecutor) error
	CreateBalanceChange(*BalanceChange, TxExecutor) error
}

type ErrNotFound struct {
	Inner error
}

func (e *ErrNotFound) Error() string {
	return "Not Found"
}

func (e *ErrNotFound) Unwrap() error {
	return e.Inner
}

type ErrInsufficientBalance struct {
}

func (e *ErrInsufficientBalance) Error() string {
	return "Insufficient Balance"
}
