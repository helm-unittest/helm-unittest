package valueutils

import (
	"fmt"
	"strconv"
)

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
