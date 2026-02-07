package main

type ProtocolDescriptor struct {
	ProtoDef *ProtocolDef

	// ProtocolDef.Name -> ProtocolDef
	ImportedProtos map[string]*ProtocolDef
}

func NewProtocolDescriptor() *ProtocolDescriptor {
	newObj := new(ProtocolDescriptor)
	newObj.ImportedProtos = make(map[string]*ProtocolDef)

	return newObj
}

func (this *ProtocolDescriptor) Close() {
	if this.ImportedProtos != nil {
		for _, def := range this.ImportedProtos {
			def.Close()
		}
		clear(this.ImportedProtos)
		this.ImportedProtos = nil
	}
	// already call Close() in ImportedProtos
	// this.ProtoDef don't have ownership
	if this.ProtoDef != nil {
		this.ProtoDef = nil
	}
}

// ----------------------------------------------------------------------------
type ProtocolDef struct {
	Name     string
	FilePath string

	// import define
	// in file define order
	Imports []*ImportDef
	// ImportDef.Name -> ImportDef
	ImportNameIndex map[string]*ImportDef

	// namespace define
	// language -> NamespaceDef
	Namespaces map[string]*NamespaceDef

	// enum define
	// in file define order
	Enums []*EnumDef
	// EnumDef.Name -> EnumDef
	EnumNameIndex map[string]*EnumDef

	// struct define
	// in file define order
	Structs []*StructDef
	// StructDef.Name -> StructDef
	StructNameIndex map[string]*StructDef

	// enum map define
	// in file define order
	EnumMaps []*EnumMapDef
	// EnumMapDef.Name -> EnumMapDef
	EnumMapNameIndex map[string]*EnumMapDef
}

func NewProtocolDef(name string, filePath string) *ProtocolDef {
	newObj := new(ProtocolDef)
	newObj.Name = name
	newObj.FilePath = filePath
	newObj.Imports = make([]*ImportDef, 0)
	newObj.ImportNameIndex = make(map[string]*ImportDef)
	newObj.Namespaces = make(map[string]*NamespaceDef)
	newObj.Enums = make([]*EnumDef, 0)
	newObj.EnumNameIndex = make(map[string]*EnumDef)
	newObj.Structs = make([]*StructDef, 0)
	newObj.StructNameIndex = make(map[string]*StructDef)
	newObj.EnumMaps = make([]*EnumMapDef, 0)
	newObj.EnumMapNameIndex = make(map[string]*EnumMapDef)

	return newObj
}

func (this *ProtocolDef) Close() {
	if this.EnumMapNameIndex != nil {
		clear(this.EnumMapNameIndex)
		this.EnumMapNameIndex = nil
	}
	if this.EnumMaps != nil {
		for _, def := range this.EnumMaps {
			def.Close()
		}
		clear(this.EnumMaps)
		this.EnumMaps = nil
	}
	if this.StructNameIndex != nil {
		clear(this.StructNameIndex)
		this.StructNameIndex = nil
	}
	if this.Structs != nil {
		for _, def := range this.Structs {
			def.Close()
		}
		clear(this.Structs)
		this.Structs = nil
	}
	if this.EnumNameIndex != nil {
		clear(this.EnumNameIndex)
		this.EnumNameIndex = nil
	}
	if this.Enums != nil {
		for _, def := range this.Enums {
			def.Close()
		}
		clear(this.Enums)
		this.Enums = nil
	}
	if this.Namespaces != nil {
		for _, def := range this.Namespaces {
			def.Close()
		}
		clear(this.Namespaces)
		this.Namespaces = nil
	}
	if this.ImportNameIndex != nil {
		clear(this.ImportNameIndex)
		this.ImportNameIndex = nil
	}
	if this.Imports != nil {
		for _, def := range this.Imports {
			def.Close()
		}
		clear(this.Imports)
		this.Imports = nil
	}
}

// ----------------------------------------------------------------------------
type ImportDef struct {
	// link to parent define
	ParentRef *ProtocolDef
	// import name
	Name string
	// define in line number
	LineNumber int

	ProtoDef *ProtocolDef

	// import is referenced by enum
	IsRefByEnum bool
	// import is referenced by struct
	IsRefByStruct bool
	// import is referenced by enum map
	IsRefByEnumMap bool
}

