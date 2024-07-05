package xiawuyue

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

// 定义一种 value 别名来格式化 request.Form 的内容 进行处理写入我们定义的结构体

type QiuWu map[string][]string

// 这里在独立出 router 之后 发现传入参数需要带的东西太多了，所以这里决定搞个上下文管理器，统一管理这些一次请求需要带上的全部内容

type Context struct {
	// http 库中基础的内容  在一次请求中作为上下文管理器  统一处理全部的逻辑内容
	Writer http.ResponseWriter
	Req    *http.Request

	// 访问需要带上的 pattern
	Pattern string
	Method  string
	Params  map[string]string
	// response info
	StatusCode int

	// middlewares 中间件控制
	middlewares []HandlerFunc
	index       int
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:  w,
		Req:     r,
		Pattern: r.URL.Path,
		Method:  r.Method,
		index:   -1,
	}
}

func (c *Context) NextHandle() {
	c.index++
	n := len(c.middlewares)
	for ; c.index < n; c.index++ {
		c.middlewares[c.index](c)
	}
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) Fail(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json; charset=UTF-8")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	c.Status(code)
	// 写入 UTF-8 编码声明
	c.Writer.Write([]byte("<meta charset='utf-8'>"))
	c.Writer.Write([]byte(html))
}

func (c *Context) WriteTpl(t *BuildTemplate, filename string, data ResponseXia) {
	t.WriterBuffer(c.Writer, filename, data)
}

// 将form写入 struct

func (c *Context) FormUnmarshal(data interface{}) error {
	//http.Request.ParseMultipartForm(defaultMaxMemory) 一般不需要限制最大内存
	return (*QiuWu)(&c.Req.Form).Unmarshal(data)
}

func (q *QiuWu) Unmarshal(data interface{}) error {

	// 非结构体的情况
	switch data.(type) {
	case *map[string][]string:
		var m map[string][]string = *q
		for k, v := range m {
			(*data.(*map[string][]string))[k] = v
		}
		return nil

	case *map[string]string:
		var m map[string][]string = *q
		for k, v := range m {
			(*data.(*map[string]string))[k] = v[0]
		}
		return nil
	}

	// 考虑结构体情况
	rv := reflect.ValueOf(data)

	/*
		//IsValid报告v是否表示一个值。
		//如果v为零值，则返回false。
		//如果IsValid返回false，则除String之外的所有其他方法都会死机。
		//大多数函数和方法从不返回无效值。
		//如果有，其文档会明确说明条件。
		首先要是合法的值
	*/
	if !rv.IsValid() {
		return errors.New("value struct is not valid")
		//panic("value struct is not valid")
	}

	// 传入指针才能  然的话修改内容不会保存到传入的这个 data 变量
	if rv.Kind() != reflect.Ptr {
		return errors.New("data not ptr input")
	}

	// 我们修改的内容要是非空的
	if rv.IsNil() {
		return errors.New("data can't be nil")
	}

	// 进行格式转换，我们要把对应的 key 传入对应结构体对应的位置
	elem := rv.Elem()

	var slice bool
	var f reflect.Value
	var k reflect.Kind
	var err error
	for i := 0; i < elem.NumField(); i++ {
		f = elem.Field(i)
		if !f.CanSet() {
			continue
		}
		k = f.Kind()
		if k == reflect.Struct {
			fmt.Println("struct")
			// 继续填充
			err = q.Unmarshal(rv.Field(i).Addr().Interface())
			if err != nil {
				return err
			}
			continue
		}

		if k == reflect.Slice {
			slice = true
		} else {
			slice = false
		}
		name := elem.Type().Field(i).Tag.Get("form")
		if name == "" {
			name = elem.Type().Field(i).Name
		}
		vs, b := (*q)[name]
		if !b {
			continue
		}

		if slice {
			f.Set(reflect.MakeSlice(f.Type(), len(vs), len(vs)))
			for j, v := range vs {
				err = q.setValue(v, f.Index(j))
				if err != nil {
					return err
				}
			}
		} else if len(vs) > 0 {
			err = q.setValue(vs[0], f)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (q *QiuWu) setValue(s string, f reflect.Value) error {

	k := f.Type().Kind()
	switch k {
	case reflect.String:
		f.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if s == "" {
			f.SetInt(0)
			return nil
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		if f.OverflowInt(n) {
			return fmt.Errorf("type [%v] value [%s] overflow", f.Kind(), s)
		}
		f.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if s == "" {
			f.SetInt(0)
			return nil
		}
		un, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		if f.OverflowUint(un) {
			return fmt.Errorf("type [%v] value [%s] overflow", f.Kind(), s)
		}
		f.SetUint(un)
	case reflect.Float64, reflect.Float32:
		if s == "" {
			f.SetFloat(0)
			return nil
		}
		float, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		if f.OverflowFloat(float) {
			return fmt.Errorf("type [%v] value [%s] overflow", f.Kind(), s)
		}
		f.SetFloat(float)
	case reflect.Bool:
		if s == "" || s == "false" || s == "0" {
			f.SetBool(false)
		} else {
			f.SetBool(true)
		}
	}
	return nil
}
