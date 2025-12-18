package main

type NewLineType int

const (
	NewLineType_None NewLineType = iota
	NewLineType_Unix
	NewLineType_Dos
)

type BaseCodeGenerator interface {
	Generate(descriptor *ProtocolDescriptor, newLineType NewLineType) bool
	Close()
}
