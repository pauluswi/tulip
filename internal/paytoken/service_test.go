package paytoken

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/pauluswi/tulip/internal/entity"
	"github.com/pauluswi/tulip/pkg/log"
	"github.com/stretchr/testify/assert"

	uuid "github.com/satori/go.uuid"
)

//var errCRUD = errors.New("error crud")

func Test_service_TokenCycle(t *testing.T) {
	logger, _ := log.NewForTest()
	s := NewService(&mockRepository{}, logger)

	ctx := context.Background()

	// token generation
	paytoken, err := s.Generate(ctx, entity.InputGenerate{CustomerID: "6281100099"})
	assert.Nil(t, err)
	assert.NotEmpty(t, paytoken.Token)
	token := paytoken.Token
	assert.Equal(t, 6, len(token))

	// token validation
	val, err := s.Validate(ctx, entity.InputValidate{Token: token})
	assert.Nil(t, err)
	assert.Equal(t, "6281100099", val.CustomerID)
	assert.Equal(t, false, val.IsExpired)

	//get all tokens
	all, err := s.GetPayTokens(ctx, "6281100099")
	assert.Nil(t, err)
	assert.NotEqual(t, 0, len(all))
}

type mockRepository struct {
	items []entity.PayToken
}

func (m mockRepository) Get(ctx context.Context, id string) (entity.PayToken, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return entity.PayToken{}, sql.ErrNoRows
}

func (m mockRepository) GetTodayPayToken(ctx context.Context, id string) (*entity.PayToken, error) {
	// build valid until
	now := time.Now().UTC()
	nextDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)

	tokenmaxduration := 24 * time.Hour

	// If valid until is in the next day, then force valid until only to the end of the day.
	validUntil := now.Add(tokenmaxduration)
	if nextDay.Year() == validUntil.Year() && nextDay.Month() == validUntil.Month() && nextDay.Day() == validUntil.Day() {
		validUntil = nextDay.Add(-1 * time.Millisecond)
	}

	// build output
	out := &entity.PayToken{
		Token:      "111111",
		CustomerID: "6281100099",
		ValidUntil: validUntil,
	}
	if out.Token != "" {
		return out, nil
	}
	return &entity.PayToken{}, sql.ErrNoRows
}

func (m mockRepository) GetPayTokens(ctx context.Context, customer_id string) ([]entity.PayToken, error) {
	var tok entity.PayToken
	for i := 0; i < 5; i++ {
		tok.ID = uuid.NewV4().String()
		tok.Token = "999999"
		tok.TokenDate = time.Now()
		tok.CustomerID = "6281100099"
		tok.ValidUntil = time.Now()
		tok.CreatedAt = time.Now()
		tok.UpdatedAt = time.Now()

		m.items = append(m.items, tok)
	}
	return m.items, nil
}

func (m mockRepository) Save(ctx context.Context, paytoken entity.PayToken) error {
	// if paytoken.ID == "" {
	// 	return errCRUD
	// }
	// m.items = append(m.items, paytoken)
	return nil
}

func (m mockRepository) Update(ctx context.Context, paytoken entity.PayToken) error {
	// if paytoken.Token == "" {
	// 	return errCRUD
	// }
	// for i, item := range m.items {
	// 	if item.ID == paytoken.ID {
	// 		m.items[i] = paytoken
	// 		break
	// 	}
	// }
	return nil
}
