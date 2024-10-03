package lifecycle

import "github.com/4strodev/scaffold/pkg/core/components"

// OnStart is a lifecycle hook that allows a controller to execute logic once the component is started
type OnStart interface {
	components.Component
	OnStart() error
}
