package nodejs_test

import (
	"testing"

	"github.com/moznion/go-optional"
)

func Blackhole[T any](v ...T) {}

// tl;dr basically the same
func BenchmarkOption(b *testing.B) {
	const VAL = "123456"
	val := optional.Some(VAL)

	b.Run("ReturnByUnwrap", func(*testing.B) {
		Blackhole(val.Unwrap())
	})

	b.Run("ReturnDirectly", func(*testing.B) {
		Blackhole(VAL)
	})
}

// Test if we can add an alias to a type
//
// tl;dr yes, we can
func TestPtrWrite(t *testing.T) {
	var box struct {
		val *string
	}

	string__ := "123456"
	val := &box.val
	*val = &string__

	if *box.val != string__ {
		t.Errorf("expected %s, got %s", string__, *box.val)
	}
}
