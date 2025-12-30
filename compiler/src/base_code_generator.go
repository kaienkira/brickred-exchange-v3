package main

import (
	"fmt"
	"strings"
)

type BaseCodeGenerator struct {
	descriptor *ProtocolDescriptor
	newLineStr string
}

func (this *BaseCodeGenerator) init(
	descriptor *ProtocolDescriptor, newLineType NewLineType) {

	this.descriptor = descriptor
	if newLineType == NewLineType_Dos {
		this.newLineStr = "\r\n"
	} else {
		this.newLineStr = "\n"
	}
}

func (this *BaseCodeGenerator) close() {
	this.descriptor = nil
}

func (this *BaseCodeGenerator) writeLine(
	sb *strings.Builder, line string) {

	sb.WriteString(line)
	sb.WriteString(this.newLineStr)
}

func (this *BaseCodeGenerator) writeLineFormat(
	sb *strings.Builder, format string, args ...any) {

	fmt.Fprintf(sb, format, args...)
	sb.WriteString(this.newLineStr)
}

func (this *BaseCodeGenerator) writeEmptyLine(
	sb *strings.Builder) {

	sb.WriteString(this.newLineStr)
}
