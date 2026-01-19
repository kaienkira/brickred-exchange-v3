package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/antchfx/xmlquery"
)

type ProtocolParser struct {
	Descriptor *ProtocolDescriptor
}

func NewProtocolParser() *ProtocolParser {
	newObj := new(ProtocolParser)

	return newObj
}

func (this *ProtocolParser) Close() {
	if this.Descriptor != nil {
		this.Descriptor.Close()
		this.Descriptor = nil
	}
}

func (this *ProtocolParser) Parse(
	protoFilePath string, protoSearchPath []string) bool {

	this.Descriptor = NewProtocolDescriptor()
	this.Descriptor.ProtoDef =
		this.parseProtocol(protoFilePath, protoSearchPath)
	if this.Descriptor.ProtoDef == nil {
		return false
	}

	return true
}

func (this *ProtocolParser) isStrValidVarName(str string) bool {
	return g_isVarNameRegexp.MatchString(str)
}

func (this *ProtocolParser) isStrNumber(str string) bool {
	return g_isNumberRegexp.MatchString(str)
}

func (this *ProtocolParser) printLineError(
	fileName string, lineNumber int, format string, args ...any) {

	fmt.Fprintf(os.Stderr,
		"error:%s:%d: %s\n",
		fileName, lineNumber,
		fmt.Sprintf(format, args...))
}

func (this *ProtocolParser) printNodeError(
	protoDef *ProtocolDef, node *xmlquery.Node,
	format string, args ...any) {

	this.printLineError(
		protoDef.FilePath, node.LineNumber, format, args...)
}

func (this *ProtocolParser) getNodeAttr(
	node *xmlquery.Node, attrName string) *xmlquery.Attr {

	for _, attr := range node.Attr {
		if attr.Name.Local == attrName {
			return &attr
		}
	}

	return nil
}

func (this *ProtocolParser) getProtoFileFullPath(
	protoFilePath string, protoSearchPath []string) string {

	fileExists := false
	// find proto file path directly first
	if UtilCheckFileExists(protoFilePath) {
		fileExists = true
	} else {
		// find in the search path
		for _, path := range protoSearchPath {
			checkPath := filepath.Join(path, protoFilePath)
			if UtilCheckFileExists(checkPath) {
				fileExists = true
				protoFilePath = checkPath
				break
			}
		}
	}

	if fileExists {
		return UtilGetFullPath(protoFilePath)
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

	xmlDoc, err := xmlquery.ParseWithOptions(strings.NewReader(fileText),
		xmlquery.ParserOptions{
			WithLineNumbers: true,
		})
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"error: can not parse protocol file `%s`: %s\n",
			filePath, err.Error())
		return nil
	}

	return xmlDoc
}

