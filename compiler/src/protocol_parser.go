package main

import (
	"fmt"
	"github.com/antchfx/xmlquery"
	"os"
	"path/filepath"
	"strings"
)

type ProtocolParser struct {
	Descriptor *ProtocolDescriptor
}

func (this *ProtocolParser) Parse(
	protoFilePath string, protoSearchPath []string) bool {
	this.Descriptor = new(ProtocolDescriptor)
	this.Descriptor.ImportedProtos = make(map[string]*ProtocolDef)
	this.Descriptor.ProtoDef =
		this.parseProtocol(protoFilePath, protoSearchPath)

	if this.Descriptor.ProtoDef == nil {
		return false
	}

	return true
}

func (this *ProtocolParser) getProtoFileFullPath(
	protoFilePath string, protoSearchPath []string) string {
	fileExists := false
	// find proto file path directly first
	if utilCheckFileExists(protoFilePath) {
		fileExists = true
	} else {
		// find in the search path
		for _, path := range protoSearchPath {
			checkPath := filepath.Join(path, protoFilePath)
			if utilCheckFileExists(checkPath) {
				fileExists = true
				protoFilePath = checkPath
				break
			}
		}
	}

	if fileExists {
		return utilGetFullPath(protoFilePath)
	} else {
		return ""
	}
}

func (this *ProtocolParser) loadProtoFile(filePath string) *xmlquery.Node {
	fileBin, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"error: can not read protocol file `%s`: %s\n",
			filePath, err.Error())
		return nil
	}
	fileText := string(fileBin)

	doc, err := xmlquery.Parse(strings.NewReader(fileText))
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"error: can not parse protocol file `%s`: %s\n",
			filePath, err.Error())
		return nil
	}

	return doc
}

func (this *ProtocolParser) parseProtocol(
	protoFilePath string, protoSearchPath []string) *ProtocolDef {
	var protoDef *ProtocolDef = nil

	// check is already imported
	protoName := utilGetFileNameWithoutExtension(protoFilePath)
	if protoDef, ok := this.Descriptor.ImportedProtos[protoName]; ok {
		return protoDef
	}

	// get file full path
	protoFileFullPath := this.getProtoFileFullPath(protoFilePath, protoSearchPath)
	if protoFileFullPath == "" {
		fmt.Fprintf(os.Stderr,
			"error: can not find protocol file `%s`\n",
			protoFilePath)
		return nil
	}

	// load xml doc
	doc := this.loadProtoFile(protoFileFullPath)
	if doc == nil {
		return nil
	}

	return protoDef
}
