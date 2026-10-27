package cli

// Options holds primitive option values managed by a Command.
type Options struct {
	Bools   map[string]bool
	Strings map[string]string
	Ints    map[string]int
}

// NewOptions creates an initialized Options container.
func NewOptions() *Options {
	return &Options{
		Bools:   make(map[string]bool),
		Strings: make(map[string]string),
		Ints:    make(map[string]int),
	}
}

func (o *Options) ensureMaps() {
	if o.Bools == nil {
		o.Bools = make(map[string]bool)
	}
	if o.Strings == nil {
		o.Strings = make(map[string]string)
	}
	if o.Ints == nil {
		o.Ints = make(map[string]int)
	}
}

// Bool returns the boolean value of an option by name, or false if not present.
func (o *Options) Bool(name string) bool {
	if o == nil || o.Bools == nil {
		return false
	}
	return o.Bools[name]
}

// String returns the string value of an option by name, or empty string if not present.
func (o *Options) String(name string) string {
	if o == nil || o.Strings == nil {
		return ""
	}
	return o.Strings[name]
}

// Int returns the integer value of an option by name, or 0 if not present.
func (o *Options) Int(name string) int {
	if o == nil || o.Ints == nil {
		return 0
	}
	return o.Ints[name]
}

// HasBool reports whether the boolean option has been explicitly recorded.
func (o *Options) HasBool(name string) bool {
	if o == nil || o.Bools == nil {
		return false
	}
	_, ok := o.Bools[name]
	return ok
}

// HasString reports whether the string option has been explicitly recorded.
func (o *Options) HasString(name string) bool {
	if o == nil || o.Strings == nil {
		return false
	}
	_, ok := o.Strings[name]
	return ok
}

// HasInt reports whether the integer option has been explicitly recorded.
func (o *Options) HasInt(name string) bool {
	if o == nil || o.Ints == nil {
		return false
	}
	_, ok := o.Ints[name]
	return ok
}

// SetBool sets the boolean value of an option by name.
func (o *Options) SetBool(name string, val bool) {
	o.ensureMaps()
	o.Bools[name] = val
}

// SetString sets the string value of an option by name.
func (o *Options) SetString(name string, val string) {
	o.ensureMaps()
	o.Strings[name] = val
}

// SetInt sets the integer value of an option by name.
func (o *Options) SetInt(name string, val int) {
	o.ensureMaps()
	o.Ints[name] = val
}

