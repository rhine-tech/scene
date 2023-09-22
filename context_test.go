package scene

import (
	"fmt"
	"reflect"
	"testing"
)

func TestContext_Naming(t *testing.T) {
	a := Application(nil)
	fmt.Println(reflect.TypeOf(a).Elem().String())
}
