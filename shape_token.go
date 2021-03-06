package ravendb

var _ QueryToken = &ShapeToken{}

type ShapeToken struct {
	_shape string
}

func NewShapeToken(shape string) *ShapeToken {
	return &ShapeToken{
		_shape: shape,
	}
}

func ShapeToken_circle(radiusParameterName string, latitudeParameterName string, longitudeParameterName string, radiusUnits SpatialUnits) *ShapeToken {
	if radiusUnits == "" {
		return NewShapeToken("spatial.circle($" + radiusParameterName + ", $" + latitudeParameterName + ", $" + longitudeParameterName + ")")
	}

	if radiusUnits == SpatialUnits_KILOMETERS {
		return NewShapeToken("spatial.circle($" + radiusParameterName + ", $" + latitudeParameterName + ", $" + longitudeParameterName + ", 'Kilometers')")
	}
	return NewShapeToken("spatial.circle($" + radiusParameterName + ", $" + latitudeParameterName + ", $" + longitudeParameterName + ", 'Miles')")
}

func ShapeToken_wkt(shapeWktParameterName string) *ShapeToken {
	return NewShapeToken("spatial.wkt($" + shapeWktParameterName + ")")
}

func (t *ShapeToken) WriteTo(writer *StringBuilder) {
	writer.append(t._shape)
}
