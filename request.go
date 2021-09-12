package main

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

type CreateWalletRequest struct {
	Name string `json:"name" validate:"required"`
}

func (r *CreateWalletRequest) Bind(c echo.Context, w *Wallet) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := r.Validate(); err != nil {
		return err
	}

	w.Name = r.Name
	return nil
}

func (r *CreateWalletRequest) Validate() *ValidationErrors {
	ve := NewValidationErrors()

	if len(r.Name) < 1 {
		ve.Add("name", "Should not be empty")
	}
	if !ve.HasErrors() {
		return nil
	}
	return &ve
}

type ChangeBalanceRequest struct {
	Amount    uint64 `json:"amount" validate:"gt=0"`
	Operation string `json:"operation" validate:"required"`
}

func (r *ChangeBalanceRequest) Bind(c echo.Context, bc *BalanceChange) error {
	if err := c.Bind(r); err != nil {
		fmt.Println("Bind failed")
		return err
	}
	if err := r.Validate(); err != nil {
		return err
	}

	bc.Amount = r.Amount
	bc.Operation = r.Operation
	return nil
}

func (r *ChangeBalanceRequest) Validate() *ValidationErrors {
	ve := NewValidationErrors()

	if r.Amount < 1 {
		ve.Add("amount", "Should be a positive integer")
	}

	if r.Operation != AddBalance && r.Operation != SubstractBalance {
		ve.Add("operation", fmt.Sprintf("Should be one of: %s, %s", AddBalance, SubstractBalance))
	}

	if !ve.HasErrors() {
		return nil
	}
	return &ve
}

type ValidationErrors struct {
	errors map[string][]string
}

func (ve *ValidationErrors) Add(key, err string) {
	ve.errors[key] = append(ve.errors[key], err)
}

func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.errors) >= 1
}

func (ve *ValidationErrors) Error() string {
	return fmt.Sprintf("%+v", ve.errors)
}

func (ve *ValidationErrors) GetRespError() map[string][]string {
	return ve.errors
}

func NewValidationErrors() ValidationErrors {
	return ValidationErrors{
		errors: make(map[string][]string),
	}
}
