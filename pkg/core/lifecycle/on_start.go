package lifecycle

import "github.com/4strodev/hello_api/pkg/core/components"

// OnStart is a lifecycle hook that allows a controller to execute logic once the controller is started
type OnStart interface {
	components.Component
	OnStart() error
}
