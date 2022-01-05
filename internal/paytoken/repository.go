package paytoken

import (
	"context"
	"fmt"
	"strings"
	"time"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/pauluswi/tulip/internal/entity"
	"github.com/pauluswi/tulip/pkg/dbcontext"
	"github.com/pauluswi/tulip/pkg/log"
)

// Repository encapsulates the logic to access paytoken from the data source.
type Repository interface {
	// Get returns the customer's token information with the specified token string.
	Get(ctx context.Context, token string) (entity.PayToken, error)
	// GetPayTokens return all payment token belong to a customer.
	GetPayTokens(ctx context.Context, customer_id string) ([]entity.PayToken, error)
	// GetTodayPayToken return a token that still valid and not expire with the specified today date.
	GetTodayPayToken(ctx context.Context, token string) (*entity.PayToken, error)
	// Save will store a token information into data source.
	Save(ctx context.Context, paytoken entity.PayToken) error
	// Update will store an updated token information into data source.
	Update(ctx context.Context, paytoken entity.PayToken) error
}

// repository persists paytoken in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new paytoken repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get returns the customer's token information with the specified token string.
func (r repository) Get(ctx context.Context, token string) (entity.PayToken, error) {
	var paytoken entity.PayToken
	err := r.db.With(ctx).Select("id", "token", "token_date", "customer_id", "valid_until", "metadata", "created_at", "updated_at").
		From("paytokens").
		Where(dbx.HashExp{"token": token}).
		One(&paytoken)
	return paytoken, err
}

// GetPayTokens return all payment token belong to a customer.
func (r repository) GetPayTokens(ctx context.Context, customer_id string) ([]entity.PayToken, error) {
	var paytokens []entity.PayToken
	err := r.db.With(ctx).Select("id", "token", "token_date", "customer_id", "valid_until", "metadata", "created_at", "updated_at").
		From("paytokens").
		Where(dbx.HashExp{"customer_id": customer_id}).
		All(&paytokens)
	return paytokens, err
}

// GetTodayPayToken return a token that still valid and not expire with the specified today date.
func (r repository) GetTodayPayToken(ctx context.Context, tokenString string) (*entity.PayToken, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		err := fmt.Errorf("%w: empty token string", entity.ErrInputValidation)
		return nil, err
	}

	paytoken := &entity.PayToken{}
	today := time.Now().UTC().Format("2006-01-02")
	err := r.db.With(ctx).Select("id", "token", "token_date", "customer_id", "valid_until", "metadata", "created_at", "updated_at").
		From("paytokens").
		Where(dbx.HashExp{"token": tokenString, "token_date": today}).
		One(paytoken)
	return paytoken, err

}

// Save will store a token information into data source.
func (r repository) Save(ctx context.Context, paytoken entity.PayToken) error {
	_, err := r.db.With(ctx).Insert("paytokens", dbx.Params{
		"id":          paytoken.ID,
		"token":       paytoken.Token,
		"token_date":  paytoken.TokenDate,
		"customer_id": paytoken.CustomerID,
		"valid_until": paytoken.ValidUntil,
		"metadata":    "{}",
		"created_at":  paytoken.CreatedAt,
		"updated_at":  paytoken.UpdatedAt,
	}).Execute()
	return err
}

// Update will store an updated token information into data source.
func (r repository) Update(ctx context.Context, paytoken entity.PayToken) error {
	_, err := r.db.With(ctx).Update("paytokens", dbx.Params{
		"metadata":   paytoken.Metadata,
		"updated_at": paytoken.UpdatedAt,
	}, dbx.HashExp{"id": paytoken.ID}).Execute()
	return err
}
