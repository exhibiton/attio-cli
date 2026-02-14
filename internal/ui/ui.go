package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/muesli/termenv"
)

type Options struct {
	Stdout io.Writer
	Stderr io.Writer
	Color  string // auto|always|never
}

const colorNever = "never"

type UI struct {
	out *Printer
	err *Printer
}

type ParseError struct{ msg string }

func (e *ParseError) Error() string { return e.msg }

func New(opts Options) (*UI, error) {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	mode := strings.ToLower(strings.TrimSpace(opts.Color))
	if mode == "" {
		mode = "auto"
	}
	if mode != "auto" && mode != "always" && mode != colorNever {
		return nil, &ParseError{msg: "invalid --color (expected auto|always|never)"}
	}

	out := termenv.NewOutput(opts.Stdout, termenv.WithProfile(termenv.EnvColorProfile()))
	errOut := termenv.NewOutput(opts.Stderr, termenv.WithProfile(termenv.EnvColorProfile()))

	outProfile := chooseProfile(out.Profile, mode)
	errProfile := chooseProfile(errOut.Profile, mode)

	return &UI{
		out: newPrinter(out, outProfile),
		err: newPrinter(errOut, errProfile),
	}, nil
}

func chooseProfile(detected termenv.Profile, mode string) termenv.Profile {
	if termenv.EnvNoColor() {
		return termenv.Ascii
	}
	switch mode {
	case colorNever:
		return termenv.Ascii
	case "always":
		return termenv.TrueColor
	default:
		return detected
	}
}

func (u *UI) Out() *Printer { return u.out }
func (u *UI) Err() *Printer { return u.err }

func (u *UI) OutWriter() io.Writer {
	if u == nil || u.out == nil || u.out.output == nil {
		return os.Stdout
	}
	return u.out.output
}

func (u *UI) ErrWriter() io.Writer {
	if u == nil || u.err == nil || u.err.output == nil {
		return os.Stderr
	}
	return u.err.output
}

type Printer struct {
	output  *termenv.Output
	profile termenv.Profile
}

func newPrinter(output *termenv.Output, profile termenv.Profile) *Printer {
	return &Printer{output: output, profile: profile}
}

func (p *Printer) ColorEnabled() bool { return p.profile != termenv.Ascii }

func (p *Printer) Print(msg string) {
	_, _ = io.WriteString(p.output, msg)
}

func (p *Printer) Println(msg string) {
	_, _ = io.WriteString(p.output, msg+"\n")
}

func (p *Printer) Printf(format string, args ...any) {
	p.Println(fmt.Sprintf(format, args...))
}

func (p *Printer) Successf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if p.ColorEnabled() {
		msg = termenv.String(msg).Foreground(p.profile.Color("#22c55e")).String()
	}
	p.Println(msg)
}

func (p *Printer) Error(msg string) {
	if p.ColorEnabled() {
		msg = termenv.String(msg).Foreground(p.profile.Color("#ef4444")).String()
	}
	p.Println(msg)
}

func (p *Printer) Errorf(format string, args ...any) {
	p.Error(fmt.Sprintf(format, args...))
}

type ctxKey struct{}

func WithUI(ctx context.Context, ui *UI) context.Context {
	return context.WithValue(ctx, ctxKey{}, ui)
}

func FromContext(ctx context.Context) *UI {
	v := ctx.Value(ctxKey{})
	if v == nil {
		return nil
	}
	u, _ := v.(*UI)
	return u
}
