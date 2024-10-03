package adapters

import "github.com/4strodev/scaffold/pkg/core/components"

// An adapter is a type that exposes the app via any protocol or service that allows clients to interact with the app.
// The main difference between an adapter and any component with I/O is that an adapter exposes a way to interact
// with the application and if an adapter is down other adapters should keep alive. If a component fails or panics
// all the application is broken.
//
// Any application could have multiple adapters. For example one for a REST API and another for a gRPC server.
// Adapters should be independent from each other and self-contained. That means that they should not depend on
// another adapters or it's side effects.
//
// Adapters do not respond to lifecycle hooks
type Adapter interface {
	components.Component
	Start() error
	Stop() error
}