func NewImportDef(
	parentRef *ProtocolDef, name string, lineNumber int) *ImportDef {

	newObj := new(ImportDef)
	newObj.ParentRef = parentRef
	newObj.Name = name
	newObj.LineNumber = lineNumber

	return newObj
}

func (this *ImportDef) Close() {
	this.ProtoDef = nil
	this.ParentRef = nil
}

// ----------------------------------------------------------------------------
type NamespaceDef struct {
	// link to parent define
	ParentRef *ProtocolDef
	// namespace for language
	Language string
	// define in line number
	LineNumber int

	Namespace      string
	NamespaceParts []string
}

func NewNamespaceDef(
	parentRef *ProtocolDef, lang string, lineNumber int) *NamespaceDef {

	newObj := new(NamespaceDef)
	newObj.ParentRef = parentRef
	newObj.Language = lang
	newObj.LineNumber = lineNumber

	return newObj
}

func (this *NamespaceDef) Close() {
	this.NamespaceParts = nil
	this.ParentRef = nil
}

// ----------------------------------------------------------------------------
type EnumItemType int

const (
	EnumItemType_None EnumItemType = iota
	EnumItemType_Default
	EnumItemType_Int
	EnumItemType_CurrentEnumRef
	EnumItemType_OtherEnumRef
)

// ----------------------------------------------------------------------------
type EnumItemDef struct {
	// link to parent define
	ParentRef *EnumDef
	// enum item name
	Name string
	// define in line number
	LineNumber int

	Type           EnumItemType
	IntValue       int
	RefEnumItemDef *EnumItemDef
}

func NewEnumItemDef(
	parentRef *EnumDef, name string, lineNumber int) *EnumItemDef {

	newObj := new(EnumItemDef)
	newObj.ParentRef = parentRef
	newObj.Name = name
	newObj.LineNumber = lineNumber

	return newObj
}

func (this *EnumItemDef) Close() {
	this.RefEnumItemDef = nil
	this.ParentRef = nil
}

// ----------------------------------------------------------------------------
type EnumDef struct {
	// link to parent define
	ParentRef *ProtocolDef
	// enum name
	Name string
	// define in line number
	LineNumber int

	// in file define order
	Items []*EnumItemDef
	// EnumItemDef.Name -> EnumItemDef
	ItemNameIndex map[string]*EnumItemDef
}

func NewEnumDef(
	parentRef *ProtocolDef, name string, lineNumber int) *EnumDef {

	newObj := new(EnumDef)
	newObj.ParentRef = parentRef
	newObj.Name = name
	newObj.LineNumber = lineNumber
	newObj.Items = make([]*EnumItemDef, 0)
	newObj.ItemNameIndex = make(map[string]*EnumItemDef)

	return newObj
}

func (this *EnumDef) Close() {
	if this.ItemNameIndex != nil {
		clear(this.ItemNameIndex)
		this.ItemNameIndex = nil
	}
	if this.Items != nil {
		for _, def := range this.Items {
			def.Close()
		}
		clear(this.Items)
		this.Items = nil
	}
	this.ParentRef = nil
}

// ----------------------------------------------------------------------------
type StructFieldType int

const (
	StructFieldType_None StructFieldType = iota
	StructFieldType_I8
	StructFieldType_U8
	StructFieldType_I16
	StructFieldType_U16
	StructFieldType_I32
	StructFieldType_U32
	StructFieldType_I64
	StructFieldType_U64
	StructFieldType_I16V
	StructFieldType_U16V
	StructFieldType_I32V
	StructFieldType_U32V
	StructFieldType_I64V
	StructFieldType_U64V
	StructFieldType_String
	StructFieldType_Bytes
	StructFieldType_Bool
	StructFieldType_Enum
	StructFieldType_Struct
	StructFieldType_List
)

func StructFieldTypeIsInteger(t StructFieldType) bool {
	return t >= StructFieldType_I8 && t <= StructFieldType_U64V
}

// ----------------------------------------------------------------------------
type StructFieldDef struct {
	// link to parent define
	ParentRef *StructDef
	// field name
	Name string
	// define in line number
	LineNumber int

	Type               StructFieldType
	ListType           StructFieldType
	RefEnumDef         *EnumDef
	RefStructDef       *StructDef
	IsOptional         bool
	OptionalFieldIndex int
}

