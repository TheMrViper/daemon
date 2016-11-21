package daemon

type Flag interface {
	IsSet() bool
}

type stringFlag struct {
	ptr   *string
	value string
}

func StringFlag(ptr *string, value string) Flag {
	return &stringFlag{
		ptr:   ptr,
		value: value,
	}
}

func (f *stringFlag) IsSet() bool {
	return *f.ptr == f.value
}

type boolFlag struct {
	ptr   *bool
	value bool
}

func BoolFlag(ptr *bool, value bool) Flag {
	return &boolFlag{
		ptr:   ptr,
		value: value,
	}
}

func (f *boolFlag) IsSet() bool {
	return *f.ptr == f.value
}
