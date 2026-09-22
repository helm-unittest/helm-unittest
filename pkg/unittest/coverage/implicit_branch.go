package coverage

import (
	"fmt"
	"text/template/parse"
)

// maybeEmitImplicitBranch inspects an action node looking for an "implicit
// branch" expressed inside its pipeline — currently `default` and `ternary`.
// When one is found AND the input expression is a simple field / variable
// reference (no nested function call that might run with side effects), the
// instrumenter prepends a side-effect-free probe block that emits one of two
// tokens depending on which path the helper would take.
//
// The probe block is emitted BEFORE the original action so the action's own
// value flow is untouched; we only re-evaluate the input expression in a
// boolean context. The "simple input" guard ensures that re-evaluation costs
// nothing more than a map / field lookup.
//
// This is a best-effort enhancement: pipelines that don't fit the safe shape
// (e.g. `{{ upper .X | default "y" }}`, `{{ default "y" (lookup ...) }}`) are
// left alone and only the standard action probe is emitted.
func (in *Instrumenter) maybeEmitImplicitBranch(action *parse.ActionNode, meta *TemplateMeta) {
	if action == nil || action.Pipe == nil {
		return
	}
	pipe := action.Pipe
	if len(pipe.Cmds) == 0 {
		return
	}
	last := pipe.Cmds[len(pipe.Cmds)-1]
	if last == nil || len(last.Args) == 0 {
		return
	}
	ident, ok := last.Args[0].(*parse.IdentifierNode)
	if !ok {
		return
	}

	switch ident.Ident {
	case "default":
		in.emitDefaultProbe(pipe, last, action.Pos, meta)
	case "ternary":
		in.emitTernaryProbe(pipe, last, action.Pos, meta)
	}
}

// emitDefaultProbe handles both call shapes:
//
//	{{ default FALLBACK INPUT }}
//	{{ INPUT | default FALLBACK }}
//
// The instrumented form prepends:
//
//	{{- if empty <INPUT> }}{{ "<token-fallback>" }}{{- else }}{{ "<token-primary>" }}{{- end -}}
//	<original action>
func (in *Instrumenter) emitDefaultProbe(pipe *parse.PipeNode, last *parse.CommandNode, pos parse.Pos, meta *TemplateMeta) {
	input, ok := extractImplicitInput(pipe, last, 2) // default expects (id, fallback, input)
	if !ok {
		return
	}
	primary := in.tracker.registerProbe(Probe{
		Kind: ProbeBranch, TemplateName: in.templateName,
		Line: linecol(in.lineOffsets, pos, 0), Col: linecol(in.lineOffsets, pos, 1),
		Label: "default-primary",
	})
	fallback := in.tracker.registerProbe(Probe{
		Kind: ProbeBranch, TemplateName: in.templateName,
		Line: linecol(in.lineOffsets, pos, 0), Col: linecol(in.lineOffsets, pos, 1),
		Label: "default-fallback",
	})
	meta.ProbeIdxs = append(meta.ProbeIdxs, primary, fallback)

	fmt.Fprintf(&in.out, `{{- if empty %s }}{{ "%s" }}{{- else }}{{ "%s" }}{{- end -}}`,
		input, ProbeToken(fallback), ProbeToken(primary))
}

// emitTernaryProbe handles both call shapes:
//
//	{{ ternary TRUEVAL FALSEVAL COND }}
//	{{ COND | ternary TRUEVAL FALSEVAL }}
func (in *Instrumenter) emitTernaryProbe(pipe *parse.PipeNode, last *parse.CommandNode, pos parse.Pos, meta *TemplateMeta) {
	cond, ok := extractImplicitInput(pipe, last, 3) // ternary expects (id, true, false, cond)
	if !ok {
		return
	}
	truth := in.tracker.registerProbe(Probe{
		Kind: ProbeBranch, TemplateName: in.templateName,
		Line: linecol(in.lineOffsets, pos, 0), Col: linecol(in.lineOffsets, pos, 1),
		Label: "ternary-true",
	})
	falsy := in.tracker.registerProbe(Probe{
		Kind: ProbeBranch, TemplateName: in.templateName,
		Line: linecol(in.lineOffsets, pos, 0), Col: linecol(in.lineOffsets, pos, 1),
		Label: "ternary-false",
	})
	meta.ProbeIdxs = append(meta.ProbeIdxs, truth, falsy)

	fmt.Fprintf(&in.out, `{{- if %s }}{{ "%s" }}{{- else }}{{ "%s" }}{{- end -}}`,
		cond, ProbeToken(truth), ProbeToken(falsy))
}

// extractImplicitInput pulls the "interesting" argument out of a default /
// ternary call. callArgCount is the number of args the function takes in its
// "no-pipe" form (e.g. default takes 2 args, so callArgCount is 2 + 1 for the
// identifier itself when used directly; pass that final number here).
//
// Returns the source-text rendering of the input expression, parenthesised so
// it slots into a surrounding `if` action safely. Returns ok=false when the
// expression is not safe to re-evaluate (anything that's not a plain field /
// variable / chain / dot reference).
func extractImplicitInput(pipe *parse.PipeNode, last *parse.CommandNode, callArgCount int) (string, bool) {
	// Form 1: direct call — `default FALLBACK INPUT` is one command with
	// args [identifier, FALLBACK, INPUT]. The interesting arg is the LAST one.
	if len(last.Args) == callArgCount+1 {
		input := last.Args[callArgCount]
		if !isSimpleAccess(input) {
			return "", false
		}
		return "(" + input.String() + ")", true
	}
	// Form 2: pipe call — `INPUT | default FALLBACK` means `last` is
	// `default FALLBACK` and the interesting input is the result of every
	// command before it. We only handle the case where there's exactly one
	// preceding command containing a single simple-access arg, so the
	// re-evaluation cost is bounded.
	if len(last.Args) == callArgCount && len(pipe.Cmds) == 2 {
		prev := pipe.Cmds[0]
		if len(prev.Args) != 1 {
			return "", false
		}
		if !isSimpleAccess(prev.Args[0]) {
			return "", false
		}
		return "(" + prev.Args[0].String() + ")", true
	}
	return "", false
}

// isSimpleAccess reports whether re-evaluating n in a side probe is safe.
// "Safe" means: no function calls, no template invocations, no nested
// pipelines. Plain field walks (.a.b.c), variables ($x), root context ({{.}})
// and chains rooted in any of those are allowed.
func isSimpleAccess(n parse.Node) bool {
	switch v := n.(type) {
	case *parse.FieldNode, *parse.VariableNode, *parse.DotNode:
		return true
	case *parse.ChainNode:
		return isSimpleAccess(v.Node)
	default:
		return false
	}
}

// linecol returns either the line (which=0) or column (which=1) for pos. It is
// a small adapter around positionLineCol to make probe registration sites less
// noisy.
func linecol(offsets []int, pos parse.Pos, which int) int {
	line, col := positionLineCol(offsets, int(pos))
	if which == 0 {
		return line
	}
	return col
}
