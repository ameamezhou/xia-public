package xconfig

import (
	"fmt"
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
func (c *WeConfig) GetValueInt(section string, key string) (int, error) {
	c.lock.Lock()
	if _, ok := c.sections[section]; !ok {
		return 0, fmt.Errorf("the section %s not exit! please check the input", section)
	}

	if _, ok := c.sections[section].keyValue[key]; !ok {
		return 0, fmt.Errorf("the key %s not exit! please check the input", key)
	}
	result := c.sections[section].keyValue[key]
	c.lock.Unlock()
	i, e := strconv.Atoi(result)
	if e != nil {
		return 0, e
	}
	return i, nil
}

// GetValue default get value, out put string
func (c *WeConfig) GetValue(section string, key string) (string, error) {
	c.lock.Lock()
	if _, ok := c.sections[section]; !ok {
		return "", fmt.Errorf("the section %s not exit! please check the input", section)
	}

	if _, ok := c.sections[section].keyValue[key]; !ok {
		return "", fmt.Errorf("the key %s not exit! please check the input", key)
	}
	result := c.sections[section].keyValue[key]
	c.lock.Unlock()
	return result, nil
}

func (c *WeConfig) GetValueFloat64(section string, key string) (float64, error) {
	c.lock.Lock()
	if _, ok := c.sections[section]; !ok {
		return 0, fmt.Errorf("the section %s not exit! please check the input", section)
	}

	if _, ok := c.sections[section].keyValue[key]; !ok {
		return 0, fmt.Errorf("the key %s not exit! please check the input", key)
	}
	result := c.sections[section].keyValue[key]
	c.lock.Unlock()
	f, e := strconv.ParseFloat(result, 64)
	if e != nil {
		return 0, e
	}
	return f, nil
}
