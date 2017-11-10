package sh

import "fmt"

//go:generate stringer -type=objType
const (
	StringType objType = iota + 1
	FnType
	ListType
)

type (
	objType int

	Obj interface {
		Type() objType
		String() string
	}

	ListObj struct {
		objType
		list []Obj
	}

	FnObj struct {
		objType
		fn FnDef
	}

	StrObj struct {
		objType
		runes []rune
	}

	Collection interface {
		Len() int
		Get(index int) (Obj, error)
	}

	WriteableCollection interface {
		Set(index int, val Obj) error
	}
)

func NewCollection(o Obj) (Collection, error) {
	sizer, ok := o.(Collection)
	if !ok {
		return nil, fmt.Errorf(
			"SizeError: trying to get size from type %s which is not a collection",
			o.Type(),
		)
	}
	return sizer, nil
}

func NewWriteableCollection(o Obj) (WriteableCollection, error) {
	indexer, ok := o.(WriteableCollection)
	if !ok {
		return nil, fmt.Errorf(
			"IndexError: trying to use a non write/indexable type %s to write on index: ",
			o.Type(),
		)
	}
	return indexer, nil
}

func (o objType) Type() objType {
	return o
}

func NewStrObj(val string) *StrObj {
	return &StrObj{
		runes:   []rune(val),
		objType: StringType,
	}
}

func (o *StrObj) Str() string { return string(o.runes) }

func (o *StrObj) String() string { return o.Str() }

func (o *StrObj) Get(index int) (Obj, error) {
	// FIXME: Use runes instead
	if index >= o.Len() {
		return nil, fmt.Errorf(
			"IndexError: Index[%d] out of range, string size[%d]",
			index,
			o.Len(),
		)
	}

	return NewStrObj(string(o.runes[index])), nil
}

func (o *StrObj) Len() int {
	return len(o.runes)
}

func NewFnObj(val FnDef) *FnObj {
	return &FnObj{
		fn:      val,
		objType: FnType,
	}
}

func (o *FnObj) Fn() FnDef { return o.fn }

func (o *FnObj) String() string { return fmt.Sprintf("<fn %s>", o.fn.Name()) }

func NewListObj(val []Obj) *ListObj {
	return &ListObj{
		list:    val,
		objType: ListType,
	}
}

func (o *ListObj) Len() int {
	return len(o.list)
}

func (o *ListObj) Set(index int, value Obj) error {
	if index >= len(o.list) {
		return fmt.Errorf(
			"IndexError: Index[%d] out of range, list size[%d]",
			index,
			len(o.list),
		)
	}
	o.list[index] = value
	return nil
}

func (o *ListObj) Get(index int) (Obj, error) {
	if index >= len(o.list) {
		return nil, fmt.Errorf(
			"IndexError: Index out of bounds, index[%d] but list size[%d]",
			index,
			len(o.list),
		)
	}
	return o.list[index], nil
}

func (o *ListObj) List() []Obj { return o.list }

func (o *ListObj) String() string {
	result := ""
	list := o.List()
	for i := 0; i < len(list); i++ {
		l := list[i]

		result += l.String()

		if i < len(list)-1 {
			result += " "
		}
	}

	return result
}
