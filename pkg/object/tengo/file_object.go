
package tengo

import (

	"github.com/aldor007/mort/pkg/object"
	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
)


type FileObject struct {
    tengoLib.ObjectImpl
    Value *object.FileObject
}

func (o *FileObject) String() string {
    return o.Value.Uri.String()
}

func (o *FileObject) BinaryOp(op token.Token, rhs tengoLib.Object) (tengoLib.Object, error) {
    return nil, tengoLib.ErrInvalidOperator
}

func (o *FileObject) IsFalsy() bool {
    return o.Value.Uri.String() != ""
}

func (o *FileObject) Equals(x tengoLib.Object) bool {
	return false
}

func (o *FileObject) Copy() tengoLib.Object {
    return &FileObject{
        Value: o.Value.Copy(),
    }
}

func (o *FileObject) TypeName() string {
    return "FileObject-object"
}
// IndexGet returns the value for the given key.
func (o *FileObject) IndexGet(index tengoLib.Object) (res tengoLib.Object, err error) {
	strIdx, ok := tengoLib.ToString(index)
	if !ok {
		err = tengoLib.ErrInvalidIndexType
		return
	}

    var val tengoLib.Object
    switch strIdx {
    case "uri":
        val = &URL{Value: o.Value.Uri}
    case "bucket":
        val = &tengoLib.String{Value: o.Value.Bucket}
    case "key":
        val = &tengoLib.String{Value: o.Value.Key}
    case "transforms":
        val = &Transforms{Value: o.Value.Transforms}
    }


	return val, nil
}

