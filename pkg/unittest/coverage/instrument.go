package coverage

import (
	"strconv"
	"strings"
	"text/template"
	tpl "text/template"
	"text/template/parse"

	"github.com/Masterminds/sprig/v3"
)

type InsertionStrategy interface {
	insertProbes(template []byte) []byte
}

type TreeStrategy struct {
	sb            strings.Builder
	actionCounter int
	branchCounter int
	rangeCounter  int
}

func (t *TreeStrategy) addActionProbe() {
	t.sb.WriteString("{{ $_ := set $__action \"" + strconv.Itoa(t.actionCounter) + "\" (add1 (get $__action \"" + strconv.Itoa(t.actionCounter) + "\" )) }}")
	t.actionCounter++
}

func (t *TreeStrategy) addBranchProbe() {
	t.sb.WriteString("{{ $_ := set $__branch \"" + strconv.Itoa(t.branchCounter) + "\" (add1 (get $__branch \"" + strconv.Itoa(t.branchCounter) + "\" )) }}")
	t.branchCounter++
}

func (t *TreeStrategy) addRangeProbe() {
	t.sb.WriteString("{{ $_ := set $__loop \"" + strconv.Itoa(t.rangeCounter) + "\" (add1 (get $__loop \"" + strconv.Itoa(t.rangeCounter) + "\" )) }}")
	t.rangeCounter++
}

func (t *TreeStrategy) walkIfNode(node *parse.IfNode) {
	t.sb.WriteString("{{ if " + node.Pipe.String() + " }}")
	t.addBranchProbe()
	if node.List != nil {
		t.walk(node.List)
	}
	if node.ElseList != nil {
		t.sb.WriteString("{{ else }}")
		t.addBranchProbe()
		t.walk(node.ElseList)
	}
	t.sb.WriteString("{{ end }}")
}

func (t *TreeStrategy) walkWithNode(node *parse.WithNode) {
	t.sb.WriteString("{{ with " + node.Pipe.String() + " }}")
	t.addActionProbe()
	t.addBranchProbe()
	if node.List != nil {
		t.walk(node.List)
	}
	t.sb.WriteString("{{ end }}")
}

func (t *TreeStrategy) walkRangeNode(node *parse.RangeNode) {
	t.sb.WriteString("{{ range " + node.Pipe.String() + " }}")
	t.addActionProbe()
	t.addRangeProbe()
	if node.List != nil {
		t.walk(node.List)
	}
	t.sb.WriteString("{{ end }}")
}

func (t *TreeStrategy) walk(node parse.Node) {
	switch node := node.(type) {
	case *parse.IfNode:
		t.walkIfNode(node)
	case *parse.WithNode:
		t.walkWithNode(node)
	case *parse.RangeNode:
		t.walkRangeNode(node)
	case *parse.ListNode:
		for _, node := range node.Nodes {
			t.walk(node)
		}
	case *parse.CommentNode:
		break
	case *parse.ActionNode:
		// println(node.String())
		t.sb.WriteString(node.String())
		t.addActionProbe()
	default:
		t.sb.WriteString(node.String())
	}
}

func (t *TreeStrategy) insertProbes(template []byte) []byte {

	parsedTemplate := tpl.Must(

		tpl.New("deployment").Funcs(funcMap()).Parse(string(template)),
	)

	// log.Println(top)

	for _, node := range parsedTemplate.Root.Nodes {
		t.walk(node)
	}
	// println(t.getActionCounterString())
	var temp strings.Builder
	temp.WriteString("{{- $__action := dict }}\n")
	temp.WriteString("{{- $__branch := dict }}\n")
	temp.WriteString("{{- $__loop := dict }}\n")
	for i := 0; i < t.actionCounter; i++ {
		temp.WriteString("{{ $_ := set $__action \"" + strconv.Itoa(i) + "\" 0 }}\n")
	}
	for i := 0; i < t.branchCounter; i++ {
		temp.WriteString("{{ $_ := set $__branch \"" + strconv.Itoa(i) + "\" 0 }}\n")
	}
	for i := 0; i < t.rangeCounter; i++ {
		temp.WriteString("{{ $_ := set $__loop \"" + strconv.Itoa(i) + "\" 0 }}\n")
	}
	temp.WriteString(t.sb.String())

	if t.actionCounter > 0 {
		temp.WriteString(`
__actions: {{- range $k, $timesExecuted := $__action }}
  - {{ $timesExecuted }}
{{- end }}
`)
	}

	if t.branchCounter > 0 {
		temp.WriteString(`
__branches: {{- range $k, $timesExecuted := $__branch }}
  - {{ $timesExecuted }}
{{- end }}
`)
	}

	if t.rangeCounter > 0 {
		temp.WriteString(`
__loops: {{- range $k, $timesExecuted := $__loop }}
  - {{ $timesExecuted }}
{{- end }}
`)
	}

	return []byte(temp.String())
}

// Instrumenter takes a template and returns a new template with the probes injected
type Instrumenter struct {
	strategy InsertionStrategy
	template []byte
}

// Create a new transformer
func NewInstrumenter(strategy InsertionStrategy, template []byte) *Instrumenter {
	return &Instrumenter{strategy: strategy, template: template}
}

func (t *Instrumenter) Transform() ([]byte, error) {
	newBytes := t.strategy.insertProbes(t.template)

	return newBytes, nil
}

func funcMap() template.FuncMap {
	f := sprig.TxtFuncMap()
	delete(f, "env")
	delete(f, "expandenv")

	// Add some extra functionality
	extra := template.FuncMap{
		"toToml":        func(string, interface{}) string { return "not implemented" },
		"toYaml":        func(string, interface{}) string { return "not implemented" },
		"fromYaml":      func(string, interface{}) string { return "not implemented" },
		"fromYamlArray": func(string, interface{}) string { return "not implemented" },
		"toJson":        func(string, interface{}) string { return "not implemented" },
		"fromJson":      func(string, interface{}) string { return "not implemented" },
		"fromJsonArray": func(string, interface{}) string { return "not implemented" },

		// This is a placeholder for the "include" function, which is
		// late-bound to a template. By declaring it here, we preserve the
		// integrity of the linter.
		"include":  func(string, interface{}) string { return "not implemented" },
		"tpl":      func(string, interface{}) interface{} { return "not implemented" },
		"required": func(string, interface{}) (interface{}, error) { return "not implemented", nil },
		// Provide a placeholder for the "lookup" function, which requires a kubernetes
		// connection.
		"lookup": func(string, string, string, string) (map[string]interface{}, error) {
			return map[string]interface{}{}, nil
		},
	}

	for k, v := range extra {
		f[k] = v
	}

	return f
}
