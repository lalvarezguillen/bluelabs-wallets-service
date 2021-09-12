package main

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

type DummyTx struct {
	RollbackCalls []struct{}
	CommitCalls   []struct{}
}

func (d *DummyTx) Rollback() error {
	d.RollbackCalls = append(d.RollbackCalls, struct{}{})
	return nil
}

func (d *DummyTx) Commit() error {
	d.CommitCalls = append(d.CommitCalls, struct{}{})
	return nil
}

func (d *DummyTx) Get(dest interface{}, stm string, args ...interface{}) error {
	return nil
}

func (d *DummyTx) PrepareNamed(stm string) (*sqlx.NamedStmt, error) {
	return nil, nil
}

type BeginTxResult struct {
	Tx  TxExecutor
	Err error
}

type LockAndGetByIdArgs struct {
	ID uint
	TX TxExecutor
}

type LockAndGetByIDResults struct {
	Wallet *Wallet
	Err    error
}

type UpdateWalletArgs struct {
	Wallet *Wallet
	TX     TxExecutor
}

type CreateBalanceChangeArgs struct {
	BalanceChange *BalanceChange
	TX            TxExecutor
}

type CreateBalanceChangeResult struct {
	Err error
}

type DummyWalletStoreAllSucceeds struct {
	BeginTxCalls                    []struct{}
	BeginTxCallsResults             []BeginTxResult
	LockAndGetByIdCalls             []LockAndGetByIdArgs
	LockAndGetByIdCallsResults      []LockAndGetByIDResults
	UpdateWalletCalls               []UpdateWalletArgs
	CreateBalanceChangeCalls        []CreateBalanceChangeArgs
	CreateBalanceChangeCallsResults []CreateBalanceChangeResult
}

func (s *DummyWalletStoreAllSucceeds) BeginTx() (TxExecutor, error) {
	res := s.BeginTxCallsResults[0]
	s.BeginTxCallsResults = s.BeginTxCallsResults[1:]
	return res.Tx, res.Err
}

func (s *DummyWalletStoreAllSucceeds) LockAndGetByID(wID uint, tx TxExecutor) (*Wallet, error) {
	res := s.LockAndGetByIdCallsResults[0]
	s.LockAndGetByIdCallsResults = s.LockAndGetByIdCallsResults[1:]
	return res.Wallet, res.Err
}

func (s *DummyWalletStoreAllSucceeds) UpdateWallet(w *Wallet, tx TxExecutor) error {
	return nil
}

func (s *DummyWalletStoreAllSucceeds) CreateBalanceChange(c *BalanceChange, tx TxExecutor) error {
	c.ID = 1
	res := s.CreateBalanceChangeCallsResults[0]
	s.CreateBalanceChangeCallsResults = s.CreateBalanceChangeCallsResults[1:]
	return res.Err
}

func (s *DummyWalletStoreAllSucceeds) Create(w *Wallet) error {
	return nil
}

func (s *DummyWalletStoreAllSucceeds) GetByID(id uint) (*Wallet, error) {
	return nil, nil
}

