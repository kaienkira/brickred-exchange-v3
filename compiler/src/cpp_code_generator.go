package main

import (
	"path/filepath"
)

type CppCodeGenerator struct {
	descriptor *ProtocolDescriptor
	newLineStr string
}

func NewCppCodeGenerator() *CppCodeGenerator {
	newObj := new(CppCodeGenerator)

	return newObj
}

func (this *CppCodeGenerator) Close() {
	this.descriptor = nil
}

func (this *CppCodeGenerator) Generate(
	descriptor *ProtocolDescriptor,
	outputDir string, newLineType NewLineType) bool {

	this.descriptor = descriptor
	if newLineType == NewLineType_Dos {
		this.newLineStr = "\r\n"
	} else {
		this.newLineStr = "\n"
	}

	headerFilePath := filepath.Join(
		outputDir, this.descriptor.ProtoDef.Name+".h")
	headerFileContent := this.GenerateHeaderFile()
	if utilWriteAllText(headerFilePath, headerFileContent) == false {
		return false
	}

	sourceFilePath := filepath.Join(
		outputDir, this.descriptor.ProtoDef.Name+".cc")
	sourceFileContent := this.GenerateSourceFile()
	if utilWriteAllText(sourceFilePath, sourceFileContent) == false {
		return false
	}

	return true
}

func (this *CppCodeGenerator) GenerateHeaderFile() string {
	return ""
}

func (this *CppCodeGenerator) GenerateSourceFile() string {
	return ""
}
