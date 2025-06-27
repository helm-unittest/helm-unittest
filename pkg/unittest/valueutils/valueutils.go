package valueutils

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
)

// GetValueOfSetPath get the value of the `--set` format path from a manifest
func GetValueOfSetPath(manifest common.K8sManifest, path string) ([]interface{}, error) {
	manifestResult := make([]interface{}, 0)
	if path == "" {
		return append(manifestResult, manifest), nil
	}

	byteBuffer := new(bytes.Buffer)

	// Convert K8Manifest to yaml.Node
	node := common.NewYamlNode()
	yamlEncoder := common.YamlNewEncoder(byteBuffer)
	yamlEncoder.SetIndent(common.YAMLINDENTION)

	err := yamlEncoder.Encode(manifest)
	if err != nil {
		return nil, err
	}

	yamlDecoder := common.YamlNewDecoder(byteBuffer)

	if err := yamlDecoder.Decode(&node.Node); err != nil {
		return nil, err
	}

	// Set Path
	yamlPath, err := yamlpath.NewPath(path)
	if err != nil {
		return nil, err
	}

	// Search for nodes
	manifestParts, err := yamlPath.Find(&node.Node)
	if err != nil {
		return nil, err
	}

	for _, node := range manifestParts {
		var singleResult interface{}
		if err := node.Decode(&singleResult); err != nil {
			return nil, err
		}
		manifestResult = append(manifestResult, singleResult)
	}

	return manifestResult, nil
}

// BuildValueOfSetPath build the complete form the `--set` format path and its value
func BuildValueOfSetPath(val interface{}, path string) (map[string]interface{}, error) {
	if path == "" {
		return nil, fmt.Errorf("set path is empty")
	}
	tr := buildTraverser{val, nil}
	reader := bytes.NewBufferString(path)
	if err := traverseSetPath(reader, &tr, expectKey); err != nil {
		return nil, err
	}
	return tr.getBuildedData(), nil
}

// MatchesPattern checks if the input string matches the given regex pattern.
// This method is useful for validating input against a specific format or structure when pattern is only provided at runtime.
// Note: Compiling a regex pattern each time this function is called can be a performance hit.
// To improve performance, consider compiling the regex once and reusing it if the pattern is constant.
func MatchesPattern(input, pattern string) (bool, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, fmt.Errorf("failed to compile regex pattern %s: %v", pattern, err)
	}
	return re.MatchString(input), nil
}

type parseTraverser interface {
	traverseMapKey(string)
	traverseListIdx(int)
}

type buildTraverser struct {
	data    interface{}
	cursors []interface{}
}

func (tr *buildTraverser) traverseMapKey(key string) {
	tr.cursors = append(tr.cursors, key)
}

func (tr *buildTraverser) traverseListIdx(idx int) {
	tr.cursors = append(tr.cursors, idx)
}

func (tr buildTraverser) getBuildedData() map[string]interface{} {
	builded := make(map[string]interface{})
	var current interface{} = builded
	for depth, cursor := range tr.cursors {
		var next interface{}
		if depth == len(tr.cursors)-1 {
			next = tr.data
		} else {
			if idx, ok := tr.cursors[depth+1].(int); ok {
				next = make([]interface{}, idx+1)
			} else {
				next = make(map[string]interface{})
			}
		}
		if key, isString := cursor.(string); isString {
			current.(map[string]interface{})[key] = next
		} else {
			current.([]interface{})[cursor.(int)] = next
		}
		current = next
	}
	return builded
}

const (
	expectKey        = iota
	expectIndex      = iota
	expectDenotation = iota
	expectEscaping   = iota
)

// create a buffer for escapedMapKey
var bufferedMapKey string

func traverseSetPath(in io.RuneReader, traverser parseTraverser, state int) error {
	illegal := runeSet([]rune{',', '{', '}', '='})
	stop := runeSet([]rune{'.', '[', ']', ',', '{', '}', '='})
	k, last, err := runesUntil(in, stop)
	if _, ok := illegal[last]; ok {
		return fmt.Errorf("invalid token found %s", string(last))
	}

	if err != nil {
		if err == io.EOF {
			switch {
			case len(k) != 0 && state == expectKey:
				traverser.traverseMapKey(string(k))
				return nil
			case len(k) == 0 && state == expectDenotation:
				return nil
			default:
				return fmt.Errorf("unexpected end of")
			}
		}
		return err
	}

	var nextState int
	switch state {
	case expectIndex:
		nextState, err = handleExpectIndex(k, last, traverser)
	case expectDenotation:
		nextState, err = handleExpectDenotation(last)
	case expectKey:
		nextState, err = handleExpectKey(k, last, traverser)
	case expectEscaping:
		nextState, err = handleExpectEscaping(k, last, traverser)
	}

	if err != nil {
		return err
	}

	if e := traverseSetPath(in, traverser, nextState); e != nil {
		return e
	}
	return nil
}

func handleExpectIndex(k []rune, last rune, traverser parseTraverser) (int, error) {
	if last != ']' {
		return -1, fmt.Errorf("missing index value")
	}
	idx, idxErr := strconv.Atoi(string(k))
	if idxErr != nil {
		return -1, idxErr
	}
	traverser.traverseListIdx(idx)
	return expectDenotation, nil
}

func handleExpectDenotation(last rune) (int, error) {
	switch last {
	case '.':
		return expectKey, nil
	case '[':
		return expectIndex, nil
	default:
		return -1, fmt.Errorf("invalid denotation token %s", string(last))
	}
}

func handleExpectKey(k []rune, last rune, traverser parseTraverser) (int, error) {
	switch last {
	case '.':
		traverser.traverseMapKey(string(k))
		return expectKey, nil
	case '[':
		if len(k) == 0 {
			bufferedMapKey = ""
			return expectEscaping, nil
		}
		traverser.traverseMapKey(string(k))
		return expectIndex, nil
	default:
		return -1, fmt.Errorf("invalid key %s", string(last))
	}
}

func handleExpectEscaping(k []rune, last rune, traverser parseTraverser) (int, error) {
	switch last {
	case '.':
		bufferedMapKey += string(k) + "."
		return expectEscaping, nil
	case ']':
		bufferedMapKey += string(k)
		traverser.traverseMapKey(bufferedMapKey)
		return expectDenotation, nil
	default:
		return -1, fmt.Errorf("invalid escaping token %s", string(last))
	}
}

// copy from helm
func runesUntil(in io.RuneReader, stop map[rune]bool) ([]rune, rune, error) {
	v := []rune{}
	for {
		switch r, _, e := in.ReadRune(); {
		case e != nil:
			return v, r, e
		case inMap(r, stop):
			return v, r, nil
		default:
			v = append(v, r)
		}
	}
}

// copy from helm
func inMap(k rune, m map[rune]bool) bool {
	_, ok := m[k]
	return ok
}

// copy from helm
func runeSet(r []rune) map[rune]bool {
	s := make(map[rune]bool, len(r))
	for _, rr := range r {
		s[rr] = true
	}
	return s
}