func (this *ProtocolParser) parseProtocol(
	protoFilePath string, protoSearchPath []string) *ProtocolDef {

	var protoDef *ProtocolDef = nil

	// check is already imported
	protoName := UtilGetFileNameWithoutExtension(protoFilePath)
	if protoDef, ok := this.Descriptor.ImportedProtos[protoName]; ok {
		return protoDef
	}

	// get file full path
	protoFileFullPath := this.getProtoFileFullPath(
		protoFilePath, protoSearchPath)
	if protoFileFullPath == "" {
		fmt.Fprintf(os.Stderr,
			"error: can not find protocol file `%s`\n",
			protoFilePath)
		return nil
	}

	// load xml xmlDoc
	xmlDoc := this.loadProtoFile(protoFileFullPath)
	if xmlDoc == nil {
		return nil
	}

	protoDef = NewProtocolDef(protoName, protoFileFullPath)

	// add to imported cache first to prevent circular import
	this.Descriptor.ImportedProtos[protoName] = protoDef

	// check root node name
	var rootNode *xmlquery.Node = nil
	for _, child := range xmlDoc.ChildNodes() {
		if child.Type == xmlquery.ElementNode {
			rootNode = child
			break
		}
	}
	if rootNode == nil ||
		rootNode.Type != xmlquery.ElementNode ||
		rootNode.Data != "protocol" {
		this.printNodeError(protoDef, rootNode,
			"root node must be `protocol` node")
		return nil
	}

	// parse imports
	{
		nodes := xmlquery.Find(rootNode, "/import")
		for _, node := range nodes {
			// check import self
			refProtoPath := node.InnerText()
			refProtoName := UtilGetFileNameWithoutExtension(refProtoPath)
			if refProtoName == protoName {
				this.printNodeError(protoDef, node,
					"can not import self")
				return nil
			}
			externalProtoDef := this.parseProtocol(
				refProtoPath, protoSearchPath)
			if externalProtoDef == nil {
				this.printNodeError(protoDef, node,
					"load external file `%s` failed",
					refProtoPath)
				return nil
			}

			if this.addImportDef(protoDef, node, externalProtoDef) == false {
				return nil
			}
		}
	}

	// parse namespaces
	{
		nodes := xmlquery.Find(rootNode, "/namespace")
		for _, node := range nodes {
			if this.addNamespaceDef(protoDef, node) == false {
				return nil
			}
		}
	}

	// parse enums
	{
		nodes := xmlquery.Find(rootNode, "/enum")
		for _, node := range nodes {
			if this.addEnumDef(protoDef, node) == false {
				return nil
			}
		}
	}

	// parse structs
	{
		nodes := xmlquery.Find(rootNode, "/struct")
		for _, node := range nodes {
			if this.addStructDef(protoDef, node) == false {
				return nil
			}
		}
	}

	// parse enum maps
	{
		nodes := xmlquery.Find(rootNode, "/enum_map")
		for _, node := range nodes {
			if this.addEnumMapDef(protoDef, node) == false {
				return nil
			}
		}
	}

	this.processImportedProtocols(protoDef)

	return protoDef
}

func (this *ProtocolParser) addImportDef(
	protoDef *ProtocolDef, node *xmlquery.Node,
	externalProtoDef *ProtocolDef) bool {

	if _, ok := protoDef.ImportNameIndex[externalProtoDef.Name]; ok {
		this.printNodeError(protoDef, node,
			"import `%s` duplicated", externalProtoDef.Name)
		return false
	}

	def := NewImportDef(protoDef, externalProtoDef.Name, node.LineNumber)
	def.ProtoDef = externalProtoDef

	protoDef.Imports = append(protoDef.Imports, def)
	protoDef.ImportNameIndex[def.Name] = def

	return true
}

func (this *ProtocolParser) addNamespaceDef(
	protoDef *ProtocolDef, node *xmlquery.Node) bool {

	// check lang attr
	var lang string
	{
		attr := this.getNodeAttr(node, "lang")
		if attr == nil {
			this.printNodeError(protoDef, node,
				"`namespace` node must contain a `lang` attribute")
			return false
		}
		lang = attr.Value
		if lang == "" {
			this.printNodeError(protoDef, node,
				"`namespace` node `lang` attribute is invalid")
			return false
		}
	}
	if _, ok := protoDef.Namespaces[lang]; ok {
		this.printNodeError(protoDef, node,
			"`namespace` node `lang` attribute duplicated")
		return false
	}

	// check namespace value
	namespaceStr := node.InnerText()
	if namespaceStr == "" {
		this.printNodeError(protoDef, node,
			"`namespace` node value can not be empty")
		return false
	}

	// check namespace parts
	namespaceParts := strings.Split(namespaceStr, ".")
	for _, part := range namespaceParts {
		if this.isStrValidVarName(part) == false {
			this.printNodeError(protoDef, node,
				"`namespace` node value is invalid")
			return false
		}
	}

	def := NewNamespaceDef(protoDef, lang, node.LineNumber)
	def.Namespace = namespaceStr
	def.NamespaceParts = namespaceParts

	protoDef.Namespaces[def.Language] = def

	return true
}

