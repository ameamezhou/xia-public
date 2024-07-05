package xiawuyue

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"github.com/ameamezhou/xiawuyue/xlog"

)

// 引入前缀树
type router struct {
	roots    map[string]*Trie
	handlers map[string]HandlerFunc
}

// roots key eg, roots['GET'] roots['POST']
// handlers key eg, handlers['GET-/p/:lang/doc'], handlers['POST-/p/book']

func newRouter() *router {
	return &router{
		roots:    make(map[string]*Trie),
		handlers: make(map[string]HandlerFunc),
	}
}

// Only one * is allowed
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")

	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (r *router) addRouter(method string, pattern string, handler HandlerFunc) {
	if len(pattern) > 0 && pattern[0] != '/' {
		pattern = "/" + pattern
	}
	parts := parsePattern(pattern)

	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &Trie{}
	}
	xlog.Infof("Add new router %4s - %s", method, pattern)
	r.roots[method].insert(parts, pattern, 0)
	r.handlers[key] = handler
}

func (r *router) getRouter(method string, path string) (*Trie, map[string]string, error) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root, ok := r.roots[method]

	if !ok {
		return nil, nil, fmt.Errorf("don't have the %s - %s", method, path)
	}

	t, err := root.searchPath(searchParts, 0, &params)
	if err != nil {
		return nil, nil, err
	}

	if t != nil {
		return t, params, nil
	}

	return nil, nil, nil
}

func (r *router) handle(c *Context) {
	t, params, err := r.getRouter(c.Method, c.Pattern)
	if err != nil {
		xlog.Error(err)
	}
	if t != nil {
		c.Params = params
		key := c.Method + "-" + t.Path
		c.middlewares = append(c.middlewares, r.handlers[key])
	} else {
		c.middlewares = append(c.middlewares, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Pattern)
		})
	}
	// c.NextHandle()
	Recovery()(c)
}

func TimeLogger(c *Context) {
	t := time.Now()
	c.NextHandle()
	xlog.Debugf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
}
