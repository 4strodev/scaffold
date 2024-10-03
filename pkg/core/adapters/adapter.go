package adapters

import "github.com/4strodev/hello_api/pkg/core/components"

type Adapter interface {
	components.Component
	Start() error
	Stop() error
}
