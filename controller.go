package main

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type WalletServiceProvider interface {
	Create(*Wallet) error
	GetByID(uint) (*Wallet, error)
	ChangeBalance(uint, *BalanceChange) error
}

type WalletController struct {
	walletService WalletServiceProvider
}

func (h *WalletController) CreateWallet(c echo.Context) error {
	var req CreateWalletRequest
	var w Wallet
	if err := req.Bind(c, &w); err != nil {
		var valErrs *ValidationErrors
		if errors.As(err, &valErrs) {
			return c.JSON(http.StatusBadRequest, valErrs.GetRespError())
		}
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := h.walletService.Create(&w); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, w)
}

func (h *WalletController) GetWalletById(c echo.Context) error {
	var id uint
	echo.PathParamsBinder(c).Uint("id", &id)

	w, err := h.walletService.GetByID(id)
	if err != nil {
		var err404 *ErrNotFound
		if errors.As(err, &err404) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, w)
}

func (h *WalletController) ChangeBalance(c echo.Context) error {
	var id uint
	echo.PathParamsBinder(c).Uint("id", &id)

	var req ChangeBalanceRequest
	var bc BalanceChange
	if err := req.Bind(c, &bc); err != nil {
		var valErrs *ValidationErrors
		if errors.As(err, &valErrs) {
			return c.JSON(http.StatusBadRequest, valErrs.GetRespError())
		}
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := h.walletService.ChangeBalance(id, &bc); err != nil {
		var err404 *ErrNotFound
		if errors.As(err, &err404) {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		var errInsBal *ErrInsufficientBalance
		if errors.As(err, &errInsBal) {
			valErr := NewValidationErrors()
			valErr.Add("amount", "Insufficient balance to cover the deducted amount")
			return c.JSON(http.StatusBadRequest, valErr.GetRespError())
		}
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, bc)
}

func (h *WalletController) Register(r *echo.Group) {
	r.POST("", h.CreateWallet)
	r.GET("/:id", h.GetWalletById)
	r.POST("/:id/balance-changes", h.ChangeBalance)
}

func NewWalletController(ws WalletServiceProvider) *WalletController {
	return &WalletController{
		walletService: ws,
	}
}
