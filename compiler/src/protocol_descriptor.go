package main

type ProtocolDescriptor struct {
	ProtoDef *ProtocolDef

	// ProtocolDef.Name -> ProtocolDef
	ImportedProtos map[string]*ProtocolDef
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

type EnumItemType int

const (
	EnumItemType_None EnumItemType = iota
	EnumItemType_Default
	EnumItemType_Int
	EnumItemType_CurrentEnumRef
	EnumItemType_OtherEnumRef
)

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

type EnumMapItemType int

const (
	EnumMapItemType_None EnumMapItemType = iota
	EnumMapItemType_Default
	EnumMapItemType_Int
	EnumMapItemType_CurrentEnumRef
)

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
