package sh

//go:generate stringer -type=objType
const (
	StringType objType = iota + 1
	FnType
	ListType
)

type (
	objType int

	Obj struct {
		objType

		list []*Obj
		str  string
		fn   Fn
	}
)

func (o objType) Type() objType {
	return o
}

func NewStrObj(val string) *Obj {
	return &Obj{
		str:     val,
		objType: StringType,
	}
}

func NewListObj(val []*Obj) *Obj {
	return &Obj{
		list:    val,
		objType: ListType,
	}
}

func NewFnObj(val Fn) *Obj {
	return &Obj{
		fn:      val,
		objType: FnType,
	}
}

func (o Obj) Str() string  { return o.str }
func (o Obj) Fn() Fn       { return o.fn }
func (o Obj) List() []*Obj { return o.list }

func (o Obj) String() string {
	switch o.Type() {
	case StringType:
		return o.Str()
	case FnType:
		return "<fn " + o.Fn().Name() + ">"
	case ListType:
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

	return "<unknown>"
}
