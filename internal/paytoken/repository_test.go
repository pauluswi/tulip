package paytoken

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/pauluswi/tulip/internal/entity"
	"github.com/pauluswi/tulip/internal/test"
	"github.com/pauluswi/tulip/pkg/log"
	"github.com/stretchr/testify/assert"

	uuid "github.com/satori/go.uuid"
)

func TestRepository(t *testing.T) {
	logger, _ := log.NewForTest()
	db := test.DB(t)
	test.ResetTables(t, db, "paytokens")
	repo := NewRepository(db, logger)

	ctx := context.Background()

	// create
	err := repo.Save(ctx, entity.PayToken{
		ID:         uuid.NewV4().String(),
		Token:      "999999",
		TokenDate:  time.Now(),
		CustomerID: "081100099",
		ValidUntil: time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	})
	assert.Nil(t, err)

	// get
	paytoken, err := repo.Get(ctx, "999999")
	assert.Nil(t, err)
	assert.Equal(t, "6281100099", paytoken.CustomerID)
	_, err = repo.Get(ctx, "999990")
	assert.Equal(t, sql.ErrNoRows, err)

	// get today token
	todaytoken, err := repo.GetTodayPayToken(ctx, "999999")
	assert.Nil(t, err)
	assert.Equal(t, "6281100099", todaytoken.CustomerID)
	//assert.Equal(t, sql.ErrNoRows, err)

	// get multi token
	_, err = repo.GetPayTokens(ctx, "6281100099")
	assert.Nil(t, err)
	//assert.Equal(t, sql.ErrNoRows, err)

	// update
	err = repo.Update(ctx, entity.PayToken{
		ID:        paytoken.ID,
		Metadata:  entity.Metadata{time.Now().UTC()},
		UpdatedAt: time.Now(),
	})
	assert.Nil(t, err)

	// get after update
	updatedpaytoken, err := repo.Get(ctx, "999999")
	assert.Nil(t, err)
	assert.Equal(t, false, updatedpaytoken.Metadata.ValidatedAt.IsZero())
}
