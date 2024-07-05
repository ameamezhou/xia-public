package xia

import (
	"errors"
	"fmt"
	"strings"
)

type Trie struct {
	Path     string  // 如果是结局 记录总字段
	Part     string  // 记录当前字段
	children []*Trie // 记录子节点
	isFuzzy  bool    // 判断是不是模糊匹配
}

func NewTrie() *Trie {
	return &Trie{
		children: make([]*Trie, 0),
	}
}

func (t *Trie) getTrie(part string) *Trie {
	for _, v := range t.children {
		if v.Part == part {
			return v
		}
	}
	return nil
}

func (t *Trie) matchChild(part string) *Trie {
	for _, child := range t.children {
		if child.Part == part || child.isFuzzy {
			return child
		}
	}
	return nil
}

// 所有匹配成功的节点，用于查找
func (t *Trie) matchChildren(part string) []*Trie {
	nodes := make([]*Trie, 0)
	for _, child := range t.children {
		if child.Part == part || child.isFuzzy {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

func (t *Trie) matchChildrenUseMap(part string) *Trie {
	matchMap := t.buildMap()
	if node, ok := matchMap[part]; ok {
		return node
	}
	if node, ok := matchMap[":*isFuzzy"]; ok {
		return node
	}
	return nil
}

// 构建map
func (t *Trie) buildMap() map[string]*Trie {
	result := make(map[string]*Trie)
	for _, child := range t.children {
		if child.isFuzzy {
			result[":*isFuzzy"] = child
		} else {
			result[child.Part] = child
		}
	}
	return result
}

// 模糊匹配   /xxx/xxx/:zhou   /xxx/xxx/xxx/*

func (t *Trie) insert(parts []string, path string, depth int) {
	if depth == len(parts) {

		t.Path = path
		return
	}
	var nextT *Trie
	part := parts[depth]

	nextT = t.getTrie(part)
	if nextT == nil {
		nextT = NewTrie()
		nextT.Part = part
		nextT.isFuzzy = strings.HasPrefix(part, "*") || strings.HasPrefix(part, ":")
		t.children = append(t.children, nextT)
	}
	nextT.insert(parts, path, depth+1)
}

func (t *Trie) GetAllPath() {
	for _, v := range t.children {
		fmt.Println(v.Part)
		fmt.Println(*v)
		fmt.Println("===========")
		v.GetAllPath()
	}
}

func (t *Trie) searchPath(parts []string, depth int, params interface{}) (*Trie, error) {
	if len(parts) == depth || strings.HasPrefix(t.Part, "*") {
		if t.Path == "" {
			return nil, errors.New("path is nil")
		}
		return t, nil
	}

	if params != nil {
		switch params.(type) {
		case *map[string]string:
			if t.isFuzzy && (strings.HasPrefix(t.Part, ":") || strings.HasPrefix(t.Part, "*")) {
				(*params.(*map[string]string))[t.Part[1:]] = parts[depth-1]
			}
		default:
			return nil, errors.New("params is illegal")
		}
	}

	part := parts[depth]
	children := t.matchChildren(part)

	for _, child := range children {
		result, err := child.searchPath(parts, depth+1, params)
		if err != nil {
			return nil, err
		}
		if result != nil {
			return result, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("don't have this path %s", strings.Join(parts, "/")))
}

func (t *Trie) searchPathUseMap(parts []string, depth int, params interface{}) (*Trie, error) {
	var err error
	if len(parts) == depth || strings.HasPrefix(t.Part, "*") {
		err = t.writeParams(parts, depth, params)
		if err != nil {
			return nil, err
		}
		if t.Path == "" {
			return nil, errors.New("path is nil")
		}
		return t, nil
	}

	if params != nil {
		switch params.(type) {
		case *map[string]string:
			if t.isFuzzy && strings.HasPrefix(t.Part, ":") {
				(*params.(*map[string]string))[t.Part[1:]] = parts[depth-1]
			}
		default:
			return nil, errors.New("params is illegal")
		}
	}

	part := parts[depth]
	child := t.matchChildrenUseMap(part)
	if child != nil {
		child, err = child.searchPathUseMap(parts, depth+1, params)
		if err != nil {
			return nil, err
		}
		return child, nil
	}

	return nil, errors.New(fmt.Sprintf("don't have this path %s", strings.Join(parts, "/")))
}

func (t *Trie) writeParams(parts []string, depth int, params interface{}) error {
	if params != nil {
		switch params.(type) {
		case *map[string]string:
			if t.isFuzzy && strings.HasPrefix(t.Part, ":") {
				(*params.(*map[string]string))[t.Part[1:]] = parts[depth-1]
			}

			if t.isFuzzy && strings.HasPrefix(t.Part, "*") {
				(*params.(*map[string]string))["*"] = strings.Join(parts[depth-1:], "/")
			}
		default:
			return errors.New("params is illegal")
		}
	}
	return nil
}

// func parsePattern(pattern string) []string {
// 	vs := strings.Split(pattern, "/")

// 	parts := make([]string, 0)
// 	for _, item := range vs {
// 		if item != "" {
// 			parts = append(parts, item)
// 			// 这里如果要匹配 "/path/test1/check/*/123/123/456" 这个路径的话  考虑下后面要怎么做  要怎么写入params
// 			// 我希望的是能够在找路径的时候就触发写入了   这里构建pattern的时候要想想怎么处理下
// 			//if item[0] == '*' {
// 			//	break
// 			//}
// 		}
// 	}
// 	return parts
// }
