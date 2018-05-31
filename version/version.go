package version

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	// MainVersion ss
	MainVersion, _ = FromString("0.0.0.1")
	// P2PVersion ss
	P2PVersion, _ = FromString("0.0.0.1")

	// KeyStoreVersion ss
	KeyStoreVersion, _ = FromString("0.0.0.1")
)

// Info is a version
type Info struct {
	Value       uint64
	Spec        uint64
	Major       uint64
	Minor       uint64
	Revision    uint64
	StringValue string
}

// FromString ss
func FromString(versionString string) (*Info, error) {
	var info = &Info{
		StringValue: versionString,
	}

	var data = strings.Split(versionString, ".")

	if len(data) != 4 {
		return nil, fmt.Errorf("Invalid version format %v", versionString)
	}

	var tmp, _ = strconv.Atoi(data[0])
	info.Spec = uint64(tmp)
	tmp, _ = strconv.Atoi(data[1])
	info.Major = uint64(tmp)
	tmp, _ = strconv.Atoi(data[2])
	info.Minor = uint64(tmp)
	tmp, _ = strconv.Atoi(data[3])
	info.Revision = uint64(tmp)

	info.Value |= info.Spec << 48
	info.Value |= info.Major << 32
	info.Value |= info.Minor << 16
	info.Value |= info.Revision

	return info, nil
}

// FromUint64 ss
func FromUint64(value uint64) *Info {
	var info = &Info{
		Value: value,
	}
	info.Spec = (value >> 48) & 0xffff
	info.Major = (value >> 32) & 0xffff
	info.Minor = (value >> 16) & 0xffff
	info.Revision = value & 0xffff
	info.StringValue = fmt.Sprintf("%d.%d.%d.%d", info.Spec, info.Major, info.Minor, info.Revision)
	return info
}

// Compatible ss
func (m *Info) Compatible(o *Info) error {

	if m.Spec != o.Spec {
		return fmt.Errorf("Different spec version. Expected %v, got %v", m.Spec, o.Spec)
	}

	if m.Major != o.Major {
		return fmt.Errorf("Different major version. Expected %v, got %v", m.Major, o.Major)
	}

	if m.Minor != o.Minor {
		return fmt.Errorf("Different minor version. Expected %v, got %v", m.Minor, o.Minor)
	}

	return nil
}
