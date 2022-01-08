package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
)

// PayToken struct is defined here
type PayToken struct {
	ID         string    `db:"id" validate:"required,uuid"`
	Token      string    `db:"token" validate:"required,max=10"`
	TokenDate  time.Time `db:"token_date" validate:"required"`
	CustomerID string    `db:"customer_id" validate:"required"`
	ValidUntil time.Time `db:"valid_until" validate:"required"`
	CreatedAt  time.Time `db:"created_at" validate:"required"`
	UpdatedAt  time.Time `db:"updated_at" validate:"required"`
	Metadata   Metadata  `db:"metadata" validate:"-"`
}

type Metadata struct {
	ValidatedAt time.Time `json:"validated_at"`
}

func NewToken() *PayToken {
	return &PayToken{
		ID: uuid.NewV4().String(),
	}
}

// Value returns m as a value.  This does a validating unmarshal into another
// RawMessage.  If m is invalid json, it returns an error.
func (m Metadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan stores the src in *m.  No validation is done.
func (m *Metadata) Scan(src interface{}) error {
	return json.Unmarshal([]byte(fmt.Sprintf("%s", src)), &m)
}

// --- I/O for Service function

// InputGenerate .
type InputGenerate struct {
	CustomerID string `json:"customer_id" validate:"required,numeric,startswith=62,min=10"`
}

// OutGenerate .
type OutGenerate struct {
	Token      string    `json:"token"`
	ValidUntil time.Time `json:"valid_until"`
}

// InputValidate .
type InputValidate struct {
	Token string `json:"token" validate:"required"`
}

// OutValidate .
type OutValidate struct {
	Token       string    `json:"token"`
	CustomerID  string    `json:"customer_id"`
	ValidUntil  time.Time `json:"valid_until"`
	IsExpired   bool      `json:"is_expired"`
	IsValidated bool      `json:"is_validated"`
}

//PutToken
type InputPutToken struct {
	CustomerID string `json:"customer_id" validate:"required,numeric,startswith=62,min=10"`
	Token      string `json:"token" validate:"required"`
}

// --- list of Postgres Error Code
// --- complete error codes see: https://www.postgresql.org/docs/13/errcodes-appendix.html

const (
	PGErrCodeUniqueViolation = "23505"
)

// --- list of constraint name, for list constraint see migrations/postgres sql schema

const (
	PGConstraintUniqueTokenAndTokenDate = "idx_unq_tokens_token_token_date"
)

var (
	ErrInputValidation       = fmt.Errorf("validation input error")
	ErrDuplicateTokenPerDate = fmt.Errorf("duplicate token generated in current date")
)
