package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllEnumValuesAndValidity(t *testing.T) {
	// Every Values() set must be non-empty and each member must validate.
	cases := []struct {
		values []string
		valid  func(string) bool
	}{
		{TraceStatusValues(), func(s string) bool { return TraceStatus(s).Valid() }},
		{SpanKindValues(), func(s string) bool { return SpanKind(s).Valid() }},
		{SpanStatusValues(), func(s string) bool { return SpanStatusCode(s).Valid() }},
		{EvaluationStatusValues(), func(s string) bool { return EvaluationStatus(s).Valid() }},
		{ScoreStatusValues(), func(s string) bool { return ScoreStatus(s).Valid() }},
		{ScoreSourceValues(), func(s string) bool { return ScoreSource(s).Valid() }},
		{ScoreDataTypeValues(), func(s string) bool { return ScoreDataType(s).Valid() }},
	}
	for _, c := range cases {
		assert.NotEmpty(t, c.values)
		for _, v := range c.values {
			assert.True(t, c.valid(v), "expected %q to be valid", v)
		}
		assert.False(t, c.valid("__nope__"))
	}

	// Sets without a Valid() method must still be populated.
	assert.NotEmpty(t, TraceSortByValues())
	assert.NotEmpty(t, SortOrderValues())
	assert.NotEmpty(t, SessionSortByValues())
	assert.NotEmpty(t, EvalTargetValues())
}
