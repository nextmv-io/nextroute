package nextroute

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
