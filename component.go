package gotui

// Component is a Model that can run standalone and return a result.
type Component interface {
	Model
	Result() interface{}
}

// Run runs a component in its own Program and returns its result.
func Run(component Component) (interface{}, error) {
	prog := NewProgram(component)
	if err := prog.Run(); err != nil {
		return nil, err
	}
	return component.Result(), nil
}
