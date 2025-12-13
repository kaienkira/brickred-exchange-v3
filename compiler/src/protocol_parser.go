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
	protoName := utilGetFileNameWithoutExtension(protoFilePath)
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

	protoDef = new(ProtocolDef)
	protoDef.Name = protoName
	protoDef.FilePath = protoFileFullPath
	protoDef.Imports = make([]*ImportDef, 0)
	protoDef.ImportNameIndex = make(map[string]*ImportDef)
	protoDef.Namespaces = make(map[string]*NamespaceDef)
	protoDef.Enums = make([]*EnumDef, 0)
	protoDef.EnumMapNameIndex = make(map[string]*EnumMapDef)
	protoDef.Structs = make([]*StructDef, 0)
	protoDef.StructNameIndex = make(map[string]*StructDef)
	protoDef.EnumMaps = make([]*EnumMapDef, 0)
	protoDef.EnumMapNameIndex = make(map[string]*EnumMapDef)

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
			refProtoName := utilGetFileNameWithoutExtension(refProtoPath)
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

	def := new(ImportDef)
	def.ParentRef = protoDef
	def.Name = externalProtoDef.Name
	def.LineNumber = node.LineNumber
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
		if utilIsValidVarName(part) == false {
			this.printNodeError(protoDef, node,
				"`namespace` node value is invalid")
			return false
		}
	}

	def := new(NamespaceDef)
	def.ParentRef = protoDef
	def.Language = lang
	def.LineNumber = node.LineNumber
	def.Namespace = namespaceStr
	def.NamespaceParts = namespaceParts

	protoDef.Namespaces[def.Language] = def

	return true
}
