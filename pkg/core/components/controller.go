package components

import wiring "github.com/4strodev/wiring/pkg"

type Component interface {
	Init(container wiring.Container) error
}
