package coverage

import (
	"fmt"
	"strings"
	"text/template/parse"

	"github.com/Masterminds/sprig/v3"
)

// tokenPrefix is the inline marker emitted by every probe. It is intentionally
// unlikely to appear in real chart output and uses only word characters so it
// survives most YAML quoting / escaping.
const tokenPrefix = "__HELMCOV_PROBE_"
const tokenSuffix = "__"

// ProbeToken returns the inline marker text for a given global probe index.
func ProbeToken(idx int) string {
	return fmt.Sprintf("%s%d%s", tokenPrefix, idx, tokenSuffix)
}

// stubFuncs returns a func map that satisfies the text/template parser for
// charts using Sprig and Helm-specific functions. The bodies are placeholders:
// we never execute them. Helm's real engine runs the templates later with its
// own function bindings, so what we provide here only needs to type-check the
// parse tree.
//
// We seed the map from sprig.GenericFuncMap so any Sprig function the user's
// chart references (append, replace, replace, etc.) is recognised, then layer
// on the Helm engine extras that Sprig doesn't ship.
func stubFuncs() map[string]any {
	stub := func(_ ...any) any { return nil }
	stubErr := func(_ ...any) (any, error) { return nil, nil }

	m := map[string]any{}
	for name := range sprig.GenericFuncMap() {
		m[name] = stub
	}
	// Helm engine funcs that aren't in Sprig.
	helmExtras := []string{
		"include", "tpl", "fail",
		"toYaml", "toYamlPretty", "toToml", "toJson",
		"fromYaml", "fromYamlArray", "fromJson", "fromJsonArray",
	}
	for _, n := range helmExtras {
		m[n] = stub
	}
	// Funcs that may be called in error-returning position.
	m["required"] = stubErr
	m["lookup"] = stubErr
	return m
}

// Instrumenter walks a parsed template and rewrites it with probe tokens
// inserted at every tracked construct. Probes are registered with the tracker
// so each instrumented site gets a unique global index.
type Instrumenter struct {
	tracker      *Tracker
	templateName string
	out          strings.Builder
	source       []byte
	lineOffsets  []int // byte offset where each 1-based line starts
}

// Instrument parses data as a Go template and returns the instrumented source
// plus per-template metadata. If the template fails to parse, the returned
// bytes equal the input (no probes injected) and the meta carries ParseError.
func (t *Tracker) Instrument(name string, data []byte) ([]byte, TemplateMeta) {
	meta := TemplateMeta{Name: name, Source: data}

	// SkipFuncCheck makes the parser tolerate any function reference, which
	// is what we want: we never execute these templates ourselves, so the
	// names need not resolve at parse time. Helm's real engine resolves them
	// later, with its own funcMap, when it renders the instrumented chart.
	trees := map[string]*parse.Tree{}
	tree := parse.New(name, stubFuncs())
	tree.Mode = parse.SkipFuncCheck
	if _, err := tree.Parse(string(data), "{{", "}}", trees); err != nil {
		meta.ParseError = err
		return data, meta
	}

	ins := &Instrumenter{
		tracker:      t,
		templateName: name,
		source:       data,
		lineOffsets:  computeLineOffsets(data),
	}

	// The "main" tree carries the file's top-level content. Subtrees come from
	// {{define "x"}}...{{end}} blocks and are keyed by their define name.
	mainName := name
	mainTree, hasMain := trees[mainName]
	if hasMain {
		ins.walkList(mainTree.Root, &meta)
	}

	// Emit each define block. Note that Helm-style _helpers.tpl files often
	// contain ONLY define blocks, so the main tree's Root may be empty / just
	// whitespace — but we still emit the define wrappers here so they remain
	// callable from other templates.
	for tname, tree := range trees {
		if tname == mainName {
			continue
		}
		ins.out.WriteString("\n{{- define \"")
		ins.out.WriteString(tname)
		ins.out.WriteString("\" -}}\n")
		ins.walkList(tree.Root, &meta)
		ins.out.WriteString("\n{{- end -}}\n")
	}

	return []byte(ins.out.String()), meta
}

