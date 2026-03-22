package engine

import "runtime"

// runtimeGOOS returns the current OS. It exists as a function so tests
// can potentially override it, but by default it returns runtime.GOOS.
func runtimeGOOS() string {
	return runtime.GOOS
}
