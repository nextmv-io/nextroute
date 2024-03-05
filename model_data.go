// © 2019-present nextmv.io inc

package nextroute

// ModelData is a data interface available on several model constructs. It
// allows to attach arbitrary data to a model construct.
type ModelData interface {
	// Data returns the data.
	Data() any
	// SetData sets the data.
	SetData(any)
}

type modelDataImpl struct {
	data any
}

func newModelDataImpl() modelDataImpl {
	return modelDataImpl{
		data: nil,
	}
}

func (d *modelDataImpl) Data() any {
	return d.data
}

func (d *modelDataImpl) SetData(data any) {
	d.data = data
}
