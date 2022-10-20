package tengo

import (
	"github.com/aldor007/mort/pkg/object"
	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
)

// FileObject struct wraping objectFileObject
type FileObject struct {
	tengoLib.ObjectImpl
	Value *object.FileObject
}

// String returns object uri
func (o *FileObject) String() string {
	return o.Value.Uri.String()
}

// BinaryOp not implemented
func (o *FileObject) BinaryOp(op token.Token, rhs tengoLib.Object) (tengoLib.Object, error) {
	return nil, tengoLib.ErrInvalidOperator
}

// IsFalsy returns false if uri is empty
func (o *FileObject) IsFalsy() bool {
	return o.Value.Uri.String() == ""
}

// Equals returns true if objects url are the same
func (o *FileObject) Equals(x tengoLib.Object) bool {
	return o.String() == x.String()
}

// Copy create copy using object.FileObject.Copy
func (o *FileObject) Copy() tengoLib.Object {
	return &FileObject{
		Value: o.Value.Copy(),
	}
}

func (o *FileObject) TypeName() string {
	return "FileObject-object"
}

// IndexGet returns the value for the given key.
// * `uri` return object Url
// * `bucket` return bucket name string
// * `key` return object storage path
// * `transforms` return Transforms object on which you can execute image manipulations
// Usage in tengo
//
//	obj.key // access to object key
func (o *FileObject) IndexGet(index tengoLib.Object) (val tengoLib.Object, err error) {
	strIdx, ok := tengoLib.ToString(index)
	if !ok {
		err = tengoLib.ErrInvalidIndexType
		return
	}

	val = tengoLib.UndefinedValue
	switch strIdx {
	case "uri":
		val = &URL{Value: o.Value.Uri}
	case "bucket":
		val = &tengoLib.String{Value: o.Value.Bucket}
	case "key":
		val = &tengoLib.String{Value: o.Value.Key}
	case "transforms":
		val = &Transforms{Value: &o.Value.Transforms}
	}

	return val, nil
}

// IndexSet allow to change value on FileObject
// * `allowChangeKey`
// * `checkParent`
// * `debug`
func (o *FileObject) IndexSet(index, value tengoLib.Object) (err error) {
	strIdx, ok := tengoLib.ToString(index)
	if !ok {
		err = tengoLib.ErrInvalidIndexType
		return
	}

	switch strIdx {
	case "allowChangeKey":
		o.Value.AllowChangeKey, _ = tengoLib.ToBool(value)
	case "checkParent":
		o.Value.CheckParent, _ = tengoLib.ToBool(value)
	case "debug":
		o.Value.Debug, _ = tengoLib.ToBool(value)
	}

	return nil
}
