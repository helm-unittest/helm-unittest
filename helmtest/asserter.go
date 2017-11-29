package helmtest

type Assertable interface {
	assert(manifest map[string]interface{}) error
}

type EqualAsserter struct {
	Path  string
	Value interface{}
}

func (a EqualAsserter) assert(manifest map[string]interface{}) error {
	return nil
}

type MatchRegexAsserter struct {
	Path    string
	Pattern string
}

func (a MatchRegexAsserter) assert(manifest map[string]interface{}) error {
	return nil
}

type ContainsAsserter struct {
	Path    string
	Content []interface{}
}

func (a ContainsAsserter) assert(manifest map[string]interface{}) error {
	return nil
}

type IsNullAsserter struct {
	Path string
}

func (a IsNullAsserter) assert(manifest map[string]interface{}) error {
	return nil
}

type IsEmptyAsserter struct {
	Path string
}

func (a IsEmptyAsserter) assert(manifest map[string]interface{}) error {
	return nil
}

type IsKindAsserter struct {
	of string
}

func (a IsKindAsserter) assert(manifest map[string]interface{}) error {
	return nil
}

type IsAPIVersionAsserter struct {
	of string
}

func (a IsAPIVersionAsserter) assert(manifest map[string]interface{}) error {
	return nil
}

type HasDocumentsAsserter struct {
	count int
}

func (a HasDocumentsAsserter) assert(manifest map[string]interface{}) error {
	return nil
}