func (this *ProtocolParser) addEnumDef(
	protoDef *ProtocolDef, node *xmlquery.Node) bool {

	// check name attr
	var name string
	{
		attr := this.getNodeAttr(node, "name")
		if attr == nil {
			this.printNodeError(protoDef, node,
				"`enum` node must contain a `name` attribute")
			return false
		}
		name = attr.Value
	}
	if this.isStrValidVarName(name) == false {
		this.printNodeError(protoDef, node,
			"`enum` node `name` attribute is invalid")
		return false
	}
	{
		ok := false
		if _, ok = protoDef.EnumNameIndex[name]; ok == false {
			if _, ok = protoDef.StructNameIndex[name]; ok == false {
				_, ok = protoDef.EnumMapNameIndex[name]
			}
		}
		if ok {
			this.printNodeError(protoDef, node,
				"`enum` node `name` attribute duplicated")
			return false
		}
	}

	def := NewEnumDef(protoDef, name, node.LineNumber)

	// parse items
	for _, childNode := range node.ChildNodes() {
		if childNode.Type != xmlquery.ElementNode {
			continue
		}
		if childNode.Data != "item" {
			this.printNodeError(protoDef, childNode,
				"expect a `item` node")
			return false
		}

		if this.addEnumItemDef(protoDef, def, childNode) == false {
			return false
		}
	}

	protoDef.Enums = append(protoDef.Enums, def)
	protoDef.EnumNameIndex[def.Name] = def

	return true
}

func (this *ProtocolParser) addEnumItemDef(
	protoDef *ProtocolDef, enumDef *EnumDef, node *xmlquery.Node) bool {

	// check name attr
	var name string
	{
		attr := this.getNodeAttr(node, "name")
		if attr == nil {
			this.printNodeError(protoDef, node,
				"`item` node must contain a `name` attribute")
			return false
		}
		name = attr.Value
	}
	if this.isStrValidVarName(name) == false {
		this.printNodeError(protoDef, node,
			"`item` node `name` attribute is invalid")
		return false
	}
	if _, ok := enumDef.ItemNameIndex[name]; ok {
		this.printNodeError(protoDef, node,
			"`item` node `name` attribute duplicated")
		return false
	}

	// check value attr
	value := ""
	{
		attr := this.getNodeAttr(node, "value")
		if attr != nil {
			value = attr.Value
		}
	}

	def := NewEnumItemDef(enumDef, name, node.LineNumber)

	if value == "" {
		// default
		def.Type = EnumItemType_Default
		if len(enumDef.Items) == 0 {
			def.IntValue = 0
		} else {
			def.IntValue =
				enumDef.Items[len(enumDef.Items)-1].IntValue + 1
		}
	} else if this.isStrNumber(value) {
		// int
		def.Type = EnumItemType_Int
		def.IntValue = UtilAtoi(value)
	} else {
		parts := strings.Split(value, ".")
		partsLen := len(parts)
		if partsLen == 1 {
			// current enum
			refDefName := parts[0]
			refDef, ok := enumDef.ItemNameIndex[refDefName]
			if ok == false {
				this.printNodeError(protoDef, node,
					"enum item `%s` is undefined", refDefName)
				return false
			}
			def.Type = EnumItemType_CurrentEnumRef
			def.IntValue = refDef.IntValue
			def.RefEnumItemDef = refDef

		} else if partsLen == 2 {
			// other enum in same file
			refEnumDefName := parts[0]
			refDefName := parts[1]

			refEnumDef, ok := protoDef.EnumNameIndex[refEnumDefName]
			if ok == false {
				this.printNodeError(protoDef, node,
					"enum `%s` is undefined", refEnumDefName)
				return false
			}
			refDef, ok := refEnumDef.ItemNameIndex[refDefName]
			if ok == false {
				this.printNodeError(protoDef, node,
					"enum item `%s.%s` is undefined",
					refEnumDefName, refDefName)
				return false
			}
			def.Type = EnumItemType_OtherEnumRef
			def.IntValue = refDef.IntValue
			def.RefEnumItemDef = refDef

		} else if partsLen == 3 {
			// other enum in other file
			refProtoDefName := parts[0]
			refEnumDefName := parts[1]
			refDefName := parts[2]

			refProtoDef, ok := this.Descriptor.ImportedProtos[refProtoDefName]
			if ok == false {
				this.printNodeError(protoDef, node,
					"protocol `%s` is undefined", refProtoDefName)
				return false
			}
			refEnumDef, ok := refProtoDef.EnumNameIndex[refEnumDefName]
			if ok == false {
				this.printNodeError(protoDef, node,
					"enum `%s.%s` is undefined",
					refProtoDefName, refEnumDefName)
				return false
			}
			refDef, ok := refEnumDef.ItemNameIndex[refDefName]
			if ok == false {
				this.printNodeError(protoDef, node,
					"enum item `%s.%s.%s` is undefined",
					refProtoDefName, refEnumDefName, refDefName)
				return false
			}
			def.Type = EnumItemType_OtherEnumRef
			def.IntValue = refDef.IntValue
			def.RefEnumItemDef = refDef

		} else {
			this.printNodeError(protoDef, node,
				"enum value `%s` is invalid", value)
			return false
		}
	}

	enumDef.Items = append(enumDef.Items, def)
	enumDef.ItemNameIndex[def.Name] = def

	return true
}

