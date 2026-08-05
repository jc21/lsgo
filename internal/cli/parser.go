package cli

import (
	"fmt"
	"strings"
)

// ParseError is returned for any malformed or unrecognised command line.
type ParseError struct {
	Msg string
}

func (e *ParseError) Error() string { return e.Msg }

// ParsedFlags is the result of tokenizing argv: which flags were seen (and
// how many times, which matters for -a/-aa), the last value given to each
// value-taking flag, and the remaining positional (non-flag) arguments in
// their original order.
//
// Flags are last-wins when repeated -- e.g. `--sort=size --sort=name`
// behaves as though only `--sort=name` were given -- which matches
// running with an alias/wrapper script that appends flags after the
// user's own.
type ParsedFlags struct {
	counts map[string]int
	values map[string]string
	order  map[string]int // canonical name -> sequence number of its last occurrence
	seq    int
	Free   []string
}

// Has reports whether a flag was given at all.
func (p *ParsedFlags) Has(long string) bool { return p.counts[long] > 0 }

// Count reports how many times a flag was given.
func (p *ParsedFlags) Count(long string) int { return p.counts[long] }

// Value returns the last value given to a value-taking flag.
func (p *ParsedFlags) Value(long string) (string, bool) {
	v, ok := p.values[long]
	return v, ok
}

// LastOf returns whichever of the given flags was set most recently
// (comparing across different flags, not just repeats of the same one),
// used to resolve combinations like "--oneline --long" where the later
// flag on the command line should win. ok is false if none of them were
// given at all.
func (p *ParsedFlags) LastOf(longs ...string) (winner string, ok bool) {
	best := -1
	for _, l := range longs {
		if seq, present := p.order[l]; present && seq > best {
			best = seq
			winner = l
			ok = true
		}
	}
	return winner, ok
}

func (p *ParsedFlags) mark(canon string) {
	p.seq++
	p.counts[canon]++
	if p.order == nil {
		p.order = map[string]int{}
	}
	p.order[canon] = p.seq
}

// Parse tokenizes argv (not including the program name) into flags and
// free arguments.
func Parse(argv []string) (*ParsedFlags, error) {
	p := &ParsedFlags{counts: map[string]int{}, values: map[string]string{}, order: map[string]int{}}

	i := 0
	freeOnly := false
	for i < len(argv) {
		arg := argv[i]
		i++

		if freeOnly {
			p.Free = append(p.Free, arg)
			continue
		}

		switch {
		case arg == "--":
			freeOnly = true

		case strings.HasPrefix(arg, "--"):
			if err := p.parseLong(arg[2:], argv, &i); err != nil {
				return nil, err
			}

		case len(arg) > 1 && arg[0] == '-':
			if err := p.parseShortCluster(arg[1:], argv, &i); err != nil {
				return nil, err
			}

		default:
			p.Free = append(p.Free, arg)
		}
	}

	return p, nil
}

func (p *ParsedFlags) parseLong(body string, argv []string, i *int) error {
	name, value, hasEq := strings.Cut(body, "=")

	def, ok := findByLong(name)
	if !ok {
		return &ParseError{Msg: fmt.Sprintf("unknown option '--%s'", name)}
	}
	canon := canonicalLong(name)

	if def.kind == noValue {
		if hasEq {
			return &ParseError{Msg: fmt.Sprintf("option '--%s' takes no value", name)}
		}
		p.mark(canon)
		return nil
	}

	if !hasEq {
		if *i >= len(argv) {
			return &ParseError{Msg: fmt.Sprintf("option '--%s' requires a value", name)}
		}
		value = argv[*i]
		*i++
	}

	p.mark(canon)
	p.values[canon] = value
	return nil
}

func (p *ParsedFlags) parseShortCluster(body string, argv []string, i *int) error {
	for idx := 0; idx < len(body); idx++ {
		ch := body[idx]

		def, ok := findByShort(ch)
		if !ok {
			return &ParseError{Msg: fmt.Sprintf("unknown option '-%c'", ch)}
		}

		if def.kind == noValue {
			p.mark(def.long)
			continue
		}

		// A value-taking short flag consumes the rest of this cluster
		// as its value (optionally after a '='), or the next argument
		// if nothing follows it here -- so e.g. "-RL4" sets level to
		// "4", and "-L" "4" does too.
		rest := body[idx+1:]
		rest = strings.TrimPrefix(rest, "=")

		if rest == "" {
			if *i >= len(argv) {
				return &ParseError{Msg: fmt.Sprintf("option '-%c' requires a value", ch)}
			}
			rest = argv[*i]
			*i++
		}

		p.mark(def.long)
		p.values[def.long] = rest
		return nil
	}
	return nil
}
