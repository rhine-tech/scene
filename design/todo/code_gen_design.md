比如我自定义一个
```
struct {
A type name1
B type name2
C type ignored
}
```
然后`MethodA(para1)`  `select name1, name2 from A where x`

最后生成一个class

```go
func MethodA(para1) {
	return value
}
```