func (this *ProtocolParser) addStructDef(
	protoDef *ProtocolDef, node *xmlquery.Node) bool {

	// check name attr
	var name string
	{
		attr := this.getNodeAttr(node, "name")
		if attr == nil {
			this.printNodeError(protoDef, node,
				"`struct` node must contain a `name` attribute")
			return false
		}
		name = attr.Value
	}
	if this.isStrValidVarName(name) == false {
		this.printNodeError(protoDef, node,
			"`struct` node `name` attribute is invalid")
		return false
	}
	{
		ok := false
		if _, ok = protoDef.StructNameIndex[name]; ok == false {
			if _, ok = protoDef.EnumNameIndex[name]; ok == false {
				_, ok = protoDef.EnumMapNameIndex[name]
			}
		}
		if ok {
			this.printNodeError(protoDef, node,
				"`struct` node `name` attribute duplicated")
			return false
		}
	}

	def := NewStructDef(protoDef, name, node.LineNumber)

	// parse fields
	for _, childNode := range node.ChildNodes() {
		if childNode.Type != xmlquery.ElementNode {
			continue
		}
		if childNode.Data != "required" &&
			childNode.Data != "optional" {
			this.printNodeError(protoDef, childNode,
				"expect a `required` or `optional` node")
			return false
		}

		if this.addStructFieldDef(protoDef, def, childNode) == false {
			return false
		}
	}

	if def.OptionalFieldCount > 0 {
		def.OptionalByteCount = (def.OptionalFieldCount-1)/8 + 1
	}

	protoDef.Structs = append(protoDef.Structs, def)
	protoDef.StructNameIndex[def.Name] = def

	return true
}

