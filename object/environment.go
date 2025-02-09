package object

// Environment is a map of string to Object that represents the environment in which an object is evaluated.
// It also contains a reference to an outer environment, which is used to implement lexical scoping (eg. enables closures)
type Environment struct {
	store map[string]Object
	outer *Environment
}

// NewEnclosedEnvironment creates a new environment with the given outer environment.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// NewEnvironment creates a new environment without an outer environment.
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

// GetStore returns the store of the environment.
func (e *Environment) GetStore() map[string]Object {
	return e.store
}

// Get retrieves the object with the given name from the environment.
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set sets the object with the given name in the environment.
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

// Unset removes the object with the given name from the environment.
func (e *Environment) Unset(name string) {
	delete(e.store, name)
}
