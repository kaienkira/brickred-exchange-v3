package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
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
	headerFileContent := this.generateHeaderFile()
	if utilWriteAllText(headerFilePath, headerFileContent) == false {
		return false
	}

	sourceFilePath := filepath.Join(
		outputDir, this.descriptor.ProtoDef.Name+".cc")
	sourceFileContent := this.generateSourceFile()
	if utilWriteAllText(sourceFilePath, sourceFileContent) == false {
		return false
	}

	return true
}

func (this *CppCodeGenerator) writeLine(
	sb *strings.Builder, line string) {

	sb.WriteString(line)
	sb.WriteString(this.newLineStr)
}

func (this *CppCodeGenerator) writeLineFormat(
	sb *strings.Builder, format string, args ...any) {

	fmt.Fprintf(sb, format, args...)
	sb.WriteString(this.newLineStr)
}

func (this *CppCodeGenerator) writeEmptyLine(
	sb *strings.Builder) {

	sb.WriteString(this.newLineStr)
}

func (this *CppCodeGenerator) getEnumFullQualifiedName(
	enumDef *EnumDef) string {

	protoDef := enumDef.ParentRef
	namespaceDef, ok := protoDef.Namespaces["cpp"]
	if ok && len(namespaceDef.NamespaceParts) > 0 {
		return fmt.Sprintf(
			"%s::%s",
			strings.Join(namespaceDef.NamespaceParts, "::"),
			enumDef.Name)
	} else {
		return enumDef.Name
	}
}

func (this *CppCodeGenerator) getEnumItemFullQualifiedName(
	enumItemDef *EnumItemDef) string {

	enumDef := enumItemDef.ParentRef
	protoDef := enumDef.ParentRef
	namespaceDef, ok := protoDef.Namespaces["cpp"]
	if ok && len(namespaceDef.NamespaceParts) > 0 {
		return fmt.Sprintf(
			"%s::%s::%s",
			strings.Join(namespaceDef.NamespaceParts, "::"),
			enumDef.Name,
			enumItemDef.Name)
	} else {
		return fmt.Sprintf(
			"%s::%s",
			enumDef.Name,
			enumItemDef.Name)
	}
}

func (this *CppCodeGenerator) getStructFullQualifiedName(
	structDef *StructDef) string {

	protoDef := structDef.ParentRef
	namespaceDef, ok := protoDef.Namespaces["cpp"]
	if ok && len(namespaceDef.NamespaceParts) > 0 {
		return fmt.Sprintf(
			"%s::%s",
			strings.Join(namespaceDef.NamespaceParts, "::"),
			structDef.Name)
	} else {
		return structDef.Name
	}
}

func (this *CppCodeGenerator) getStructFieldCppType(
	fieldDef *StructFieldDef) string {

	checkType := StructFieldType_None
	if fieldDef.Type == StructFieldType_List {
		checkType = fieldDef.ListType
	} else {
		checkType = fieldDef.Type
	}

	cppType := ""
	if checkType == StructFieldType_I8 {
		cppType = "int8_t"
	} else if checkType == StructFieldType_U8 {
		cppType = "uint8_t"
	} else if checkType == StructFieldType_I16 ||
		checkType == StructFieldType_I16V {
		cppType = "int16_t"
	} else if checkType == StructFieldType_U16 ||
		checkType == StructFieldType_U16V {
		cppType = "uint16_t"
	} else if checkType == StructFieldType_I32 ||
		checkType == StructFieldType_I32V {
		cppType = "int32_t"
	} else if checkType == StructFieldType_U32 ||
		checkType == StructFieldType_U32V {
		cppType = "uint32_t"
	} else if checkType == StructFieldType_I64 ||
		checkType == StructFieldType_I64V {
		cppType = "int64_t"
	} else if checkType == StructFieldType_U64 ||
		checkType == StructFieldType_U64V {
		cppType = "uint64_t"
	} else if checkType == StructFieldType_String ||
		checkType == StructFieldType_Bytes {
		cppType = "std::string"
	} else if checkType == StructFieldType_Bool {
		cppType = "bool"
	} else if checkType == StructFieldType_Enum {
		cppType = this.getEnumFullQualifiedName(fieldDef.RefEnumDef) + "::type"
	} else if checkType == StructFieldType_Struct {
		cppType = this.getStructFullQualifiedName(fieldDef.RefStructDef)
	}

	if fieldDef.Type == StructFieldType_List {
		return fmt.Sprintf("std::vector<%s>", cppType)
	} else {
		return cppType
	}
}