func (this *ProtocolParser) addStructFieldDef(
	protoDef *ProtocolDef, structDef *StructDef, node *xmlquery.Node) bool {

	// check name attr
	var name string
	{
		attr := this.getNodeAttr(node, "name")
		if attr == nil {
			this.printNodeError(protoDef, node,
				"`%s` node must contain a `name` attribute", node.Data)
			return false
		}
		name = attr.Value
	}
	if this.isStrValidVarName(name) == false {
		this.printNodeError(protoDef, node,
			"`%s` node `name` attribute is invalid", node.Data)
		return false
	}
	if _, ok := structDef.FieldNameIndex[name]; ok {
		this.printNodeError(protoDef, node,
			"`%s` node `name` attribute duplicated", node.Data)
		return false
	}

	// check type attr
	var typ string
	{
		attr := this.getNodeAttr(node, "type")
		if attr == nil {
			this.printNodeError(protoDef, node,
				"`%s` node must contain a `type` attribute", node.Data)
			return false
		}
		typ = attr.Value
	}

	def := NewStructFieldDef(structDef, name, node.LineNumber)

	// get type info
	fieldTypeStr := typ
	{
		m := g_fetchListTypeRegexp.FindStringSubmatch(fieldTypeStr)
		if m != nil {
			fieldTypeStr = m[1]
			def.Type = StructFieldType_List
		}
	}

	fieldType := StructFieldType_None
	if fieldTypeStr == "i8" {
		fieldType = StructFieldType_I8
	} else if fieldTypeStr == "u8" {
		fieldType = StructFieldType_U8
	} else if fieldTypeStr == "i16" {
		fieldType = StructFieldType_I16
	} else if fieldTypeStr == "u16" {
		fieldType = StructFieldType_U16
	} else if fieldTypeStr == "i32" {
		fieldType = StructFieldType_I32
	} else if fieldTypeStr == "u32" {
		fieldType = StructFieldType_U32
	} else if fieldTypeStr == "i64" {
		fieldType = StructFieldType_I64
	} else if fieldTypeStr == "u64" {
		fieldType = StructFieldType_U64
	} else if fieldTypeStr == "i16v" {
		fieldType = StructFieldType_I16V
	} else if fieldTypeStr == "u16v" {
		fieldType = StructFieldType_U16V
	} else if fieldTypeStr == "i32v" {
		fieldType = StructFieldType_I32V
	} else if fieldTypeStr == "u32v" {
		fieldType = StructFieldType_U32V
	} else if fieldTypeStr == "i64v" {
		fieldType = StructFieldType_I64V
	} else if fieldTypeStr == "u64v" {
		fieldType = StructFieldType_U64V
	} else if fieldTypeStr == "string" {
		fieldType = StructFieldType_String
	} else if fieldTypeStr == "bytes" {
		fieldType = StructFieldType_Bytes
	} else if fieldTypeStr == "bool" {
		fieldType = StructFieldType_Bool
	} else {
		var refProtoDef *ProtocolDef = nil
		refDefName := ""

		parts := strings.Split(fieldTypeStr, ".")
		partsLen := len(parts)

		if partsLen == 1 {
			// in same file
			refProtoDef = protoDef
			refDefName = parts[0]

		} else if partsLen == 2 {
			// in other file
			refProtoDefName := parts[0]
			ok := false
			refProtoDef, ok = this.Descriptor.ImportedProtos[refProtoDefName]
			if ok == false {
				this.printNodeError(protoDef, node,
					"protocol `%s` is undefined", refProtoDefName)
				return false
			}
			refDefName = parts[1]

		} else {
			this.printNodeError(protoDef, node,
				"type `%s` is invalid", fieldTypeStr)
			return false
		}

		if refEnumDef, ok := refProtoDef.EnumNameIndex[refDefName]; ok {
			fieldType = StructFieldType_Enum
			def.RefEnumDef = refEnumDef
		} else if refStructDef, ok := refProtoDef.StructNameIndex[refDefName]; ok {
			fieldType = StructFieldType_Struct
			def.RefStructDef = refStructDef
		} else {
			this.printNodeError(protoDef, node,
				"type `%s` is undefined", refDefName)
			return false
		}
	}

	if def.Type == StructFieldType_List {
		def.ListType = fieldType
	} else {
		def.Type = fieldType
	}

	// optional
	if node.Data == "optional" {
		def.IsOptional = true
		def.OptionalFieldIndex = structDef.OptionalFieldCount
		structDef.OptionalFieldCount++
	}

	structDef.Fields = append(structDef.Fields, def)
	structDef.FieldNameIndex[def.Name] = def

	return true
}

