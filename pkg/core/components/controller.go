package components

import wiring "github.com/4strodev/wiring/pkg"

// A component is the base type for the scaffold. In fact, everything is a component
// a component is a type that allows it to inject a container to initialize their dependencies.
// Components could depend on another components and if a component is broken or fails their initalization,
// all the application fails
type Component interface {
	Init(container wiring.Container) error
}
