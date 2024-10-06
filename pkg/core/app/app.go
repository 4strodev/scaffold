package app

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"sync"

	"github.com/4strodev/scaffold/pkg/core/adapters"
	"github.com/4strodev/scaffold/pkg/core/components"
	"github.com/4strodev/scaffold/pkg/core/lifecycle"
	wiring "github.com/4strodev/wiring/pkg"
)

func NewApp(container wiring.Container) *App {
	app := &App{
		adapters:   make(map[adapters.Adapter]struct{}),
		components: make(map[components.Component]struct{}),
		container:  container,
	}
	if container.HasType(reflect.TypeFor[*slog.Logger]()) {
		var logger *slog.Logger
		container.Resolve(&logger)
		app.logger = logger
	} else {
		// Set default logger
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		app.logger = logger
		container.Singleton(func() *slog.Logger {
			return logger
		})
	}
	return app
}

// App is an application where the components and adapters are attached
type App struct {
	container  wiring.Container
	// adapters and components are a map because if at some point metadata is necessary,
	// it will be easy to make the refactor
	adapters   map[adapters.Adapter]struct{}
	components map[components.Component]struct{}
	logger     *slog.Logger
}

// Start starts the adapters and execute the lifecycle hooks of the attached components
func (app *App) Start() (chan error, error) {
	errorsChannel := make(chan error, len(app.adapters))
	if len(app.adapters) == 0 {
		return errorsChannel, errors.New("no adapters attached to the app")
	}
	// TODO it will be nice that the adapters has the ability to restart itself and do not affect other services
	// for example if the REST API is down for some reason do not affect for example the grpc server
	waitGroup := sync.WaitGroup{}
	for adapter := range app.adapters {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			err := adapter.Start()
			errorsChannel <- err
		}()
	}

	for controller := range app.components {
		onStartHook, ok := controller.(lifecycle.OnStart)
		if !ok {
			continue
		}

		controller.Init(app.container)

		err := onStartHook.OnStart()
		if err != nil {
			errorsChannel <- err
		}
	}

	waitGroup.Wait()
	return errorsChannel, nil
}

// Stop stops the adapters and shutdowns the application gracefully
func (app *App) Stop() error {
	errs := make([]error, 0)
	for adapter := range app.adapters {
		err := adapter.Stop()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		// join errors
		return errors.Join(errs...)
	}

	return nil
}

// AddAdapter adds an adapter to the application
func (app *App) AddAdapter(adapter adapters.Adapter) error {
	_, exists := app.adapters[adapter]
	if exists {
		return fmt.Errorf("adapter '%s' already exists", reflect.TypeOf(adapter).Elem().String())
	}

	err := adapter.Init(app.container)
	if err != nil {
		panic(err)
	}
	app.logger.Info("adapter {adapter} initialized", "adapter", reflect.TypeOf(adapter).Elem().String())

	app.adapters[adapter] = struct{}{}

	return nil
}

// AddComponent adds a component to the app
func (app *App) AddComponent(component components.Component) error {
	_, exists := app.components[component]
	if exists {
		return fmt.Errorf("component '%s' already exists", reflect.TypeOf(component).Elem().String())
	}

	err := component.Init(app.container)
	if err != nil {
		panic(err)
	}
	app.logger.Info("component {component} initialized\n", "component", reflect.TypeOf(component).Elem().String())

	app.components[component] = struct{}{}

	return nil
}
