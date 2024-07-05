package xconfig

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// load file

func LoadConfig(source interface{}) (*WeConfig, error) {
	switch source.(type) {
	case string:
		return load(source.(string))
	default:
		return nil, fmt.Errorf("please input the correct file path, must input string type")
	}
}

func load(path string) (*WeConfig, error) {
	var file *os.File
	var err error
	// section flag rule: must first match [section] then do other logic
	var sectionFlag = true
	var tempSection string
	file, err = os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	theSections := make(map[string]weSection)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) < 1 {
			continue
		}
		if line[0] == '#' || line[0] == ';' {
			continue
		}
		// skip before match first section
		if sectionFlag && line[0] != '[' {
			continue
		}

		if line[0] == '[' {
			if line[len(line)-1] != ']' {
				return nil, fmt.Errorf("config file error, section less '[' or ']'")
			}
			sectionFlag = false
			theSections[line[1:len(line)-1]] = weSection{make(map[string]string)}
			tempSection = line[1 : len(line)-1]
			continue
		}

		if _, ok := theSections[tempSection]; ok {
			eqindex := strings.Index(line, "=")
			if eqindex == -1 {
				return nil, fmt.Errorf("config file error, key value need use '='!")
			}
			key := strings.TrimSpace(line[0:eqindex])
			value := strings.TrimSpace(line[eqindex+1:])
			if len(value) > 2 {
				if value[0] == '"' || value[0] == '\'' {
					value = value[1:]
				}
				l := len(value)
				if value[l-1] == '"' || value[l-1] == '\'' {
					value = value[:l-1]
				}
			}

			theSections[tempSection].keyValue[key] = strings.TrimSpace(value)
		}
	}
	return &WeConfig{
		sections: theSections,
	}, nil
}
