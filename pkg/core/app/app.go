package app

import (
	"fmt"
	"log/slog"
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
	if !container.HasType(reflect.TypeFor[slog.Logger]()) {
		// Set default logger
		container.Singleton(func() *slog.Logger {
			logger := slog.New(slog.Default().Handler())
			return logger
		})
	}
	return app
}

// App is an application where the components and adapters are attached
type App struct {
	container  wiring.Container
	adapters   map[adapters.Adapter]struct{}
	components map[components.Component]struct{}
	logger     *slog.Logger
}

func (app *App) Start() error {
	// TODO it will be nice that the adapters has the ability to restart itself and do not affect other services
	// for example if the REST API is down for some reason do not affect for example the grpc server
	waitGroup := sync.WaitGroup{}
	errChannel := make(chan error, len(app.adapters))
	for adapter := range app.adapters {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			err := adapter.Start()
			errChannel <- err
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
			return err
		}
	}

	go func() {
		for err := range errChannel {
			app.logger.Error("error on adapter: {error}", "error", err.Error())
		}
	}()

	waitGroup.Wait()
	return nil
}

func (app *App) AddAdapter(adapter adapters.Adapter) error {
	_, exists := app.adapters[adapter]
	if exists {
		return fmt.Errorf("adapter '%s' already exists", reflect.TypeOf(adapter).Elem().String())
	}

	err := adapter.Init(app.container)
	if err != nil {
		panic(err)
	}
	app.logger.Error("adapter {adapter} initialized", "adapter", reflect.TypeOf(adapter).Elem().String())

	app.adapters[adapter] = struct{}{}

	return nil
}

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