func (this *ProtocolParser) addEnumMapDef(
	protoDef *ProtocolDef, node *xmlquery.Node) bool {

	// check name attr
	var name string
	{
		attr := this.getNodeAttr(node, "name")
		if attr == nil {
			this.printNodeError(protoDef, node,
				"`enum_map` node must contain a `name` attribute")
			return false
		}
		name = attr.Value
	}
	if this.isStrValidVarName(name) == false {
		this.printNodeError(protoDef, node,
			"`enum_map` node `name` attribute is invalid")
		return false
	}
	{
		ok := false
		if _, ok = protoDef.EnumMapNameIndex[name]; ok == false {
			if _, ok = protoDef.EnumNameIndex[name]; ok == false {
				_, ok = protoDef.StructNameIndex[name]
			}
		}
		if ok {
			this.printNodeError(protoDef, node,
				"`enum_map` node `name` attribute duplicated")
			return false
		}
	}

	def := NewEnumMapDef(protoDef, name, node.LineNumber)

	// parse items
	for _, childNode := range node.ChildNodes() {
		if childNode.Type != xmlquery.ElementNode {
			continue
		}
		if childNode.Data != "item" {
			this.printNodeError(protoDef, childNode,
				"expect a `item` node")
			return false
		}

		if this.addEnumMapItemDef(protoDef, def, childNode) == false {
			return false
		}
	}

	protoDef.EnumMaps = append(protoDef.EnumMaps, def)
	protoDef.EnumMapNameIndex[def.Name] = def

	return true
}

func (this *ProtocolParser) addEnumMapItemDef(
	protoDef *ProtocolDef, enumMapDef *EnumMapDef, node *xmlquery.Node) bool {

	// check name attr
	var name string
	{
		attr := this.getNodeAttr(node, "name")
		if attr == nil {
			this.printNodeError(protoDef, node,
				"`item` node must contain a `name` attribute")
			return false
		}
		name = attr.Value
	}
	if this.isStrValidVarName(name) == false {
		this.printNodeError(protoDef, node,
			"`item` node `name` attribute is invalid")
		return false
	}
	if _, ok := enumMapDef.ItemNameIndex[name]; ok {
		this.printNodeError(protoDef, node,
			"`item` node `name` attribute duplicated")
		return false
	}

	// check value attr
	value := ""
	{
		attr := this.getNodeAttr(node, "value")
		if attr != nil {
			value = attr.Value
		}
	}

	// check struct attr
	structValue := ""
	{
		attr := this.getNodeAttr(node, "struct")
		if attr != nil {
			structValue = attr.Value
		}
	}

	def := NewEnumMapItemDef(enumMapDef, name, node.LineNumber)

	if value == "" {
		// default
		def.Type = EnumMapItemType_Default
		if len(enumMapDef.Items) == 0 {
			def.IntValue = 0
		} else {
			def.IntValue =
				enumMapDef.Items[len(enumMapDef.Items)-1].IntValue + 1
		}
	} else if this.isStrNumber(value) {
		// int
		def.Type = EnumMapItemType_Int
		def.IntValue = UtilAtoi(value)
	} else {
		// current enum
		refDef, ok := enumMapDef.ItemNameIndex[value]
		if ok == false {
			this.printNodeError(protoDef, node,
				"enum_map item `%s` is undefined", value)
			return false
		}
		def.Type = EnumMapItemType_CurrentEnumRef
		def.IntValue = refDef.IntValue
		def.RefEnumItemDef = refDef
	}

	if len(enumMapDef.Items) > 0 &&
		def.IntValue < enumMapDef.Items[len(enumMapDef.Items)-1].IntValue {
		this.printNodeError(protoDef, node, ""+
			"`item` node `value` attribute can not be "+
			"less than previous one")
		return false
	}

	if structValue != "" {
		var refProtoDef *ProtocolDef = nil
		refDefName := ""

		parts := strings.Split(structValue, ".")
		partsLen := len(parts)

		if partsLen == 1 {
			// in same file
			refProtoDef = protoDef
			refDefName = parts[0]

		} else if partsLen == 2 {
			// in other file
			refProtoDefName := parts[0]
			ok := false
			refProtoDef, ok = this.Descriptor.ImportedProtos[refProtoDefName]
			if ok == false {
				this.printNodeError(protoDef, node,
					"protocol `%s` is undefined", refProtoDefName)
				return false
			}
			refDefName = parts[1]

		} else {
			this.printNodeError(protoDef, node,
				"struct `%s` is invalid", structValue)
			return false
		}

		refStructDef, ok := refProtoDef.StructNameIndex[refDefName]
		if ok == false {
			this.printNodeError(protoDef, node,
				"struct `%s` is undefined", refDefName)
			return false
		}

		def.RefStructDef = refStructDef

		if _, ok := enumMapDef.IdToStructIndex[def.IntValue]; ok {
			this.printNodeError(protoDef, node,
				"id `%d` is already mapped to a struct", def.IntValue)
			return false
		}
		if _, ok := enumMapDef.StructToIdIndex[def.RefStructDef]; ok {
			this.printNodeError(protoDef, node,
				"struct `%s` is already mapped to a id", def.RefStructDef.Name)
			return false
		}

		enumMapDef.IdToStructIndex[def.IntValue] = def.RefStructDef
		enumMapDef.StructToIdIndex[def.RefStructDef] = def.IntValue
	}

	enumMapDef.Items = append(enumMapDef.Items, def)
	enumMapDef.ItemNameIndex[def.Name] = def

	return true
}

