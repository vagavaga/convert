package convert

// Interface is simplest conversion api
type Interface interface {
	Convert(in, out interface{}, i Interface) error
}

// Function that implements Interface
type Function func(in, out interface{}, i Interface) error

// Convert implements Interface#Convert
func (f Function) Convert(in, out interface{}, i Interface) error {
	return f(in, out, i)
}
