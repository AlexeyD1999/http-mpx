package dependencies

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"mpx/config"
	"mpx/internal/service"
	"mpx/internal/transport"
	"os"
)

type Container struct {
	Logger *zerolog.Logger
	Config *config.Config

	Services *Services
	Handlers *Handlers
}

type Services struct {
	Service *service.Service
}

type Handlers struct {
	Handler *transport.Handler
}

func NewContainer(_ context.Context, cfg *config.Config) (*Container, error) {
	container := new(Container)

	logger := zerolog.New(zerolog.ConsoleWriter{
		Out: os.Stdout,
	}).
		With().
		Timestamp().
		Logger()
	container.Logger = &logger
	container.Config = cfg

	if err := container.initApplicationServices(); err != nil {
		return nil, fmt.Errorf("error initializing application services: %w", err)
	}

	container.initHandlers()

	return container, nil
}

func (c *Container) initApplicationServices() error {
	c.Services = new(Services)

	c.Services.Service = service.NewService()

	return nil
}

func (c *Container) initHandlers() {
	c.Handlers = new(Handlers)

	c.Handlers.Handler = transport.NewHandler(c.Services.Service)
}
