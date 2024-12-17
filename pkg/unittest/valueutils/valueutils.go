package valueutils

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	yaml "gopkg.in/yaml.v3"
)

// GetValueOfSetPath get the value of the `--set` format path from a manifest
func GetValueOfSetPath(manifest common.K8sManifest, path string) ([]interface{}, error) {
	manifestResult := make([]interface{}, 0)
	if path == "" {
		return append(manifestResult, manifest), nil
	}

	byteBuffer := new(bytes.Buffer)

	// Convert K8Manifest to yaml.Node
	var rawManifest yaml.Node
	yamlEncoder := yaml.NewEncoder(byteBuffer)
	yamlEncoder.SetIndent(common.YAMLINDENTION)

	err := yamlEncoder.Encode(manifest)
	if err != nil {
		return nil, err
	}

	yamlDecoder := yaml.NewDecoder(byteBuffer)

	if err := yamlDecoder.Decode(&rawManifest); err != nil {
		return nil, err
	}

	// Set Path
	yamlPath, err := yamlpath.NewPath(path)
	if err != nil {
		return nil, err
	}

	// Search for nodes
	manifestParts, err := yamlPath.Find(&rawManifest)
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

// MergeValues deeply merge values, copied from helm
func MergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	fmt.Println(dest, src)
	for k, v := range src {
		fmt.Println("K:", k, "V:", v)
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}

		if v != nil && reflect.TypeOf(v).Kind() == reflect.Slice && reflect.TypeOf(dest[k]).Kind() == reflect.Slice {
			var result []interface{}
			for i, item := range v.([]interface{}) {
				if item != nil {
					result = append(result, item)
				} else {
					result = append(result, dest[k].([]interface{})[i])
				}
			}
			dest[k] = result
			continue
		}

		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = nextMap
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = MergeValues(destMap, nextMap)
	}
	return dest
}

type parseTraverser interface {
	traverseMapKey(string) error
	traverseListIdx(int) error
}

type buildTraverser struct {
	data    interface{}
	cursors []interface{}
}

func (tr *buildTraverser) traverseMapKey(key string) error {
	tr.cursors = append(tr.cursors, key)
	return nil
}

func (tr *buildTraverser) traverseListIdx(idx int) error {
	tr.cursors = append(tr.cursors, idx)
	return nil
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
				return traverser.traverseMapKey(string(k))
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
	if e := traverser.traverseListIdx(idx); e != nil {
		return -1, e
	}
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
		if e := traverser.traverseMapKey(string(k)); e != nil {
			return -1, e
		}
		return expectKey, nil
	case '[':
		if len(k) == 0 {
			bufferedMapKey = ""
			return expectEscaping, nil
		}
		if e := traverser.traverseMapKey(string(k)); e != nil {
			return -1, e
		}
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
		if e := traverser.traverseMapKey(bufferedMapKey); e != nil {
			return -1, e
		}
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
		case r == '\\':
			next, _, e := in.ReadRune()
			if e != nil {
				return v, next, e
			}
			v = append(v, next)
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
