package app

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

	"github.com/4strodev/scaffold/pkg/core/adapters"
	"github.com/4strodev/scaffold/pkg/core/components"
	"github.com/4strodev/scaffold/pkg/core/lifecycle"
	wiring "github.com/4strodev/wiring/pkg"
)

// Creates a new app with the provided cotnainer. To customize the app logger
// the container has to have a resolver for [*slog.Logger]. If a resolver does not exist then
// a resolver will be added by the app.
func NewApp(container wiring.Container) *App {
	app := &App{
		adapters:   make(map[adapters.Adapter]struct{}),
		components: make(map[components.Component]struct{}),
		container:  container,
	}
	if container.HasType(reflect.TypeFor[*slog.Logger]()) {
		var logger *slog.Logger
		err := container.Resolve(&logger)
		if err != nil {
			panic(err)
		}
		app.logger = logger
	} else {
		// Set default logger
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		app.logger = logger
		err := container.Singleton(func() *slog.Logger {
			return logger
		})
		if err != nil {
			panic(err)
		}
	}
	return app
}

// App is an application where the components and adapters are attached
type App struct {
	container wiring.Container
	// adapters and components are a map because if at some point metadata is necessary,
	// it will be easy to make the refactor
	adapters   map[adapters.Adapter]struct{}
	components map[components.Component]struct{}
	logger     *slog.Logger
}

// Start starts the adapters and execute the lifecycle hooks of the attached components.
// Returns an error if there are no adapters on the app. Errors returned by the adapters
// will be collected an returned in a single error.
func (app *App) Start() error {
	errorsChannel := make(chan error, len(app.adapters))
	if len(app.adapters) == 0 {
		return errors.New("no adapters attached to the app")
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
		err := controller.Init(app.container)
		if err != nil {
			return err
		}

		onStartHook, ok := controller.(lifecycle.OnStart)
		if !ok {
			continue
		}

		err = onStartHook.OnStart()
		if err != nil {
			return err
		}
	}

	var adaptersErrors []error
	go func() {
		for err := range errorsChannel {
			adaptersErrors = append(adaptersErrors, err)
		}
	}()

	app.handleShutdown()
	waitGroup.Wait()
	close(errorsChannel)
	if len(adaptersErrors) != 0 {
		return errors.Join(adaptersErrors...)
	}
	return nil
}

// handleShutdown setups signal notifiers and shutdowns the app if a signal is received from the os
func (a *App) handleShutdown() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigs
		err := a.Stop()
		if err != nil {
			a.logger.Error(err.Error())
		}
	}()
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