func TestWalletServiceChangeBalance(t *testing.T) {
	t.Run("ADD succeeds", func(t *testing.T) {
		var walletID uint = 1
		bc := BalanceChange{Operation: "ADD", Amount: 200}

		tx := DummyTx{}
		store := DummyWalletStoreAllSucceeds{
			BeginTxCallsResults: []BeginTxResult{{&tx, nil}},
			LockAndGetByIdCallsResults: []LockAndGetByIDResults{
				{&Wallet{Balance: 500, ID: 1}, nil},
			},
			CreateBalanceChangeCallsResults: []CreateBalanceChangeResult{
				{nil},
			},
		}
		service := NewWalletService(&store)
		err := service.ChangeBalance(walletID, &bc)
		assert.NoError(t, err)

		assert.Equal(t, bc.ID, uint(1))
		assert.NotNil(t, bc.Wallet)
		assert.Equal(t, bc.Wallet.Balance, uint64(700))
		assert.Equal(t, bc.BalanceAfter, uint64(700))
		assert.Equal(t, bc.BalanceBefore, uint64(500))

		assert.Equal(t, len(tx.CommitCalls), 1)
		assert.Equal(t, len(tx.RollbackCalls), 0)
	})

	t.Run("SUBSTRACT succeeds", func(t *testing.T) {
		var walletID uint = 1
		bc := BalanceChange{Operation: "SUBSTRACT", Amount: 200}

		tx := DummyTx{}
		store := DummyWalletStoreAllSucceeds{
			BeginTxCallsResults: []BeginTxResult{{&tx, nil}},
			LockAndGetByIdCallsResults: []LockAndGetByIDResults{
				{&Wallet{Balance: 500, ID: 1}, nil},
			},
			CreateBalanceChangeCallsResults: []CreateBalanceChangeResult{
				{nil},
			},
		}
		service := NewWalletService(&store)
		err := service.ChangeBalance(walletID, &bc)
		assert.NoError(t, err)

		assert.Equal(t, bc.ID, uint(1))
		assert.NotNil(t, bc.Wallet)
		assert.Equal(t, bc.Wallet.Balance, uint64(300))
		assert.Equal(t, bc.BalanceAfter, uint64(300))
		assert.Equal(t, bc.BalanceBefore, uint64(500))

		assert.Equal(t, len(tx.CommitCalls), 1)
		assert.Equal(t, len(tx.RollbackCalls), 0)
	})

	t.Run("SUBSTRACT fails: insufficient balance", func(t *testing.T) {
		var walletID uint = 1
		bc := BalanceChange{Operation: "SUBSTRACT", Amount: 200}

		tx := DummyTx{}
		store := DummyWalletStoreAllSucceeds{
			BeginTxCallsResults: []BeginTxResult{{&tx, nil}},
			LockAndGetByIdCallsResults: []LockAndGetByIDResults{
				{&Wallet{Balance: 150, ID: 1}, nil},
			},
			CreateBalanceChangeCallsResults: []CreateBalanceChangeResult{
				{nil},
			},
		}
		service := NewWalletService(&store)
		err := service.ChangeBalance(walletID, &bc)
		assert.Error(t, err)
		var errInsfBal *ErrInsufficientBalance
		assert.True(t, errors.As(err, &errInsfBal))

		assert.Equal(t, len(tx.CommitCalls), 0)
		assert.Equal(t, len(tx.RollbackCalls), 1)
	})

	t.Run("fails: ErrNotFound if Wallet doesn't exist", func(t *testing.T) {
		var walletID uint = 1
		bc := BalanceChange{Operation: "ADD", Amount: 200}

		tx := DummyTx{}
		store := DummyWalletStoreAllSucceeds{
			BeginTxCallsResults: []BeginTxResult{{&tx, nil}},
			LockAndGetByIdCallsResults: []LockAndGetByIDResults{
				{nil, sql.ErrNoRows},
			},
			CreateBalanceChangeCallsResults: []CreateBalanceChangeResult{
				{nil},
			},
		}
		service := NewWalletService(&store)
		err := service.ChangeBalance(walletID, &bc)
		assert.Error(t, err)
		var err404 *ErrNotFound
		assert.True(t, errors.As(err, &err404))

		assert.Equal(t, len(tx.CommitCalls), 0)
		assert.Equal(t, len(tx.RollbackCalls), 1)
	})

	t.Run("fails: unexpected store error", func(t *testing.T) {
		var walletID uint = 1
		bc := BalanceChange{Operation: "ADD", Amount: 200}

		tx := DummyTx{}
		dummyErr := errors.New("Dummy Store Error")
		store := DummyWalletStoreAllSucceeds{
			BeginTxCallsResults: []BeginTxResult{{&tx, nil}},
			LockAndGetByIdCallsResults: []LockAndGetByIDResults{
				{&Wallet{Balance: 150, ID: 1}, nil},
			},
			CreateBalanceChangeCallsResults: []CreateBalanceChangeResult{
				{dummyErr},
			},
		}
		service := NewWalletService(&store)
		err := service.ChangeBalance(walletID, &bc)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, dummyErr))

		assert.Equal(t, len(tx.CommitCalls), 0)
		assert.Equal(t, len(tx.RollbackCalls), 1)
	})
}
