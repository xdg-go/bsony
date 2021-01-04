package bsony

import (
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func testArrayAdd(t *testing.T, addCases []AddTestCase) {
	// Array.Add is a wrapper around Doc.Add, so we can use the same input data.
	t.Run("Array.Add", func(t *testing.T) {
		for _, c := range addCases {
			// replace key in document with "0" (ASCII 0x30)
			aHex := c.D[0:10] + "30" + c.D[12:]

			a := fct.NewArray()
			a.Add(c.V)
			compareArrayHex(t, a, aHex, c.L)
			a.Release()
		}
	})

	// Test adding to array with type-specific functions.  This is needed
	// because Doc.Add delegates to the type-specific adders, but Array.Add
	// delegates to Doc.Add.  Therefore, we have to do our own type-specific
	// test for arrays.
	t.Run("Array.AddType", func(t *testing.T) {
		for _, c := range addCases {
			// replace key in document with "0" (ASCII 0x30)
			aHex := c.D[0:10] + "30" + c.D[12:]

			a := fct.NewArray()
			if addByType(a, c.V) {
				compareArrayHex(t, a, aHex, c.L)
			}
			a.Release()
		}
	})

}

// Switch adapted from Doc.Add and will need to be kept in sync if new types
// are added.
func addByType(a *Array, v interface{}) bool {
	switch x := v.(type) {
	// Type 01 - double
	case float32:
		a.AddDouble(float64(x))
	case float64:
		a.AddDouble(x)

	// Type 02 - string
	case string:
		a.AddString(x)

	// Type 03 - document
	case Doc:
		a.AddDoc(&x)
	case *Doc:
		a.AddDoc(x)

	// Type 04 - array
	case Array:
		a.AddArray(&x)
	case *Array:
		a.AddArray(x)

	// Type 05 - binary
	case primitive.Binary:
		a.AddBinary(&x)
	case *primitive.Binary:
		a.AddBinary(x)

	// Type 06 - undefined (deprecated)
	case primitive.Undefined:
		a.AddUndefined()

	// Type 07 - ObjectID
	case primitive.ObjectID:
		a.AddOID(x)

	// Type 08 - boolean
	case bool:
		a.AddBool(x)

	// Type 09 - UTC DateTime
	case primitive.DateTime:
		a.AddDateTime(x)
	case time.Time:
		a.AddDateTimeFromTime(x)
	case *time.Time:
		a.AddDateTimeFromTime(*x)

	// Type 0A - null
	case nil:
		a.AddNull()

	// Type 0B - regular expression
	case primitive.Regex:
		a.AddRegex(x)
	case *primitive.Regex:
		a.AddRegex(*x)

	// Type 0C - DBPointer (deprecated)
	case primitive.DBPointer:
		a.AddDBPointer(x)
	case *primitive.DBPointer:
		a.AddDBPointer(*x)

	// Type 0D - JavaScript code
	case primitive.JavaScript:
		a.AddJavaScript(x)

	// Type 0E - Symbol (deprecated)
	case primitive.Symbol:
		a.AddSymbol(x)

		// Type 0F - JavaScript code with scope
	case CodeWithScope:
		a.AddCodeScope(x)
	case primitive.CodeWithScope:
		// AddCodeScope doesn't convert this, only Add, so skip test
		return false

	// Type 10 - 32-bit integer
	case int32:
		a.AddInt32(x)

		// Type 11 - timestamp
	case primitive.Timestamp:
		a.AddTimestamp(x)

	// Type 12 - 64-bit integer
	case int64:
		a.AddInt64(x)

		// Type 13 - 128-bit decimal floating point
	case primitive.Decimal128:
		a.AddDecimal128(x)
	case *primitive.Decimal128:
		a.AddDecimal128(*x)

	// Type FF - Min key
	case primitive.MinKey:
		a.AddMinKey()

	// Type 7F - Max key
	case primitive.MaxKey:
		a.AddMaxKey()

	default:
		panic(fmt.Sprintf("unsupported type: %T", v))
	}

	return true
}
