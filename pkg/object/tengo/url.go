package tengo

import (
	"net/url"

	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
)

// URL tengo struct wrapping net/url
type URL struct {
	tengoLib.ObjectImpl
	Value *url.URL
}

// Strings returns full url
func (o *URL) String() string {
	return o.Value.String()
}

// BinaryOp not implemented
func (o *URL) BinaryOp(op token.Token, rhs tengoLib.Object) (tengoLib.Object, error) {
	return nil, tengoLib.ErrInvalidOperator
}

// IsFalsy returns true if url is emptu
func (o *URL) IsFalsy() bool {
	return o.String() == ""
}

// Equals returns true if url are the same
func (o *URL) Equals(x tengoLib.Object) bool {
	other := x.(*URL)
	return o.Value.String() == other.Value.String()
}

// Copy create copy of url
func (o *URL) Copy() tengoLib.Object {
	newUrl, _ := url.Parse(o.Value.String())

	return &URL{
		Value: newUrl,
	}
}

// IndexGet returns the value for the given key.
// Avaiable operations

func (o *URL) IndexGet(index tengoLib.Object) (val tengoLib.Object, err error) {
	strIdx, ok := tengoLib.ToString(index)
	if !ok {
		err = tengoLib.ErrInvalidIndexType
		return
	}

	val = tengoLib.UndefinedValue
	switch strIdx {
	case "scheme":
		val = &tengoLib.String{
			Value: o.Value.Scheme,
		}
	case "host":
		val = &tengoLib.String{
			Value: o.Value.Host,
		}
	case "path":
		val = &tengoLib.String{
			Value: o.Value.Path,
		}
	case "rawquery":
		val = &tengoLib.String{
			Value: o.Value.RawQuery,
		}
	case "query":
		query := o.Value.Query()
		internalMap := make(map[string]tengoLib.Object)
		for k, v := range query {
			array := make([]tengoLib.Object, 0)
			for _, q := range v {
				array = append(array, &tengoLib.String{Value: q})
			}
			internalMap[k] = &tengoLib.Array{
				Value: array,
			}
		}
		val = &tengoLib.Map{Value: internalMap}

	}

	return val, nil
}

func (o *URL) TypeName() string {
	return "url-object"
}
