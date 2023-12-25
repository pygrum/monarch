package http

import (
	"errors"
	"fmt"
	"os"
)

const (
	// TODO:stage payloads as shellcode if applicable
	fmtShellcode = "shellcode"
)

type StageItem struct {
	Path     string
	Agent    string
	StagedBy string
}

type stage struct {
	fileNameMappings map[string]StageItem
}

var Stage *stage

func init() {
	Stage = &stage{
		fileNameMappings: make(map[string]StageItem),
	}
}
func (s *stage) Add(name, agent, path, stagedBy string) error {
	item, ok := s.fileNameMappings[name]
	if ok {
		if item.StagedBy != stagedBy {
			return errors.New("a player is already using this endpoint to stage an agent")
		}
	}
	s.fileNameMappings[name] = StageItem{
		Path:     path,
		Agent:    agent,
		StagedBy: stagedBy,
	}
	return nil
}

func (s *stage) Rm(name string) bool {
	if _, ok := s.fileNameMappings[name]; !ok {
		return ok
	}
	delete(s.fileNameMappings, name)
	return true
}

func (s *stage) get(name string) ([]byte, error) {
	item, ok := s.fileNameMappings[name]
	if !ok {
		return nil, fmt.Errorf("nothing staged as %s", name)
	}
	bytes, err := os.ReadFile(item.Path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read %s: %v", item.Path, err)
	}
	return bytes, nil
}

func (s *stage) View() *map[string]StageItem {
	return &s.fileNameMappings
}