func (this *CppCodeGenerator) generateHeaderFile() string {
	var sb strings.Builder

	this.writeDontEditComment(&sb)
	this.writeHeaderFileIncludeGuardStart(&sb)
	this.writeHeaderFileIncludeFileDecl(&sb)
	this.writeHeaderFileClassForwardDecl(&sb)
	this.writeNamespaceDeclStart(&sb)
	this.writeHeaderFileEnumDecl(&sb)
	this.writeHeaderFileStructDecl(&sb)
	this.writeHeaderFileEnumMapDecl(&sb)
	this.writeNamespaceDeclEnd(&sb)
	this.writeHeaderFileIncludeGuardEnd(&sb)

	return sb.String()
}

func (this *CppCodeGenerator) generateSourceFile() string {
	var sb strings.Builder

	this.writeDontEditComment(&sb)
	this.writeSourceFileIncludeFileDecl(&sb)
	this.writeNamespaceDeclStart(&sb)
	this.writeNamespaceDeclEnd(&sb)

	return sb.String()
}

func (this *CppCodeGenerator) writeDontEditComment(
	sb *strings.Builder) {

	this.writeLine(sb,
		"/*")
	this.writeLine(sb,
		" * Generated by brickred exchange compiler.")
	this.writeLine(sb,
		" * Do not edit unless you are sure that you know what you are doing.")
	this.writeLine(sb,
		" */")
}

func (this *CppCodeGenerator) writeNamespaceDeclStart(
	sb *strings.Builder) {

	namespaceDef, ok := this.descriptor.ProtoDef.Namespaces["cpp"]
	if ok == false {
		return
	}
	namespaceName := strings.Join(namespaceDef.NamespaceParts, "::")

	this.writeEmptyLine(sb)
	this.writeLineFormat(sb,
		"namespace %s {",
		namespaceName)
}

func (this *CppCodeGenerator) writeNamespaceDeclEnd(
	sb *strings.Builder) {

	namespaceDef, ok := this.descriptor.ProtoDef.Namespaces["cpp"]
	if ok == false {
		return
	}
	namespaceName := strings.Join(namespaceDef.NamespaceParts, "::")

	this.writeEmptyLine(sb)
	this.writeLineFormat(sb,
		"} // namespace %s",
		namespaceName)
}

func (this *CppCodeGenerator) writeHeaderFileIncludeGuardStart(
	sb *strings.Builder) {

	protoDef := this.descriptor.ProtoDef

	guardNameParts := make([]string, 0)
	guardNameParts = append(guardNameParts, "BRICKRED_EXCHANGE_GENERATED")
	namespaceDef, ok := protoDef.Namespaces["cpp"]
	if ok {
		guardNameParts = append(
			guardNameParts, namespaceDef.NamespaceParts...)
	}
	guardNameParts = append(guardNameParts,
		g_notWordRegexp.ReplaceAllString(protoDef.Name, "_"))
	guardNameParts = append(guardNameParts, "H")
	guardName := strings.ToUpper(strings.Join(guardNameParts, "_"))

	this.writeLineFormat(sb,
		"#ifndef %s",
		guardName)
	this.writeLineFormat(sb,
		"#define %s",
		guardName)
}

func (this *CppCodeGenerator) writeHeaderFileIncludeGuardEnd(
	sb *strings.Builder) {

	this.writeEmptyLine(sb)
	this.writeLine(sb,
		"#endif")
}

