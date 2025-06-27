package helmutils

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// inspired https://github.com/helm/helm/blob/v3.18.3/pkg/ignore/rules.go

// Rules is a collection of path matching rules.
//
// AddRules() and ParseFile() will construct and populate new Rules.
// RulesWithDefaults() will create an immutable empty ruleset.
type Rules struct {
	patterns []*pattern
	rules    map[string]bool
}

func (r *Rules) getPatterns() []*pattern {
	return r.patterns
}

// RulesWithDefaults builds a ruleset with default patterns.
func RulesWithDefaults() *Rules {
	r := &Rules{patterns: []*pattern{}, rules: make(map[string]bool)}
	_ = r.parseRule(`templates/.?*`)
	return r
}

// Ignore evaluates the file at the given path, and returns true if it should be ignored.
//
// Ignore evaluates path against the rules in order. Evaluation stops when a match is found
func (r *Rules) Ignore(path string, fi os.FileInfo) bool {
	// Don't match on empty dirs.
	if path == "" {
		return false
	}

	// Disallow ignoring the current working directory.
	if path == "." || path == "./" {
		return false
	}
	for _, p := range r.patterns {
		// If the rule is looking for directories, and this is not a directory, skip it.
		if p.mustDir && !fi.IsDir() {
			continue
		}
		if p.match(path, fi) {
			return true
		}
	}
	return false
}

func (r *Rules) AddRules(patterns []string) error {
	for _, rule := range patterns {
		if err := r.parseRule(rule); err != nil {
			return errors.Wrapf(err, "failed to parse rule %q", rule)
		}
	}
	return nil
}

// parseRule parses a rule string and creates a pattern, which is then stored in the Rules object.
func (r *Rules) parseRule(rule string) error {
	rule = strings.TrimSpace(rule)

	if rule == "" {
		// Ignore blank lines
		return nil
	}

	if _, ok := r.rules[rule]; ok {
		// If the rule is already present, skip it.
		return nil
	}

	// Fail any rules that start with a #, as these are comments.
	// Fail any rules that contain **
	if strings.Contains(rule, "**") {
		return errors.New("double-star (**) syntax is not supported")
	}

	// Fail any patterns that can't compile. A non-empty string must be
	// given to Match() to avoid optimization that skips rule evaluation.
	if _, err := filepath.Match(rule, "abc"); err != nil {
		return err
	}

	p := &pattern{raw: rule}

	// Negation is handled at a higher level, so strip the leading ! from the
	// string.
	if strings.HasPrefix(rule, "!") {
		p.negate = true
		rule = rule[1:]
	}

	// Directory verification is handled by a higher level, so the trailing /
	// is removed from the rule. That way, a directory named "foo" matches,
	// even if the supplied string does not contain a literal slash character.
	if strings.HasSuffix(rule, "/") {
		p.mustDir = true
		rule = strings.TrimSuffix(rule, "/")
	}

	if strings.HasPrefix(rule, "/") {
		// Require path matches the root path.
		p.match = func(n string, _ os.FileInfo) bool {
			rule = strings.TrimPrefix(rule, "/")
			ok, err := filepath.Match(rule, n)
			if err != nil {
				log.Printf("Failed to compile %q: %s", rule, err)
				return false
			}
			return ok
		}
	} else if strings.Contains(rule, "/") {
		// require structural match.
		p.match = func(n string, _ os.FileInfo) bool {
			ok, err := filepath.Match(rule, n)
			if err != nil {
				log.Printf("Failed to compile %q: %s", rule, err)
				return false
			}
			return ok
		}
	} else {
		p.match = func(n string, _ os.FileInfo) bool {
			// When there is no slash in the pattern, we evaluate ONLY the
			// filename.
			n = filepath.Base(n)
			ok, err := filepath.Match(rule, n)
			if err != nil {
				log.Printf("Failed to compile %q: %s", rule, err)
				return false
			}
			return ok
		}
	}
	if p.match != nil {
		r.patterns = append(r.patterns, p)
		r.rules[rule] = true
	}
	return nil
}

// matcher is a function capable of computing a match.
//
// It returns true if the rule matches.
type matcher func(name string, fi os.FileInfo) bool

// pattern describes a pattern to be matched in a rule set.
type pattern struct {
	// raw is the unparsed string, with nothing stripped.
	raw string
	// match is the matcher function.
	match matcher
	// negate indicates that the rule's outcome should be negated.
	negate bool
	// mustDir indicates that the matched file must be a directory.
	mustDir bool
}
