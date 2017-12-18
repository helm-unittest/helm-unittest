package helmtest

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	yaml "gopkg.in/yaml.v2"
)

func GetValueOfSetPath(m map[interface{}]interface{}, path string) (interface{}, error) {
	tr := fetchTraverser{m}
	reader := bytes.NewBufferString(path)
	if e := traverseSetPath(reader, &tr, expectKey); e != nil {
		return nil, e
	}
	return tr.data, nil
}

func BuildValueOfSetPath(val interface{}, path string) (interface{}, error) {
	tr := buildTraverser{val, nil}
	reader := bytes.NewBufferString(path)
	if e := traverseSetPath(reader, &tr, expectKey); e != nil {
		return nil, e
	}
	return tr.getBuildedData(), nil
}

type parseTraverser interface {
	traverseMapKey(string) error
	traverseListIdx(int) error
}

type fetchTraverser struct {
	data interface{}
}

func (tr *fetchTraverser) traverseMapKey(key string) error {
	if d, ok := tr.data.(map[interface{}]interface{}); ok {
		tr.data = d[key]
		return nil
	}
	return fmt.Errorf("can't get [\"%s\"] from a non map type:\n%s", key, trustedMarshalYAML(tr.data))
}

func (tr *fetchTraverser) traverseListIdx(idx int) error {
	if d, ok := tr.data.([]interface{}); ok {
		if idx < 0 || idx >= len(d) {
			return fmt.Errorf("[%d] :\n%s", idx, trustedMarshalYAML(d))
		}
		tr.data = d[idx]
		return nil
	}
	return fmt.Errorf("can't get [%d] from a non array type:\n%s", idx, trustedMarshalYAML(tr.data))
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

func (tr buildTraverser) getBuildedData() interface{} {
	builded := make(map[interface{}]interface{})
	var current interface{} = builded
	for depth, cursor := range tr.cursors {
		var next interface{}
		if depth == len(tr.cursors)-1 {
			next = tr.data
		} else {
			if idx, ok := tr.cursors[depth+1].(int); ok {
				next = make([]interface{}, idx+1)
			} else {
				next = make(map[interface{}]interface{})
			}
		}
		if key, isString := cursor.(string); isString {
			current.(map[interface{}]interface{})[key] = next
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
)

func traverseSetPath(in io.RuneReader, traverser parseTraverser, state int) error {
	illegal := runeSet([]rune{',', '{', '}', '='})
	stop := runeSet([]rune{'.', '[', ']', ',', '{', '}', '='})
	k, last, err := runesUntil(in, stop)
	if _, ok := illegal[last]; ok {
		return fmt.Errorf("")
	}

	if err != nil {
		if err == io.EOF {
			switch {
			case len(k) != 0 && state == expectKey:
				return traverser.traverseMapKey(string(k))
			case len(k) == 0 && state == expectDenotation:
				return nil
			default:
				return fmt.Errorf("Unexpected end of")
			}
		}
		return err
	}

	var nextState int
	switch state {
	case expectIndex:
		if last != ']' {
			return fmt.Errorf("")
		}
		idx, idxErr := strconv.Atoi(string(k))
		if idxErr != nil {
			return idxErr
		}
		if e := traverser.traverseListIdx(idx); e != nil {
			return e
		}
		nextState = expectDenotation

	case expectDenotation:
		if len(k) != 0 {
			return fmt.Errorf("")
		}
		switch last {
		case '.':
			nextState = expectKey
		case '[':
			nextState = expectIndex
		default:
			return fmt.Errorf("")
		}

	case expectKey:
		switch last {
		case '.':
			if e := traverser.traverseMapKey(string(k)); e != nil {
				return e
			}
			nextState = expectKey
		case '[':
			if e := traverser.traverseMapKey(string(k)); e != nil {
				return e
			}
			nextState = expectIndex
		default:
			return fmt.Errorf("")
		}
	}

	if e := traverseSetPath(in, traverser, nextState); e != nil {
		return e
	}
	return nil
}

// func setValueOfSetPath(m map[string]interface{}, path string, value interface{}) error {
//
// }

// copy from helm-template wich copied from helm
func mergeValues(dest map[interface{}]interface{}, src map[interface{}]interface{}) map[interface{}]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[interface{}]interface{})
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
		destMap, isMap := dest[k].(map[interface{}]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}

type parser struct {
	sc   *bytes.Buffer
	data map[string]interface{}
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

func trustedMarshalYAML(d interface{}) string {
	s, err := yaml.Marshal(d)
	if err != nil {
		panic(err)
	}
	return string(s)
}
