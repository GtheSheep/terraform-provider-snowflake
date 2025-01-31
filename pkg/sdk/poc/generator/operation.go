package generator

type OperationKind string

const (
	OperationKindCreate   OperationKind = "Create"
	OperationKindAlter    OperationKind = "Alter"
	OperationKindDrop     OperationKind = "Drop"
	OperationKindShow     OperationKind = "Show"
	OperationKindShowByID OperationKind = "ShowByID"
	OperationKindDescribe OperationKind = "Describe"
)

type DescriptionMappingKind string

const (
	DescriptionMappingKindSingleValue DescriptionMappingKind = "single_value"
	DescriptionMappingKindSlice       DescriptionMappingKind = "slice"
)

// Operation defines a single operation for given object or objects family (e.g. CREATE DATABASE ROLE)
type Operation struct {
	// Name is the operation's name, e.g. "Create"
	Name OperationKind
	// ObjectInterface points to the containing interface
	ObjectInterface *Interface
	// Doc is the URL for the doc used to create given operation, e.g. https://docs.snowflake.com/en/sql-reference/sql/create-database-role
	Doc string
	// OptsField defines opts used to create SQL for given operation
	OptsField *Field
	// HelperStructs are struct definitions that are not tied to OptsField, but tied to the Operation itself, e.g. Show() return type
	HelperStructs []*Field
	// ShowMapping is a definition of mapping needed by Operation kind of OperationKindShow
	ShowMapping *Mapping
	// DescribeKind defines a kind of mapping that needs to be performed in particular case of Describe implementation
	DescribeKind *DescriptionMappingKind
	// DescribeMapping is a definition of mapping needed by Operation kind of OperationKindDescribe
	DescribeMapping *Mapping
}

type Mapping struct {
	MappingFuncName string
	From            *Field
	To              *Field
}

func newOperation(kind OperationKind, doc string) *Operation {
	return &Operation{
		Name:          kind,
		Doc:           doc,
		HelperStructs: make([]*Field, 0),
	}
}

func newMapping(mappingFuncName string, from, to *Field) *Mapping {
	return &Mapping{
		MappingFuncName: mappingFuncName,
		From:            from,
		To:              to,
	}
}

func (s *Operation) withOptionsStruct(optsField *Field) *Operation {
	s.OptsField = optsField
	return s
}

func (s *Operation) withHelperStruct(helperStruct *Field) *Operation {
	s.HelperStructs = append(s.HelperStructs, helperStruct)
	return s
}

func (s *Operation) withHelperStructs(helperStructs ...*Field) *Operation {
	s.HelperStructs = append(s.HelperStructs, helperStructs...)
	return s
}

func addShowMapping(op *Operation, from, to *Field) {
	op.ShowMapping = newMapping("convert", from, to)
}

func addDescriptionMapping(op *Operation, from, to *Field) {
	op.DescribeMapping = newMapping("convert", from, to)
}

func (i *Interface) newSimpleOperation(kind OperationKind, doc string, queryStruct *queryStruct, helperStructs ...IntoField) *Interface {
	if queryStruct.identifierField != nil {
		queryStruct.identifierField.Kind = i.IdentifierKind
	}
	f := make([]*Field, len(helperStructs))
	if len(f) > 0 {
		for i, hs := range helperStructs {
			f[i] = hs.IntoField()
		}
	}
	operation := newOperation(kind, doc).
		withOptionsStruct(queryStruct.IntoField()).
		withHelperStructs(f...)
	i.Operations = append(i.Operations, operation)
	return i
}

func (i *Interface) newOperationWithDBMapping(
	kind OperationKind,
	doc string,
	dbRepresentation *dbStruct,
	resourceRepresentation *plainStruct,
	queryStruct *queryStruct,
	addMappingFunc func(op *Operation, from, to *Field),
) *Operation {
	db := dbRepresentation.IntoField()
	res := resourceRepresentation.IntoField()
	if queryStruct.identifierField != nil {
		queryStruct.identifierField.Kind = i.IdentifierKind
	}
	op := newOperation(kind, doc).
		withHelperStruct(db).
		withHelperStruct(res).
		withOptionsStruct(queryStruct.IntoField())
	addMappingFunc(op, db, res)
	i.Operations = append(i.Operations, op)
	return op
}

type IntoField interface {
	IntoField() *Field
}

func (i *Interface) CreateOperation(doc string, queryStruct *queryStruct, helperStructs ...IntoField) *Interface {
	return i.newSimpleOperation(OperationKindCreate, doc, queryStruct, helperStructs...)
}

func (i *Interface) AlterOperation(doc string, queryStruct *queryStruct) *Interface {
	return i.newSimpleOperation(OperationKindAlter, doc, queryStruct)
}

func (i *Interface) DropOperation(doc string, queryStruct *queryStruct) *Interface {
	return i.newSimpleOperation(OperationKindDrop, doc, queryStruct)
}

func (i *Interface) ShowOperation(doc string, dbRepresentation *dbStruct, resourceRepresentation *plainStruct, queryStruct *queryStruct) *Interface {
	i.newOperationWithDBMapping(OperationKindShow, doc, dbRepresentation, resourceRepresentation, queryStruct, addShowMapping)
	return i
}

func (i *Interface) DescribeOperation(describeKind DescriptionMappingKind, doc string, dbRepresentation *dbStruct, resourceRepresentation *plainStruct, queryStruct *queryStruct) *Interface {
	op := i.newOperationWithDBMapping(OperationKindDescribe, doc, dbRepresentation, resourceRepresentation, queryStruct, addDescriptionMapping)
	op.DescribeKind = &describeKind
	return i
}
