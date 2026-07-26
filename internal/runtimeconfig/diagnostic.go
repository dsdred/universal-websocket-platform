package runtimeconfig

import (
	"sort"
	"strconv"
	"strings"
)

// Severity identifies the blocking impact of a Builder Diagnostic.
type Severity string

const SeverityError Severity = "error"

// Diagnostic is one immutable blocking semantic violation.
type Diagnostic struct {
	severity Severity
	code     string
	location string
	message  string
}

func (d Diagnostic) Severity() Severity { return d.severity }
func (d Diagnostic) Code() string       { return d.code }
func (d Diagnostic) Location() string   { return d.location }
func (d Diagnostic) Message() string    { return d.message }

type diagnosticCollector struct {
	values map[string]Diagnostic
}

func newDiagnosticCollector() *diagnosticCollector {
	return &diagnosticCollector{values: make(map[string]Diagnostic)}
}

func (c *diagnosticCollector) add(code, location, message string) {
	key := code + "\x00" + location
	c.values[key] = Diagnostic{
		severity: SeverityError,
		code:     code,
		location: location,
		message:  message,
	}
}

func (c *diagnosticCollector) diagnostics() []Diagnostic {
	result := make([]Diagnostic, 0, len(c.values))
	for _, diagnostic := range c.values {
		result = append(result, diagnostic)
	}
	sort.Slice(result, func(i, j int) bool {
		comparison := compareLocations(result[i].location, result[j].location)
		if comparison != 0 {
			return comparison < 0
		}
		return result[i].code < result[j].code
	})
	return result
}

type locationSegment struct {
	field   string
	index   int
	isIndex bool
}

func compareLocations(first, second string) int {
	a := parseLocation(first)
	b := parseLocation(second)
	for index := 0; index < len(a) && index < len(b); index++ {
		if a[index].isIndex != b[index].isIndex {
			if a[index].isIndex {
				return 1
			}
			return -1
		}
		if a[index].isIndex {
			if a[index].index < b[index].index {
				return -1
			}
			if a[index].index > b[index].index {
				return 1
			}
			continue
		}
		if a[index].field < b[index].field {
			return -1
		}
		if a[index].field > b[index].field {
			return 1
		}
	}
	switch {
	case len(a) < len(b):
		return -1
	case len(a) > len(b):
		return 1
	default:
		return 0
	}
}

func parseLocation(value string) []locationSegment {
	value = strings.TrimPrefix(value, "$.")
	var result []locationSegment
	for len(value) > 0 {
		fieldEnd := strings.IndexAny(value, ".[")
		if fieldEnd < 0 {
			result = append(result, locationSegment{field: value})
			break
		}
		if fieldEnd > 0 {
			result = append(result, locationSegment{field: value[:fieldEnd]})
			value = value[fieldEnd:]
		}
		if strings.HasPrefix(value, ".") {
			value = value[1:]
			continue
		}
		indexEnd := strings.IndexByte(value, ']')
		number, _ := strconv.Atoi(value[1:indexEnd])
		result = append(result, locationSegment{index: number, isIndex: true})
		value = value[indexEnd+1:]
	}
	return result
}
