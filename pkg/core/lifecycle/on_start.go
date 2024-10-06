package lifecycle

import "github.com/4strodev/scaffold/pkg/core/components"

// OnStart is a lifecycle hook that allows a component to execute logic once the application is started
type OnStart interface {
	components.Component
	OnStart() error
}
