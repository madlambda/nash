package sh

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
		fn Fn
	}

	StrObj struct {
		objType
		str string
	}
)

func (o objType) Type() objType {
	return o
}

func NewStrObj(val string) *StrObj {
	return &StrObj{
		str:     val,
		objType: StringType,
	}
}

func (o *StrObj) Str() string { return o.str }

func (o *StrObj) String() string { return o.Str() }

func NewFnObj(val Fn) *FnObj {
	return &FnObj{
		fn:      val,
		objType: FnType,
	}
}

func (o *FnObj) Fn() Fn { return o.fn }

func (o *FnObj) String() string { return "<fn " + o.Fn().Name() + ">" }

func NewListObj(val []Obj) *ListObj {
	return &ListObj{
		list:    val,
		objType: ListType,
	}
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
