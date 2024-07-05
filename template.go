package xiawuyue

import (
	"bytes"
	"errors"
	"html/template"
	"net/http"
	"os"
	"strings"
	"sync"
)

type BuildTemplate struct {
	BaseDir     string           // base dir
	EnableCache bool             // 是否启用缓存
	FuncMap     template.FuncMap // template 中用到的 Func map
	// 在 Go 的 `html/template` 包中，`template.FuncMap` 是一个映射表，用于存储模板函数及其对应的实现函数。具体来说，它可以用于在模板中注册自定义的函数，以便在模板渲染时使用。
	//
	//`template.FuncMap` 的作用是将模板函数名与实际的函数实现绑定起来，以便在模板中使用。例如，我们可以定义一个 `add` 函数实现加法运算，然后将其与模板函数名 `add` 绑定起来，这样在模板中就可以使用 `{{ add 1 2 }}` 的语法进行加法运算。

	cacheMap map[string]*template.Template // 缓存map
	wrMux    sync.RWMutex                  // 模版读写锁   读写之前要先拿锁
}

func (t *BuildTemplate) baseDirCheck() error {
	if t.BaseDir == "" {
		return errors.New("you should set template's base dir attr")
	}
	if !strings.HasSuffix(t.BaseDir, "/") {
		t.BaseDir += "/"
	}
	return nil
}

func (t *BuildTemplate) Get(name string) (*template.Template, error) {
	var err error
	err = t.baseDirCheck()
	if err != nil {
		return nil, err
	}
	path := t.BaseDir + name
	var p *template.Template
	if t.EnableCache {
		var ok = false
		t.wrMux.RLock()
		if t.cacheMap != nil {
			p, ok = t.cacheMap[name]

		}
		t.wrMux.RUnlock()
		if ok {
			return p, nil
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if t.FuncMap != nil {
		p = template.Must(template.New(name).Funcs(t.FuncMap).Parse(string(data)))
	} else {
		p = template.Must(template.New(name).Parse(string(data)))
	}

	if t.EnableCache {
		t.wrMux.Lock()
		if t.cacheMap == nil {
			t.cacheMap = make(map[string]*template.Template)
		}
		t.cacheMap[name] = p
		t.wrMux.Unlock()
	}

	return p, nil
}

// template 写入 writer

func (t *BuildTemplate) WriterBuffer(w http.ResponseWriter, filename string, data ResponseXia) {
	p, err := t.Get(filename)
	if err != nil {
		panic(err)
	}

	buffer := &bytes.Buffer{}
	err = p.Execute(buffer, data)
	if err != nil {
		panic(err)
	}
	// 设置 http header
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Write(buffer.Bytes())
}
