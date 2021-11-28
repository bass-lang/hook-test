package bass

import (
	"context"
	"fmt"
)

// A path created by a thunk.
type ThunkPath struct {
	Thunk Thunk         `json:"thunk"`
	Path  FileOrDirPath `json:"path"`
}

var _ Value = ThunkPath{}

func (value ThunkPath) String() string {
	return fmt.Sprintf("%s/%s", value.Thunk, value.Path)
}

func (value ThunkPath) Equal(other Value) bool {
	var o ThunkPath
	return other.Decode(&o) == nil &&
		value.Path.ToValue().Equal(o.Path.ToValue())
}

func (value *ThunkPath) UnmarshalJSON(payload []byte) error {
	var obj *Scope
	err := UnmarshalJSON(payload, &obj)
	if err != nil {
		return err
	}

	return value.FromValue(obj)
}

func (value ThunkPath) Decode(dest interface{}) error {
	switch x := dest.(type) {
	case *ThunkPath:
		*x = value
		return nil
	case *Path:
		*x = value
		return nil
	case *Value:
		*x = value
		return nil
	case *Applicative:
		*x = value
		return nil
	case *Combiner:
		*x = value
		return nil
	case Decodable:
		return x.FromValue(value)
	default:
		return DecodeError{
			Source:      value,
			Destination: dest,
		}
	}
}

func (value *ThunkPath) FromValue(val Value) error {
	var obj *Scope
	if err := val.Decode(&obj); err != nil {
		return fmt.Errorf("%T.FromValue: %w", value, err)
	}

	return decodeStruct(obj, value)
}

// Eval returns the value.
func (value ThunkPath) Eval(_ context.Context, _ *Scope, cont Cont) ReadyCont {
	return cont.Call(value, nil)
}

var _ Applicative = ThunkPath{}

func (app ThunkPath) Unwrap() Combiner {
	return PathOperative{app}
}

var _ Combiner = ThunkPath{}

func (combiner ThunkPath) Call(ctx context.Context, val Value, scope *Scope, cont Cont) ReadyCont {
	return Wrap(PathOperative{combiner}).Call(ctx, val, scope, cont)
}

var _ Path = ThunkPath{}

func (path ThunkPath) Extend(ext Path) (Path, error) {
	extended := path

	var err error
	extended.Path, err = path.Path.Extend(ext)
	if err != nil {
		return nil, err
	}

	return extended, nil
}