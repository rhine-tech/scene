package registry

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

type testIface interface {
	HelloA()
}

type testStruct struct {
	A int
}

func (t testStruct) HelloA() {
	fmt.Println("hello a")
}

func TestUse(t *testing.T) {
	require.False(t, reflect.ValueOf(testIface(nil)).IsValid())
	require.True(t, reflect.ValueOf(testIface(testStruct{})).IsValid())
	require.True(t, reflect.ValueOf(testIface(&testStruct{})).IsValid())
	require.True(t, reflect.ValueOf((*testStruct)(nil)).IsValid())
	// empty interface
	require.False(t, canUse(testIface(nil)))
	// correct interface
	require.True(t, canUse(testIface(testStruct{})))
	// correct interface with pointer receiver
	require.True(t, canUse(testIface(&testStruct{})))
	// pointer to struct
	require.False(t, canUse((*testStruct)(nil)))
	// valid pointer to struct
	require.True(t, canUse(&testStruct{}))
	// empty struct
	require.False(t, canUse(testStruct{}))
	//// If val is a nil interface or a nil pointer, map, slice, etc.
	//if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Slice || rv.Kind() == reflect.Map || rv.Kind() == reflect.Chan || rv.Kind() == reflect.Func {
	//	if rv.IsNil() {
	//		return AcquireSingleton(val)
	//	}
	//} else if rv.IsValid() && !rv.IsZero() {
	//	return val
	//}
	//fmt.Println(reflect.ValueOf(x).IsNil())
}
