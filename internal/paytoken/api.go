package paytoken

import (
	"net/http"

	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/pauluswi/tulip/internal/entity"
	"github.com/pauluswi/tulip/internal/errors"
	"github.com/pauluswi/tulip/pkg/log"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *routing.RouteGroup, service Service, authHandler routing.Handler, logger log.Logger) {
	res := resource{service, logger}

	// endpoint get
	r.Get("/getpaytokens/<id>", res.getpaytokens)

	r.Use(authHandler)

	// the following endpoints require a valid JWT
	r.Post("/generate", res.generate)
	r.Post("/validate", res.validate)
}

type resource struct {
	service Service
	logger  log.Logger
}

func (r resource) getpaytokens(c *routing.Context) error {
	paytoken, err := r.service.GetPayTokens(c.Request.Context(), c.Param("id"))
	if err != nil {
		return err
	}

	return c.Write(paytoken)
}

func (r resource) generate(c *routing.Context) error {
	var input entity.InputGenerate
	if err := c.Read(&input); err != nil {
		r.logger.With(c.Request.Context()).Info(err)
		return errors.BadRequest("Bad Request")
	}
	paytoken, err := r.service.Generate(c.Request.Context(), input)
	if err != nil {
		return err
	}

	return c.WriteWithStatus(paytoken, http.StatusCreated)
}

func (r resource) validate(c *routing.Context) error {
	var input entity.InputValidate
	if err := c.Read(&input); err != nil {
		r.logger.With(c.Request.Context()).Info(err)
		return errors.BadRequest("Bad Request")
	}
	paytoken, err := r.service.Validate(c.Request.Context(), input)
	if err != nil {
		return err
	}
	return c.WriteWithStatus(paytoken, http.StatusCreated)
}