func (this *CppCodeGenerator) writeHeaderFileIncludeFileDecl(
	sb *strings.Builder) {

	protoDef := this.descriptor.ProtoDef
	useCStdDefH := false
	useCStdIntH := false
	useStringH := false
	useVectorH := false
	useBrickredBaseStructH := false
	useOtherProtoH := false

	if len(protoDef.Structs) > 0 {
		useCStdDefH = true
		useBrickredBaseStructH = true
	}
	if len(protoDef.EnumMaps) > 0 {
		useBrickredBaseStructH = true
	}

	for _, structDef := range protoDef.Structs {
		for _, fieldDef := range structDef.Fields {
			if fieldDef.IsOptional {
				useCStdIntH = true
			}

			checkType := StructFieldType_None
			if fieldDef.Type == StructFieldType_List {
				checkType = fieldDef.ListType
				useVectorH = true
			} else {
				checkType = fieldDef.Type
			}

			if checkType == StructFieldType_I8 ||
				checkType == StructFieldType_U8 ||
				checkType == StructFieldType_I16 ||
				checkType == StructFieldType_U16 ||
				checkType == StructFieldType_I32 ||
				checkType == StructFieldType_U32 ||
				checkType == StructFieldType_I64 ||
				checkType == StructFieldType_U64 ||
				checkType == StructFieldType_I16V ||
				checkType == StructFieldType_U16V ||
				checkType == StructFieldType_I32V ||
				checkType == StructFieldType_U32V ||
				checkType == StructFieldType_I64V ||
				checkType == StructFieldType_U64V {
				useCStdIntH = true
			} else if checkType == StructFieldType_String ||
				checkType == StructFieldType_Bytes {
				useStringH = true
			}
		}
	}

	for _, importDef := range protoDef.Imports {
		if importDef.IsRefByStruct == false &&
			importDef.IsRefByEnumMap {
			continue
		}
		useOtherProtoH = true
		break
	}

	if useCStdDefH == false &&
		useCStdIntH == false &&
		useStringH == false &&
		useVectorH == false &&
		useBrickredBaseStructH == false {
		return
	}

	if useCStdDefH || useCStdIntH || useStringH || useVectorH {
		this.writeEmptyLine(sb)
	}
	if useCStdDefH {
		this.writeLine(sb,
			"#include <cstddef>")
	}
	if useCStdIntH {
		this.writeLine(sb,
			"#include <cstdint>")
	}
	if useStringH {
		this.writeLine(sb,
			"#include <string>")
	}
	if useVectorH {
		this.writeLine(sb,
			"#include <vector>")
	}

	if useBrickredBaseStructH || useOtherProtoH {
		this.writeEmptyLine(sb)
	}
	if useBrickredBaseStructH {
		this.writeLine(sb,
			"#include <brickred/exchange/base_struct.h>")
	}
	for _, importDef := range protoDef.Imports {
		if importDef.IsRefByStruct == false &&
			importDef.IsRefByEnumMap {
			continue
		}
		this.writeLineFormat(sb,
			"#include \"%s.h\"",
			importDef.ProtoDef.Name)
	}
}

func (this *CppCodeGenerator) writeHeaderFileClassForwardDecl(
	sb *strings.Builder) {

	protoDef := this.descriptor.ProtoDef

	refStructDefs := make([]*StructDef, 0)
	for _, enumMapDef := range protoDef.EnumMaps {
		for _, enumMapItemDef := range enumMapDef.Items {
			def := enumMapItemDef.RefStructDef
			if def == nil {
				continue
			}
			if def.ParentRef == protoDef {
				continue
			}
			if slices.Contains(refStructDefs, def) {
				continue
			}
			refStructDefs = append(refStructDefs, def)
		}
	}

	if len(refStructDefs) > 0 {
		this.writeEmptyLine(sb)
	}
	for _, refStructDef := range refStructDefs {
		refProtoDef := refStructDef.ParentRef
		refNamespaceDef, ok := refProtoDef.Namespaces["cpp"]
		if ok && len(refNamespaceDef.NamespaceParts) > 0 {
			this.writeLineFormat(sb,
				"namespace %s { class %s; }",
				strings.Join(refNamespaceDef.NamespaceParts, "::"),
				refStructDef.Name)
		} else {
			this.writeLineFormat(sb,
				"class %s;",
				refStructDef.Name)
		}
	}
}

func (this *CppCodeGenerator) writeHeaderFileEnumDecl(
	sb *strings.Builder) {

	protoDef := this.descriptor.ProtoDef

	for _, def := range protoDef.Enums {
		this.writeHeaderFileOneEnumDecl(sb, def)
	}
}

func (this *CppCodeGenerator) writeHeaderFileOneEnumDecl(
	sb *strings.Builder, enumDef *EnumDef) {

	this.writeEmptyLine(sb)
	this.writeLineFormat(sb,
		"struct %s {",
		enumDef.Name)
	this.writeLine(sb,
		"    enum type {")

	for _, def := range enumDef.Items {
		if def.Type == EnumItemType_Default {
			this.writeLineFormat(sb,
				"        %s,",
				def.Name)
		} else if def.Type == EnumItemType_Int {
			this.writeLineFormat(sb,
				"        %s = %d,",
				def.Name, def.IntValue)
		} else if def.Type == EnumItemType_CurrentEnumRef {
			this.writeLineFormat(sb,
				"        %s = %s,",
				def.Name, def.RefEnumItemDef.Name)
		} else if def.Type == EnumItemType_OtherEnumRef {
			this.writeLineFormat(sb,
				"        %s = %s,",
				def.Name,
				this.getEnumItemFullQualifiedName(def.RefEnumItemDef))
		}
	}

	this.writeLine(sb,
		"    };")
	this.writeLine(sb,
		"};")
}

