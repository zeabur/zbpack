package plan

import (
	"testing"

	"github.com/iancoleman/strcase"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
)

func TestToWeakBoolE(t *testing.T) {
	t.Parallel()

	trueValue := []string{"true", "True", "TRUE", "1"}
	falseValue := []string{"false", "False", "FALSE", "0"}
	otherValue := []any{"owo", "uwu", "uwu", "owo", 345566, -1024, 0}

	for _, value := range trueValue {
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			v, err := ToWeakBoolE(value)
			assert.Nil(t, err)
			assert.Equal(t, true, v)
		})
	}

	for _, value := range falseValue {
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			v, err := ToWeakBoolE(value)
			assert.Nil(t, err)
			assert.Equal(t, false, v)
		})
	}

	for _, value := range otherValue {
		t.Run(cast.ToString(value), func(t *testing.T) {
			t.Parallel()

			wb, wErr := ToWeakBoolE(value)
			stb, stErr := cast.ToBoolE(value)

			assert.Equal(t, stb, wb)
			assert.Equal(t, stErr, wErr)
		})
	}
}

func TestScreamingCase(t *testing.T) {
	assert.Equal(t, "A_B_C", strcase.ToScreamingSnake("a.b.c"))
	assert.Equal(t, "A_B", strcase.ToScreamingSnake("a.b"))
	assert.Equal(t, "ZOLA_VERSION", strcase.ToScreamingSnake("zolaVersion"))
	assert.Equal(t, "ZOLA_VERSION", strcase.ToScreamingSnake("zola_version"))
	assert.Equal(t, "ZOLA_VERSION", strcase.ToScreamingSnake("zola.version"))
	assert.Equal(t, "ZOLA_VERSION", strcase.ToScreamingSnake("Zola_Version"))
}
