package ravendb

var _ QueryToken = &GroupBySumToken{}

type GroupBySumToken struct {
	_projectedName string
	_fieldName     string
}

func NewGroupBySumToken(fieldName string, projectedName string) *GroupBySumToken {
	return &GroupBySumToken{
		_fieldName:     fieldName,
		_projectedName: projectedName,
	}
}

func GroupBySumToken_create(fieldName string, projectedName string) *GroupBySumToken {
	return NewGroupBySumToken(fieldName, projectedName)
}

func (t *GroupBySumToken) WriteTo(writer *StringBuilder) {
	writer.append("sum(")
	writer.append(t._fieldName)
	writer.append(")")

	if t._projectedName == "" {
		return
	}

	writer.append(" as ")
	writer.append(t._projectedName)
}