func (in *Instrumenter) walkList(list *parse.ListNode, meta *TemplateMeta) {
	if list == nil {
		return
	}
	for _, node := range list.Nodes {
		in.walk(node, meta)
	}
}

func (in *Instrumenter) walk(node parse.Node, meta *TemplateMeta) {
	switch n := node.(type) {
	case *parse.ActionNode:
		in.maybeEmitImplicitBranch(n, meta)
		in.out.WriteString(n.String())
		in.emitProbe(ProbeAction, n.Pos, "action", meta)
	case *parse.TemplateNode:
		in.out.WriteString(n.String())
		in.emitProbe(ProbeAction, n.Pos, "template-call", meta)
	case *parse.IfNode:
		in.out.WriteString("{{ if ")
		in.out.WriteString(n.Pipe.String())
		in.out.WriteString(" }}")
		in.emitProbe(ProbeBranch, n.Pos, "if", meta)
		in.walkList(n.List, meta)
		if n.ElseList != nil {
			in.out.WriteString("{{ else }}")
			in.emitProbe(ProbeBranch, n.Pos, "else", meta)
			in.walkList(n.ElseList, meta)
		}
		in.out.WriteString("{{ end }}")
	case *parse.WithNode:
		in.out.WriteString("{{ with ")
		in.out.WriteString(n.Pipe.String())
		in.out.WriteString(" }}")
		in.emitProbe(ProbeBranch, n.Pos, "with", meta)
		in.walkList(n.List, meta)
		if n.ElseList != nil {
			in.out.WriteString("{{ else }}")
			in.emitProbe(ProbeBranch, n.Pos, "with-else", meta)
			in.walkList(n.ElseList, meta)
		}
		in.out.WriteString("{{ end }}")
	case *parse.RangeNode:
		in.out.WriteString("{{ range ")
		in.out.WriteString(n.Pipe.String())
		in.out.WriteString(" }}")
		in.emitProbe(ProbeLoop, n.Pos, "range-body", meta)
		in.walkList(n.List, meta)
		if n.ElseList != nil {
			in.out.WriteString("{{ else }}")
			in.emitProbe(ProbeLoop, n.Pos, "range-else", meta)
			in.walkList(n.ElseList, meta)
		}
		in.out.WriteString("{{ end }}")
	case *parse.ListNode:
		in.walkList(n, meta)
	default:
		// TextNode, CommentNode and anything else we don't instrument:
		// emit verbatim.
		in.out.WriteString(node.String())
	}
}

func (in *Instrumenter) emitProbe(kind ProbeKind, pos parse.Pos, label string, meta *TemplateMeta) {
	line, col := positionLineCol(in.lineOffsets, int(pos))
	idx := in.tracker.registerProbe(Probe{
		Kind:         kind,
		TemplateName: in.templateName,
		Line:         line,
		Col:          col,
		Label:        label,
	})
	meta.ProbeIdxs = append(meta.ProbeIdxs, idx)
	// Wrap the token in a string-literal action so any surrounding template
	// whitespace ("{{-" / "-}}") still applies cleanly; the literal is emitted
	// verbatim into the output.
	in.out.WriteString("{{ \"")
	in.out.WriteString(ProbeToken(idx))
	in.out.WriteString("\" }}")
}

// computeLineOffsets returns the byte offset of the start of each 1-based line.
func computeLineOffsets(data []byte) []int {
	offsets := []int{0}
	for i, b := range data {
		if b == '\n' {
			offsets = append(offsets, i+1)
		}
	}
	return offsets
}

// positionLineCol converts a byte offset to a 1-based (line, col).
func positionLineCol(offsets []int, pos int) (int, int) {
	if len(offsets) == 0 {
		return 1, pos + 1
	}
	lo, hi := 0, len(offsets)-1
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if offsets[mid] <= pos {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return lo + 1, pos - offsets[lo] + 1
}
