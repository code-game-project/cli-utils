package versions

func (v Version) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *Version) UnmarshalText(text []byte) error {
	if text == nil {
		*v = nil
	}
	var err error
	*v, err = Parse(string(text))
	return err
}
