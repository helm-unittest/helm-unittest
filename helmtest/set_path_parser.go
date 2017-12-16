package helmtest

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	yaml "gopkg.in/yaml.v2"
)

func GetValueOfSetPath(m map[interface{}]interface{}, path string) (interface{}, error) {
	return fetchValueOfPath(bytes.NewBufferString(path), m, expectKey)
}

const (
	expectKey        = iota
	expectIndex      = iota
	expectDenotation = iota
)

func get(data interface{}, key string) (interface{}, error) {
	if d, ok := data.(map[interface{}]interface{}); ok {
		return d[key], nil
	}
	dataYAML, _ := yaml.Marshal(data)
	return nil, fmt.Errorf("can't get '%s' key from a non map type:\n%s", key, dataYAML)
}

func fetchValueOfPath(in io.RuneReader, data interface{}, state int) (interface{}, error) {
	illegal := runeSet([]rune{',', '{', '}', '='})
	stop := runeSet([]rune{'.', '[', ']', ',', '{', '}', '='})
	k, last, err := runesUntil(in, stop)
	if _, ok := illegal[last]; ok {
		return nil, fmt.Errorf("")
	}

	if err != nil {
		if err == io.EOF {
			switch {
			case len(k) != 0 && state == expectKey:
				return get(data, string(k))
			case len(k) == 0 && state == expectDenotation:
				return data, nil
			default:
				return nil, fmt.Errorf("Unexpected end of")
			}
		}
		return nil, err
	}

	var next interface{}
	var nextState int
	switch state {
	case expectIndex:
		if last != ']' {
			return nil, fmt.Errorf("")
		}
		idx, idxErr := strconv.Atoi(string(k))
		if idxErr != nil {
			return nil, fmt.Errorf("")
		}
		next = data.([]interface{})[idx]
		nextState = expectDenotation

	case expectDenotation:
		if len(k) != 0 {
			return nil, fmt.Errorf("")
		}
		switch last {
		case '.':
			nextState = expectKey
		case '[':
			nextState = expectIndex
		default:
			return nil, fmt.Errorf("")
		}
		next = data

	case expectKey:
		switch last {
		case '.':
			n, e := get(data, string(k))
			if e != nil {
				return nil, e
			}
			next = n
			nextState = expectKey
		case '[':
			n, e := get(data, string(k))
			if e != nil {
				return nil, e
			}
			next = n
			nextState = expectIndex
		default:
			return nil, fmt.Errorf("")
		}
	}

	result, nextErr := fetchValueOfPath(in, next, nextState)
	if nextErr != nil {
		return nil, nextErr
	}
	return result, nil
}

// func setValueOfSetPath(m map[string]interface{}, path string, value interface{}) error {
//
// }

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