func NewStructFieldDef(
	parentRef *StructDef, name string, lineNumber int) *StructFieldDef {

	newObj := new(StructFieldDef)
	newObj.ParentRef = parentRef
	newObj.Name = name
	newObj.LineNumber = lineNumber

	return newObj
}

func (this *StructFieldDef) Close() {
	this.RefStructDef = nil
	this.RefEnumDef = nil
	this.ParentRef = nil
}

// ----------------------------------------------------------------------------
type StructDef struct {
	// link to parent define
	ParentRef *ProtocolDef
	// struct name
	Name string
	// define in line number
	LineNumber int

	// in file define order
	Fields []*StructFieldDef
	// StructFieldDef.Name -> StructFieldDef
	FieldNameIndex map[string]*StructFieldDef

	OptionalFieldCount int
	OptionalByteCount  int
}

func NewStructDef(
	parentRef *ProtocolDef, name string, lineNumber int) *StructDef {

	newObj := new(StructDef)
	newObj.ParentRef = parentRef
	newObj.Name = name
	newObj.LineNumber = lineNumber
	newObj.Fields = make([]*StructFieldDef, 0)
	newObj.FieldNameIndex = make(map[string]*StructFieldDef)

	return newObj
}

func (this *StructDef) Close() {
	if this.FieldNameIndex != nil {
		clear(this.FieldNameIndex)
		this.FieldNameIndex = nil
	}
	if this.Fields != nil {
		for _, def := range this.Fields {
			def.Close()
		}
		clear(this.Fields)
		this.Fields = nil
	}
	this.ParentRef = nil
}

// ----------------------------------------------------------------------------
type EnumMapItemType int

const (
	EnumMapItemType_None EnumMapItemType = iota
	EnumMapItemType_Default
	EnumMapItemType_Int
	EnumMapItemType_CurrentEnumRef
)

// ----------------------------------------------------------------------------
type EnumMapItemDef struct {
	// link to parent define
	ParentRef *EnumMapDef
	// enum map item name
	Name string
	// define in line number
	LineNumber int

	Type           EnumMapItemType
	IntValue       int
	RefEnumItemDef *EnumMapItemDef
	RefStructDef   *StructDef
}

func NewEnumMapItemDef(
	parentRef *EnumMapDef, name string, lineNumber int) *EnumMapItemDef {

	newObj := new(EnumMapItemDef)
	newObj.ParentRef = parentRef
	newObj.Name = name
	newObj.LineNumber = lineNumber

	return newObj
}

func (this *EnumMapItemDef) Close() {
	this.RefStructDef = nil
	this.RefEnumItemDef = nil
	this.ParentRef = nil
}

// ----------------------------------------------------------------------------
type EnumMapDef struct {
	// link to parent define
	ParentRef *ProtocolDef
	// enum map name
	Name string
	// define in line number
	LineNumber int

	// in file define order
	Items []*EnumMapItemDef
	// EnumMapItemDef.Name -> EnumMapItemDef
	ItemNameIndex map[string]*EnumMapItemDef
	// StructId -> StructDef
	IdToStructIndex map[int]*StructDef
	// StructDef -> StructId
	StructToIdIndex map[*StructDef]int
}

func NewEnumMapDef(
	parentRef *ProtocolDef, name string, lineNumber int) *EnumMapDef {

	newObj := new(EnumMapDef)
	newObj.ParentRef = parentRef
	newObj.Name = name
	newObj.LineNumber = lineNumber
	newObj.Items = make([]*EnumMapItemDef, 0)
	newObj.ItemNameIndex = make(map[string]*EnumMapItemDef)
	newObj.IdToStructIndex = make(map[int]*StructDef)
	newObj.StructToIdIndex = make(map[*StructDef]int)

	return newObj
}

func (this *EnumMapDef) Close() {
	if this.StructToIdIndex != nil {
		clear(this.StructToIdIndex)
		this.StructToIdIndex = nil
	}
	if this.IdToStructIndex != nil {
		clear(this.IdToStructIndex)
		this.IdToStructIndex = nil
	}
	if this.ItemNameIndex != nil {
		clear(this.ItemNameIndex)
		this.ItemNameIndex = nil
	}
	if this.Items != nil {
		for _, def := range this.Items {
			def.Close()
		}
		clear(this.Items)
		this.Items = nil
	}
	this.ParentRef = nil
}
