package gotui

// Password is a masked text input.
type Password struct {
	*Input
}

func NewPassword() *Password {
	return &Password{Input: NewInput().AsPassword()}
}

func (p *Password) Run() (string, error) {
	return p.Input.Run()
}
