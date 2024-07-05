package xconfig

import (
	"strconv"
	"sync"
)

// using function
type WeConfig struct {
	// sections
	sections map[string]weSection
	// groutine scure
	lock sync.Mutex
}

type weSection struct {
	keyValue map[string]string
}

// GetValueInt out put int type
func (c *WeConfig) GetValueInt(section, key string, def int) int {
	c.lock.Lock()
	if _, ok := c.sections[section]; !ok {
		return def
	}

	if _, ok := c.sections[section].keyValue[key]; !ok {
		return def
	}
	result := c.sections[section].keyValue[key]
	c.lock.Unlock()
	i, e := strconv.Atoi(result)
	if e != nil {
		return def
	}
	return i
}

// GetValue default get value, out put string
func (c *WeConfig) GetValue(section, key, def string) string {
	c.lock.Lock()
	if _, ok := c.sections[section]; !ok {
		return def
	}

	if _, ok := c.sections[section].keyValue[key]; !ok {
		return def
	}
	result := c.sections[section].keyValue[key]
	c.lock.Unlock()
	return result
}

func (c *WeConfig) GetValueFloat64(section, key string, def float64) float64 {
	c.lock.Lock()
	if _, ok := c.sections[section]; !ok {
		return def
	}

	if _, ok := c.sections[section].keyValue[key]; !ok {
		return def
	}
	result := c.sections[section].keyValue[key]
	c.lock.Unlock()
	f, e := strconv.ParseFloat(result, 64)
	if e != nil {
		return def
	}
	return f
}
