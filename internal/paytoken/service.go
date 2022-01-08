package paytoken

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/pauluswi/tulip/internal/entity"
	generator "github.com/pauluswi/tulip/pkg/generator"
	"github.com/pauluswi/tulip/pkg/log"
	"github.com/pauluswi/tulip/pkg/validator"
)

// Service encapsulates usecase logic for albums.
type Service interface {
	GetPayTokens(ctx context.Context, customer_id string) ([]entity.PayToken, error)
	Generate(ctx context.Context, req entity.InputGenerate) (out entity.OutGenerate, err error)
	Validate(ctx context.Context, req entity.InputValidate) (out entity.OutValidate, err error)
}

// PayToken represents the data about an payment token.
type PayToken struct {
	entity.PayToken
}

type service struct {
	repo   Repository
	logger log.Logger
}

// NewService creates a new payment token service.
func NewService(repo Repository, logger log.Logger) Service {
	return service{repo, logger}
}

// --- list of error and constants
var (
	ErrValidation    = fmt.Errorf("validation error")
	ErrGenerateToken = fmt.Errorf("token generate error")
	ErrDBPersist     = fmt.Errorf("persist to database error")
	ErrTokenNotFound = fmt.Errorf("token not found in database")
)

// GetPayTokens returns all payment tokens belong to a customer
func (s service) GetPayTokens(ctx context.Context, id string) (out []entity.PayToken, err error) {
	paytokens, err := s.repo.GetPayTokens(ctx, id)
	if err != nil {
		return paytokens, err
	}
	return paytokens, nil
}

// Generate creates a payment token
func (s service) Generate(ctx context.Context, req entity.InputGenerate) (out entity.OutGenerate, err error) {
	defer func() {
		if err != nil {
			s.logger.Error(ctx, err.Error())
		}
	}()

	for i := 0; i < 5; i++ {
		err = validator.ValidateWithOpts(req, validator.Opts{Mode: validator.ModeVerbose})
		if err != nil {
			err = fmt.Errorf("%w: %s", ErrValidation, err)
			return entity.OutGenerate{}, err
		}

		// ** I use a very simple algorithm to generate payment token
		// ** in real world the algorithm must be more details and secure
		token, err := generator.EncodeToString(6)
		if err != nil {
			err = fmt.Errorf("%w: %s", ErrValidation, err)
			return entity.OutGenerate{}, err
		}

		now := time.Now().UTC()
		nextDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)

		tokenmaxduration := 24 * time.Hour

		// If valid until is in the next day, then force valid until only to the end of the day.
		validUntil := now.Add(tokenmaxduration)
		if nextDay.Year() == validUntil.Year() && nextDay.Month() == validUntil.Month() && nextDay.Day() == validUntil.Day() {
			validUntil = nextDay.Add(-1 * time.Millisecond)
		}

		// build new token
		paytoken := entity.NewToken()
		paytoken.Token = token
		paytoken.TokenDate = now
		paytoken.CustomerID = req.CustomerID
		paytoken.ValidUntil = validUntil
		paytoken.CreatedAt = now
		paytoken.UpdatedAt = now

		err = validator.ValidateWithOpts(paytoken, validator.Opts{Mode: validator.ModeVerbose})
		if err != nil {
			err = fmt.Errorf("%w: %s", ErrValidation, err)
			return entity.OutGenerate{}, err
		}

		err = s.repo.Save(ctx, *paytoken)
		if err != nil {
			if errors.Is(err, entity.ErrDuplicateTokenPerDate) {
				// retry if duplicate up to maximum retry
				continue
			}

			err = fmt.Errorf("%w: %s", ErrDBPersist, err)
			return entity.OutGenerate{}, err
		}

		output := entity.OutGenerate{
			Token:      token,
			ValidUntil: validUntil,
		}
		return output, err
	}

	return entity.OutGenerate{}, err
}

// Validate to check whether a token stil valid and not expired.
func (s service) Validate(ctx context.Context, req entity.InputValidate) (out entity.OutValidate, err error) {
	defer func() {
		if err != nil {
			s.logger.Error(ctx, err.Error())
		}
	}()

	err = validator.ValidateWithOpts(req, validator.Opts{Mode: validator.ModeCompact})
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrValidation, err)
		return
	}

	// Find today token only,
	// separate select and update query since select assumed faster than update with no matching records
	inputToken, err := s.repo.GetTodayPayToken(ctx, req.Token)
	if errors.Is(err, sql.ErrNoRows) {
		// if return sql.Row error then don't leak sql error to response
		err = fmt.Errorf("%w: no token found", ErrTokenNotFound)
		return
	}

	if err != nil {
		err = fmt.Errorf("token search failed: %w", err)
		return
	}

	if inputToken == nil {
		err = fmt.Errorf("%w: token empty result", ErrTokenNotFound)
		return
	}

	now := time.Now().UTC()

	// build output
	out = entity.OutValidate{
		Token:       inputToken.Token,
		CustomerID:  inputToken.CustomerID,
		ValidUntil:  inputToken.ValidUntil.UTC(),
		IsExpired:   now.After(inputToken.ValidUntil),
		IsValidated: now.After(inputToken.Metadata.ValidatedAt) && !inputToken.Metadata.ValidatedAt.IsZero(), // first call will return false because validatedAt is zero
	}

	// update only when token hasn't validated
	shouldUpdate := false
	if inputToken.Metadata.ValidatedAt.IsZero() {
		inputToken.Metadata.ValidatedAt = now
		shouldUpdate = true
	}

	if shouldUpdate {
		inputToken.UpdatedAt = now
		err = s.repo.Update(ctx, *inputToken)
	}

	if err != nil {
		err = fmt.Errorf("%w: %s", ErrDBPersist, err)
		return entity.OutValidate{}, err
	}

	return out, err
}
