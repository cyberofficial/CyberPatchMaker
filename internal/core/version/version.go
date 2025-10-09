package version

// Version information for CyberPatchMaker
// Update these values when releasing a new version
const (
	// Major version number
	Major = 1

	// Minor version number
	Minor = 0

	// Patch version number
	Patch = 12

	// PreRelease suffix (e.g., "alpha", "beta", "rc1")
	// Leave empty for stable releases
	PreRelease = ""
)

// GetVersion returns the full version string
func GetVersion() string {
	v := formatVersion(Major, Minor, Patch)
	if PreRelease != "" {
		v += "-" + PreRelease
	}
	return v
}

// GetShortVersion returns the version without pre-release suffix
func GetShortVersion() string {
	return formatVersion(Major, Minor, Patch)
}

// formatVersion formats major, minor, and patch into a version string
func formatVersion(major, minor, patch int) string {
	return sprintf("%d.%d.%d", major, minor, patch)
}

// sprintf is a simple sprintf implementation for version formatting
func sprintf(format string, major, minor, patch int) string {
	// Convert integers to strings manually
	result := ""
	majorStr := itoa(major)
	minorStr := itoa(minor)
	patchStr := itoa(patch)

	// Replace placeholders
	for i := 0; i < len(format); i++ {
		if format[i] == '%' && i+1 < len(format) && format[i+1] == 'd' {
			if result == "" {
				result += majorStr
			} else if len(result) == len(majorStr)+1 {
				result += minorStr
			} else {
				result += patchStr
			}
			i++ // Skip the 'd'
		} else {
			result += string(format[i])
		}
	}
	return result
}

// itoa converts an integer to a string
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}
