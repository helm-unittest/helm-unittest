package coverage

import (
	"fmt"
	"strings"
	"text/template/parse"

	"github.com/Masterminds/sprig/v3"
)

// probeFuncName must emit no output so helpers whose result is consumed as data (include ... | fromJson) stay valid after instrumentation.
const probeFuncName = "covprobe"

// stubFuncs provides placeholder funcs (seeded from Sprig plus Helm's extras) so the
// parser accepts any function reference; Helm binds and executes the real ones at render time.
func stubFuncs() map[string]any {
	stub := func(_ ...any) any { return nil }
	stubErr := func(_ ...any) (any, error) { return nil, nil }

	m := map[string]any{}
	for name := range sprig.GenericFuncMap() {
		m[name] = stub
	}
	helmExtras := []string{
		"include", "tpl", "fail",
		"toYaml", "toYamlPretty", "toToml", "toJson",
		"fromYaml", "fromYamlArray", "fromJson", "fromJsonArray",
	}
	for _, n := range helmExtras {
		m[n] = stub
	}
	m["required"] = stubErr
	m["lookup"] = stubErr
	return m
}

// Instrumenter rewrites a parsed template with a covprobe call at each tracked construct.
type Instrumenter struct {
	tracker      *Tracker
	templateName string
	out          strings.Builder
	source       []byte
	lineOffsets  []int // byte offset where each 1-based line starts
	lastByte     byte
}

// emit writes s, inserting a space when a reconstructed action ("{{") would collide with a preceding literal "{" and lex as "{{{" (the original {{- trim is not preserved by the AST).
func (in *Instrumenter) emit(s string) {
	if s == "" {
		return
	}
	if in.lastByte == '{' && len(s) >= 2 && s[0] == '{' && s[1] == '{' {
		in.out.WriteByte(' ')
	}
	in.out.WriteString(s)
	in.lastByte = s[len(s)-1]
}

// Instrument returns the instrumented source; on parse failure it returns the input unchanged with meta.ParseError set.
func (t *Tracker) Instrument(name string, data []byte) ([]byte, TemplateMeta) {
	meta := TemplateMeta{Name: name, Source: data}

	// SkipFuncCheck: we never execute these templates, so unknown funcs need not resolve; Helm resolves them at render time.
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
		ins.emit("\n{{- define \"")
		ins.emit(tname)
		ins.emit("\" -}}\n")
		ins.walkList(tree.Root, &meta)
		ins.emit("\n{{- end -}}\n")
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
		in.emit(n.String())
		in.emitProbe(ProbeAction, n.Pos, "action", meta)
	case *parse.TemplateNode:
		in.emit(n.String())
		in.emitProbe(ProbeAction, n.Pos, "template-call", meta)
	case *parse.IfNode:
		in.emit("{{ if ")
		in.emit(n.Pipe.String())
		in.emit(" }}")
		in.emitProbe(ProbeBranch, n.Pos, "if", meta)
		in.walkList(n.List, meta)
		if n.ElseList != nil {
			in.emit("{{ else }}")
			in.emitProbe(ProbeBranch, n.Pos, "else", meta)
			in.walkList(n.ElseList, meta)
		}
		in.emit("{{ end }}")
	case *parse.WithNode:
		in.emit("{{ with ")
		in.emit(n.Pipe.String())
		in.emit(" }}")
		in.emitProbe(ProbeBranch, n.Pos, "with", meta)
		in.walkList(n.List, meta)
		if n.ElseList != nil {
			in.emit("{{ else }}")
			in.emitProbe(ProbeBranch, n.Pos, "with-else", meta)
			in.walkList(n.ElseList, meta)
		}
		in.emit("{{ end }}")
	case *parse.RangeNode:
		in.emit("{{ range ")
		in.emit(n.Pipe.String())
		in.emit(" }}")
		in.emitProbe(ProbeLoop, n.Pos, "range-body", meta)
		in.walkList(n.List, meta)
		if n.ElseList != nil {
			in.emit("{{ else }}")
			in.emitProbe(ProbeLoop, n.Pos, "range-else", meta)
			in.walkList(n.ElseList, meta)
		}
		in.emit("{{ end }}")
	case *parse.ListNode:
		in.walkList(n, meta)
	default:
		in.emit(node.String())
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
	in.emit(fmt.Sprintf("{{ %s %d }}", probeFuncName, idx))
}

func computeLineOffsets(data []byte) []int {
	offsets := []int{0}
	for i, b := range data {
		if b == '\n' {
			offsets = append(offsets, i+1)
		}
	}
	return offsets
}

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
