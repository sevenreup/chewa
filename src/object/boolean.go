package object

import (
	"fmt"
	"hash/fnv"
)

const BOOLEAN_OBJ = "BOOLEAN"

type Boolean struct {
	Object
	Mappable
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }

func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }

func (i *Boolean) Method(method string, args []Object) (Object, bool) {
	//TODO implement me
	panic("implement me")
}

func (s *Boolean) MapKey() MapKey {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprint(s.Value)))
	return MapKey{Type: s.Type(), Value: h.Sum64()}
}
