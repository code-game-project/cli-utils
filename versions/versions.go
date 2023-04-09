package versions

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrInvalidVersion      = errors.New("invalid version")
	ErrNoCompatibleVersion = errors.New("no compatible version")
)

type Version []int

// Parses the version string into the Version type.
func Parse(version string) (Version, error) {
	parts := strings.Split(strings.TrimPrefix(version, "v"), ".")
	if len(parts) == 0 || len(parts) > 3 {
		return Version{}, ErrInvalidVersion
	}

	components := make(Version, len(parts))
	for i, p := range parts {
		c, err := strconv.Atoi(p)
		if err != nil {
			return Version{}, ErrInvalidVersion
		}
		components[i] = c
	}
	return components, nil
}

// MustParse is like Parse but panics on error.
func MustParse(version string) Version {
	v, err := Parse(version)
	if err != nil {
		panic(err)
	}
	return v
}

func (v Version) String() string {
	if v == nil {
		return "nil"
	}
	str := fmt.Sprintf("%d", v[0])
	for i, c := range v {
		if i == 0 {
			continue
		}
		str = fmt.Sprintf("%s.%d", str, c)
	}
	return str
}

// Compare returns -1 if a is larger than b, 1 if b is larger than a and 0 if they are equal.
// x.y is treated as x.y.0
func Compare(a, b Version) int {
	for i := 0; i < 3; i++ {
		if i == len(a) || i == len(b) {
			if len(a) > len(b) {
				return -1
			} else if len(b) > len(a) {
				return 1
			} else {
				return 0
			}
		}
		if a[i] > b[i] {
			return -1
		} else if b[i] > a[i] {
			return 1
		}
	}
	return 0
}

// FindCompatibleInMap returns the next best compatible module version in versionsMap (library/protocol version -> application version).
func FindCompatibleInMap(version Version, versionsMap map[string]string) (Version, error) {
	if version, ok := versionsMap[version.String()]; ok {
		v, err := Parse(version)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", version, err)
		}
		return v, nil
	}

	var found Version
	for vStr := range versionsMap {
		v, err := Parse(vStr)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", vStr, err)
		}

		// The supported version (in versionsMap) must not be larger than the requested version (`version`), because
		// it might generate API calls, which aren't supported.
		if version.IsCompatible(v) {
			// use the largest compatible version
			if Compare(found, v) == 1 {
				found = v
			}
		}
	}
	if found == nil {
		return nil, ErrNoCompatibleVersion
	}

	v, err := Parse(versionsMap[found.String()])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", version, err)
	}
	return v, nil
}

// IsCompatible returns false if one of the following conditions is true:
//   - the major component differs
//   - the major component is 0 AND the minor component differs
//   - the minor component of `other` is greater than the minor component of `v`
func (v Version) IsCompatible(other Version) bool {
	if v[0] != other[0] {
		return false
	}
	if len(v) == 1 || len(other) == 1 {
		return true
	}
	if v[0] == 0 {
		return v[1] == other[1]
	}
	return v[1] >= other[1]
}
