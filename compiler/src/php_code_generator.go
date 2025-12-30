package main

type PhpCodeGenerator struct {
	BaseCodeGenerator
}

func NewPhpCodeGenerator() *PhpCodeGenerator {
	newObj := new(PhpCodeGenerator)

	return newObj
}

func (this *PhpCodeGenerator) Close() {
	this.close()
}

func (this *PhpCodeGenerator) Generate(
	descriptor *ProtocolDescriptor,
	outputDir string, newLineType NewLineType) bool {

	this.init(descriptor, newLineType)

	return true
}
