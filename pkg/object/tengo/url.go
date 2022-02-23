package tengo

import (
	"net/url"

	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
)


type URL struct {
    tengoLib.ObjectImpl
    Value *url.URL
}

func (o *URL) String() string {
    return o.Value.String()
}

func (o *URL) BinaryOp(op token.Token, rhs tengoLib.Object) (tengoLib.Object, error) {
    return nil, tengoLib.ErrInvalidOperator
}

func (o *URL) IsFalsy() bool {
    return o.Value.Host == ""
}

func (o *URL) Equals(x tengoLib.Object) bool {
	other := x.(*URL)
	return  o.Value.String() == other.Value.String()
}

func (o *URL) Copy() tengoLib.Object {
	newUrl, _ := url.Parse(o.Value.String())

    return &URL{
        Value: newUrl,
    }
}
// IndexGet returns the value for the given key.
func (o *URL) IndexGet(index tengoLib.Object) (res tengoLib.Object, err error) {
	strIdx, ok := tengoLib.ToString(index)
	if !ok {
		err = tengoLib.ErrInvalidIndexType
		return
	}


    var val tengoLib.Object
    switch strIdx {
    case "schema":
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