func (this *ProtocolParser) processImportedProtocols(
	protoDef *ProtocolDef) {

	usedProtos := make(map[string]*ProtocolDef)
	enumRefProtos := make(map[string]*ProtocolDef)
	structRefProtos := make(map[string]*ProtocolDef)
	enumMapRefProtos := make(map[string]*ProtocolDef)

	// collect enum ref protocols
	for _, enumDef := range protoDef.Enums {
		for _, def := range enumDef.Items {
			if def.RefEnumItemDef != nil {
				refProtoDef := def.RefEnumItemDef.ParentRef.ParentRef
				usedProtos[refProtoDef.Name] = refProtoDef
				enumRefProtos[refProtoDef.Name] = refProtoDef
			}
		}
	}

	// collect struct ref protocols
	for _, structDef := range protoDef.Structs {
		for _, def := range structDef.Fields {
			if def.RefEnumDef != nil {
				refProtoDef := def.RefEnumDef.ParentRef
				usedProtos[refProtoDef.Name] = refProtoDef
				structRefProtos[refProtoDef.Name] = refProtoDef
			}
			if def.RefStructDef != nil {
				refProtoDef := def.RefStructDef.ParentRef
				usedProtos[refProtoDef.Name] = refProtoDef
				structRefProtos[refProtoDef.Name] = refProtoDef
			}
		}
	}

	// collect enum map ref protocols
	for _, enumMapDef := range protoDef.EnumMaps {
		for _, def := range enumMapDef.Items {
			if def.RefEnumItemDef != nil {
				refProtoDef := def.RefEnumItemDef.ParentRef.ParentRef
				usedProtos[refProtoDef.Name] = refProtoDef
			}
			if def.RefStructDef != nil {
				refProtoDef := def.RefStructDef.ParentRef
				usedProtos[refProtoDef.Name] = refProtoDef
				enumMapRefProtos[refProtoDef.Name] = refProtoDef
			}
		}
	}

	for _, importDef := range protoDef.Imports {
		protoName := importDef.ProtoDef.Name

		// check imported protocol is used
		if _, ok := usedProtos[protoName]; ok == false {
			fmt.Fprintf(os.Stderr,
				"warning:%s:%d: protocol `%s` is not used but imported",
				protoDef.FilePath, importDef.LineNumber, importDef.Name)
		}

		if _, ok := enumRefProtos[protoName]; ok {
			importDef.IsRefByEnum = true
		}
		if _, ok := structRefProtos[protoName]; ok {
			importDef.IsRefByStruct = true
		}
		if _, ok := enumMapRefProtos[protoName]; ok {
			importDef.IsRefByEnumMap = true
		}
	}
}