func (this *CppCodeGenerator) writeHeaderFileStructDecl(
	sb *strings.Builder) {

	protoDef := this.descriptor.ProtoDef

	for _, def := range protoDef.Structs {
		this.writeHeaderFileOneStructDecl(sb, def)
	}
}

func (this *CppCodeGenerator) writeHeaderFileOneStructDecl(
	sb *strings.Builder, structDef *StructDef) {

	this.writeEmptyLine(sb)
	this.writeLineFormat(sb,
		"class %s : public brickred::exchange::BaseStruct {",
		structDef.Name)
	this.writeLine(sb,
		"public:")
	this.writeLineFormat(sb,
		"    %s();",
		structDef.Name)
	this.writeLineFormat(sb,
		"    ~%s() override;",
		structDef.Name)
	this.writeLineFormat(sb,
		"    void swap(%s &other);",
		structDef.Name)
	this.writeEmptyLine(sb)
	this.writeLineFormat(sb,
		"    static brickred::exchange::BaseStruct *create() { return new %s(); }",
		structDef.Name)
	this.writeLineFormat(sb,
		"    %s *clone() const override { return new %s(*this); }",
		structDef.Name, structDef.Name)
	this.writeLine(sb,
		"    int encode(char *buffer, size_t size) const override;")
	this.writeLine(sb,
		"    int decode(const char *buffer, size_t size) override;")
	this.writeLine(sb,
		"    std::string dump() const override;")
	this.writeHeaderFileOneStructDeclOptionalFuncDecl(sb, structDef)
	this.writeHeaderFileOneStructDeclFieldDecl(sb, structDef)
	this.writeLine(sb,
		"};")
}

func (this *CppCodeGenerator) writeHeaderFileOneStructDeclOptionalFuncDecl(
	sb *strings.Builder, structDef *StructDef) {

	if structDef.OptionalFieldCount <= 0 {
		return
	}

	for _, def := range structDef.Fields {
		if def.IsOptional == false {
			continue
		}

		byteIndex := def.OptionalFieldIndex / 8
		byteMask := fmt.Sprintf("0x%02x", 1<<(def.OptionalFieldIndex%8))
		cppType := this.getStructFieldCppType(def)
		if def.Type == StructFieldType_String ||
			def.Type == StructFieldType_Bytes ||
			def.Type == StructFieldType_List ||
			def.Type == StructFieldType_Struct {
			cppType = fmt.Sprintf("const %s &", cppType)
		} else {
			cppType = fmt.Sprintf("%s ", cppType)
		}

		this.writeEmptyLine(sb)
		this.writeLineFormat(sb,
			"    bool has_%s() const { return _has_bits_[%d] & %s; }",
			def.Name, byteIndex, byteMask)
		this.writeLineFormat(sb,
			"    void set_has_%s() { _has_bits_[%d] |= %s; }",
			def.Name, byteIndex, byteMask)
		this.writeLineFormat(sb,
			"    void clear_has_%s() { _has_bits_[%d] &= ~%s; }",
			def.Name, byteIndex, byteMask)
		this.writeLineFormat(sb,
			"    void set_%s(%svalue) { set_has_%s(); this->%s = value; }",
			def.Name, cppType, def.Name, def.Name)
	}

	this.writeEmptyLine(sb)
	this.writeLine(sb,
		"private:")
	this.writeLineFormat(sb,
		"    uint8_t _has_bits_[%d];",
		structDef.OptionalByteCount)
}

func (this *CppCodeGenerator) writeHeaderFileOneStructDeclFieldDecl(
	sb *strings.Builder, structDef *StructDef) {

	if len(structDef.Fields) <= 0 {
		return
	}

	this.writeEmptyLine(sb)
	this.writeLine(sb,
		"public:")

	for _, def := range structDef.Fields {
		cppType := this.getStructFieldCppType(def)
		this.writeLineFormat(sb,
			"    %s %s;",
			cppType, def.Name)
	}
}

func (this *CppCodeGenerator) writeHeaderFileEnumMapDecl(
	sb *strings.Builder) {

	protoDef := this.descriptor.ProtoDef

	for _, def := range protoDef.EnumMaps {
		this.writeHeaderFileOneEnumMapDecl(sb, def)
	}
}

