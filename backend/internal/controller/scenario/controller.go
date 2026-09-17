package scenario

import "github.com/go-chi/chi/v5"

type Service interface{}

type Controller struct{}

func NewController() *Controller { return &Controller{} }

func (c *Controller) InitRoutes(router chi.Router) {}
