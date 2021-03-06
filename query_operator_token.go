package ravendb

var _ QueryToken = &QueryOperatorToken{}

type QueryOperatorToken struct {
	_queryOperator QueryOperator
}

var (
	QueryOperatorToken_AND = NewQueryOperatorToken(QueryOperator_AND)
	QueryOperatorToken_OR  = NewQueryOperatorToken(QueryOperator_OR)
)

func NewQueryOperatorToken(queryOperator QueryOperator) *QueryOperatorToken {
	return &QueryOperatorToken{
		_queryOperator: queryOperator,
	}
}

func (t *QueryOperatorToken) WriteTo(writer *StringBuilder) {
	if t._queryOperator == QueryOperator_AND {
		writer.append("and")
		return
	}

	writer.append("or")
}
