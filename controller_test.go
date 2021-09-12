package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type DummyWalletService struct {
	ChangeBalanceCallsResults []error
}

func (s *DummyWalletService) Create(w *Wallet) error {
	w.ID = 1
	return nil
}

func (s *DummyWalletService) GetByID(id uint) (*Wallet, error) {
	return &Wallet{ID: id}, nil
}

func (s *DummyWalletService) ChangeBalance(wID uint, bc *BalanceChange) error {
	w := Wallet{ID: wID}
	bc.Wallet = &w
	bc.WalletID = w.ID
	bc.ID = 1

	err := s.ChangeBalanceCallsResults[0]
	s.ChangeBalanceCallsResults = s.ChangeBalanceCallsResults[1:]
	return err
}

func TestWalletControllerCreate(t *testing.T) {
	t.Run("Succeeds", func(t *testing.T) {
		service := DummyWalletService{}
		ctrl := WalletController{walletService: &service}

		cw := CreateWalletRequest{Name: "new wallet"}
		jsonW, err := json.Marshal(&cw)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/",
			strings.NewReader(string(jsonW)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		e := echo.New()
		resp := httptest.NewRecorder()
		ctx := e.NewContext(req, resp)

		assert.NoError(t, ctrl.CreateWallet(ctx))
		assert.Equal(t, http.StatusCreated, resp.Code)

		var respW Wallet
		assert.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respW))
		assert.Equal(t, respW.Name, cw.Name)
		assert.True(t, respW.ID > 0)
	})

	t.Run("HTTP 400 if Wallet.Name not provided", func(t *testing.T) {
		service := DummyWalletService{}
		ctrl := WalletController{walletService: &service}

		cw := CreateWalletRequest{}
		jsonW, err := json.Marshal(&cw)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/wallets",
			strings.NewReader(string(jsonW)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		e := echo.New()
		resp := httptest.NewRecorder()
		ctx := e.NewContext(req, resp)

		assert.NoError(t, ctrl.CreateWallet(ctx))
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}

func TestWalletControllerChangeBalance(t *testing.T) {
	t.Run("Succeeds", func(t *testing.T) {
		service := DummyWalletService{
			ChangeBalanceCallsResults: []error{nil},
		}
		ctrl := WalletController{walletService: &service}

		cbr := ChangeBalanceRequest{Operation: "ADD", Amount: 200}
		jsonBc, err := json.Marshal(&cbr)
		assert.NoError(t, err)

		walletID := 1
		url := fmt.Sprintf("/wallets/%d/balance-changes", walletID)

		req := httptest.NewRequest(http.MethodPost, url,
			strings.NewReader(string(jsonBc)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		e := echo.New()
		resp := httptest.NewRecorder()
		ctx := e.NewContext(req, resp)

		assert.NoError(t, ctrl.ChangeBalance(ctx))
		assert.Equal(t, http.StatusCreated, resp.Code)
		fmt.Println(resp.Body.String())

		var bc BalanceChange
		assert.NoError(t, json.Unmarshal(resp.Body.Bytes(), &bc))
		assert.Equal(t, bc.Operation, cbr.Operation)
		assert.Equal(t, bc.Amount, cbr.Amount)
		assert.Equal(t, uint(1), bc.ID)
	})

	t.Run("HTTP 400 if Amount or Operation are invalid", func(t *testing.T) {
		service := DummyWalletService{
			ChangeBalanceCallsResults: []error{nil},
		}
		ctrl := WalletController{walletService: &service}

		invalidRequests := []ChangeBalanceRequest{
			{Amount: 200},
			{Operation: "ADD"},
			{Operation: "invalid", Amount: 200},
		}

		for idx, cbr := range invalidRequests {
			t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
				jsonBc, err := json.Marshal(&cbr)
				assert.NoError(t, err)

				walletID := 1
				url := fmt.Sprintf("/wallets/%d/balance-changes", walletID)

				req := httptest.NewRequest(http.MethodPost, url,
					strings.NewReader(string(jsonBc)))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

				e := echo.New()
				resp := httptest.NewRecorder()
				ctx := e.NewContext(req, resp)

				assert.NoError(t, ctrl.ChangeBalance(ctx))
				assert.Equal(t, http.StatusBadRequest, resp.Code)
			})
		}
	})

	t.Run("HTTP 400 if Wallet has insufficient balance", func(t *testing.T) {
		service := DummyWalletService{
			ChangeBalanceCallsResults: []error{&ErrInsufficientBalance{}},
		}
		ctrl := WalletController{walletService: &service}

		cbr := ChangeBalanceRequest{Operation: "ADD", Amount: 200}
		jsonBc, err := json.Marshal(&cbr)
		assert.NoError(t, err)

		walletID := 1
		url := fmt.Sprintf("/wallets/%d/balance-changes", walletID)

		req := httptest.NewRequest(http.MethodPost, url,
			strings.NewReader(string(jsonBc)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		e := echo.New()
		resp := httptest.NewRecorder()
		ctx := e.NewContext(req, resp)

		err = ctrl.ChangeBalance(ctx)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("HTTP 404 if Wallet doesn't exist", func(t *testing.T) {
		service := DummyWalletService{
			ChangeBalanceCallsResults: []error{&ErrNotFound{}},
		}
		ctrl := WalletController{walletService: &service}

		cbr := ChangeBalanceRequest{Operation: "ADD", Amount: 200}
		jsonBc, err := json.Marshal(&cbr)
		assert.NoError(t, err)

		walletID := 1
		url := fmt.Sprintf("/wallets/%d/balance-changes", walletID)

		req := httptest.NewRequest(http.MethodPost, url,
			strings.NewReader(string(jsonBc)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		e := echo.New()
		resp := httptest.NewRecorder()
		ctx := e.NewContext(req, resp)

		err = ctrl.ChangeBalance(ctx)
		assert.Error(t, err)
		var httpErr *echo.HTTPError
		assert.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusNotFound, httpErr.Code)
	})

	t.Run("HTTP 500 if unexpected error", func(t *testing.T) {
		service := DummyWalletService{
			ChangeBalanceCallsResults: []error{errors.New("Unexpected")},
		}
		ctrl := WalletController{walletService: &service}

		cbr := ChangeBalanceRequest{Operation: "ADD", Amount: 200}
		jsonBc, err := json.Marshal(&cbr)
		assert.NoError(t, err)

		walletID := 1
		url := fmt.Sprintf("/wallets/%d/balance-changes", walletID)

		req := httptest.NewRequest(http.MethodPost, url,
			strings.NewReader(string(jsonBc)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		e := echo.New()
		resp := httptest.NewRecorder()
		ctx := e.NewContext(req, resp)

		err = ctrl.ChangeBalance(ctx)
		assert.Error(t, err)
		var httpErr *echo.HTTPError
		assert.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	})
}