func (this *CppCodeGenerator) writeHeaderFileOneEnumMapDecl(
	sb *strings.Builder, enumMapDef *EnumMapDef) {

	this.writeEmptyLine(sb)
	this.writeLineFormat(sb,
		"struct %s {",
		enumMapDef.Name)
	this.writeLine(sb,
		"    enum type {")

	for _, def := range enumMapDef.Items {
		if def.Type == EnumMapItemType_Default {
			this.writeLineFormat(sb,
				"        %s,",
				def.Name)
		} else if def.Type == EnumMapItemType_Int {
			this.writeLineFormat(sb,
				"        %s = %d,",
				def.Name, def.IntValue)
		} else if def.Type == EnumMapItemType_CurrentEnumRef {
			this.writeLineFormat(sb,
				"        %s = %s,",
				def.Name, def.RefEnumItemDef.Name)
		}
	}

	this.writeLine(sb,
		"    };")
	this.writeEmptyLine(sb)
	this.writeLine(sb,
		"    template <class T>")
	this.writeLine(sb,
		"    struct id;")
	this.writeEmptyLine(sb)
	this.writeLine(sb,
		"    static brickred::exchange::BaseStruct *create(int id);")
	this.writeLine(sb,
		"};")

	this.writeHeaderFileOneEnumMapDeclIdTemplateDecl(sb, enumMapDef)
}

func (this *CppCodeGenerator) writeHeaderFileOneEnumMapDeclIdTemplateDecl(
	sb *strings.Builder, enumMapDef *EnumMapDef) {

	if len(enumMapDef.IdToStructIndex) <= 0 {
		return
	}

	this.writeEmptyLine(sb)

	for _, def := range enumMapDef.Items {
		if def.RefStructDef == nil {
			continue
		}

		this.writeLine(sb,
			"template <>")
		this.writeLineFormat(sb,
			"struct %s::id<%s> {",
			enumMapDef.Name,
			this.getStructFullQualifiedName(def.RefStructDef))
		this.writeLineFormat(sb,
			"    static constexpr int value = %s;",
			def.Name)
		this.writeLine(sb,
			"};")
	}
}

func (this *CppCodeGenerator) writeSourceFileIncludeFileDecl(
	sb *strings.Builder) {

	protoDef := this.descriptor.ProtoDef
	useCStringH := false
	useAlgorithmH := false
	useSStreamH := false
	useBrickredMacroInternalH := false
	useOtherProtoH := false

	if len(protoDef.EnumMaps) > 0 {
		// for std::lower_bound() in EnumMap::create()
		useAlgorithmH = true
	}

	for _, structDef := range protoDef.Structs {
		if structDef.OptionalFieldCount > 0 {
			// for memset(_has_bits_)
			useCStringH = true
			// for std::swap(_has_bits_)
			useAlgorithmH = true
		}

		if len(structDef.Fields) > 0 {
			useSStreamH = true
			useBrickredMacroInternalH = true
		}

		for _, def := range structDef.Fields {
			if def.Type != StructFieldType_List &&
				def.Type != StructFieldType_Struct {
				// for std::swap(field)
				useAlgorithmH = true
			}
		}
	}

	for _, importDef := range protoDef.Imports {
		if (importDef.IsRefByStruct == false &&
			importDef.IsRefByEnumMap) == false {
			continue
		} else {
			useOtherProtoH = true
			break
		}
	}

	if useCStringH == false &&
		useAlgorithmH == false &&
		useSStreamH == false &&
		useBrickredMacroInternalH == false {
		return
	}

	this.writeLineFormat(sb,
		"#include \"%s.h\"",
		protoDef.Name)

	if useCStringH || useAlgorithmH || useSStreamH {
		this.writeEmptyLine(sb)
	}
	if useCStringH {
		this.writeLine(sb,
			"#include <cstring>")
	}
	if useAlgorithmH {
		this.writeLine(sb,
			"#include <algorithm>")
	}
	if useSStreamH {
		this.writeLine(sb,
			"#include <sstream>")
	}

	if useBrickredMacroInternalH || useOtherProtoH {
		this.writeEmptyLine(sb)
	}
	if useBrickredMacroInternalH {
		this.writeLine(sb,
			"#include <brickred/exchange/macro_internal.h>")
	}
	for _, importDef := range protoDef.Imports {
		if (importDef.IsRefByStruct == false &&
			importDef.IsRefByEnumMap) == false {
			continue
		}
		this.writeLineFormat(sb,
			"#include \"%s.h\"",
			importDef.ProtoDef.Name)
	}
}
