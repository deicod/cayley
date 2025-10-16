package diagnostic

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Severity string

const (
	SeverityError   Severity = "ERROR"
	SeverityWarning Severity = "WARNING"
	SeverityInfo    Severity = "INFO"
)

type Diagnostic struct {
	Severity  Severity `json:"severity"`
	Message   string   `json:"message"`
	Statement string   `json:"statement,omitempty"`
	Line      int      `json:"line,omitempty"`
	Column    int      `json:"column,omitempty"`
	Code      string   `json:"code,omitempty"`
	Detail    string   `json:"detail,omitempty"`
}

type Error struct {
	Summary     string       `json:"error"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

func NewError(summary string, diags ...Diagnostic) *Error {
	return &Error{Summary: summary, Diagnostics: append([]Diagnostic(nil), diags...)}
}

func (e *Error) Error() string {
	if len(e.Diagnostics) == 0 {
		return e.Summary
	}
	var b strings.Builder
	if e.Summary != "" {
		b.WriteString(e.Summary)
	} else {
		b.WriteString("gql: error")
	}
	for _, d := range e.Diagnostics {
		if b.Len() > 0 {
			b.WriteString(": ")
		}
		if d.Statement != "" {
			fmt.Fprintf(&b, "%s at line %d, column %d", d.Message, d.Line, d.Column)
		} else {
			b.WriteString(d.Message)
		}
	}
	return b.String()
}

func (e *Error) MarshalJSON() ([]byte, error) {
	type alias Error
	return json.Marshal((*alias)(e))
}

func (e *Error) Unwrap() error {
	return nil
}

func (e *Error) DiagnosticsList() []Diagnostic {
	if e == nil {
		return nil
	}
	return append([]Diagnostic(nil), e.Diagnostics...)
}

func As(err error) (*Error, bool) {
	var target *Error
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}

func Append(err error, diags ...Diagnostic) error {
	if len(diags) == 0 {
		return err
	}
	if derr, ok := As(err); ok {
		derr.Diagnostics = append(derr.Diagnostics, diags...)
		return derr
	}
	return NewError(err.Error(), diags...)
}

func FormatDiagnostics(diags []Diagnostic) string {
	if len(diags) == 0 {
		return ""
	}
	var b strings.Builder
	for i, d := range diags {
		if i > 0 {
			b.WriteString("\n")
		}
		switch d.Severity {
		case "", SeverityError:
			b.WriteString("ERROR")
		default:
			b.WriteString(string(d.Severity))
		}
		if d.Code != "" {
			fmt.Fprintf(&b, " [%s]", d.Code)
		}
		if d.Line > 0 {
			fmt.Fprintf(&b, " at %d:%d", d.Line, d.Column)
		}
		if d.Statement != "" {
			fmt.Fprintf(&b, " (%s)", d.Statement)
		}
		if d.Message != "" {
			fmt.Fprintf(&b, ": %s", d.Message)
		}
		if d.Detail != "" {
			fmt.Fprintf(&b, " — %s", d.Detail)
		}
	}
	return b.String()
}
