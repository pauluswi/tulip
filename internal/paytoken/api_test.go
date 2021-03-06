package paytoken

import (
	"net/http"
	"testing"
	"time"

	"github.com/pauluswi/tulip/internal/auth"
	"github.com/pauluswi/tulip/internal/entity"
	"github.com/pauluswi/tulip/internal/test"
	"github.com/pauluswi/tulip/pkg/log"
	uuid "github.com/satori/go.uuid"
)

func TestAPI(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)
	repo := &mockRepository{items: []entity.PayToken{
		{uuid.NewV4().String(), "999999", time.Now(), "6281100099", time.Now(), time.Now(), time.Now(), entity.Metadata{time.Now().UTC()}},
	}}
	RegisterHandlers(router.Group(""), NewService(repo, logger), auth.MockAuthHandler, logger)
	header := auth.MockAuthHeader()

	tests := []test.APITestCase{
		{"get all", "GET", "/getpaytokens/6281100099", "", header, http.StatusOK, `*"Token":"999999"`},
		{"get unknown", "GET", "/get/paytokens/62811000991", "", header, http.StatusNotFound, ""},
		{"generate ok", "POST", "/generate", `{"customer_id":"6281100099"}`, header, http.StatusCreated, "*valid_until*"},
		{"generate auth error", "POST", "/generate", `{"customer_id":"6281100099"}`, nil, http.StatusUnauthorized, ""},
		{"generate input error", "POST", "/generate", `"customer_id":"6281100099"}`, header, http.StatusBadRequest, ""},
		{"validate ok", "POST", "/validate", `{"token":"999999"}`, header, http.StatusCreated, "*valid_until*"},
		{"validate auth error", "POST", "/validate", `{"CustomerID":"999999"}`, nil, http.StatusUnauthorized, ""},
		{"validate input error", "POST", "/validate", `"CustomerID":"999999"}`, header, http.StatusBadRequest, ""},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
