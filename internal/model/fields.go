package model

// FieldConfig carries the properties that belong to one field type. Types
// whose only properties are the common ones carry a nil config, so an empty
// config is never mistaken for a missing one.
type FieldConfig interface {
	fieldType() string
}

type StringConfig struct {
	MultiLine        bool     `json:"multiLine,omitempty"`
	TxtInsertButtons []string `json:"txtInsertButtons,omitempty"`
}

func (StringConfig) fieldType() string { return "string" }

type MultiLineConfig struct {
	Type      string `json:"-"`
	MultiLine bool   `json:"multiLine,omitempty"`
}

func (c MultiLineConfig) fieldType() string { return c.Type }

type NumberConfig struct {
	Min  *float64 `json:"min,omitempty"`
	Max  *float64 `json:"max,omitempty"`
	Step *float64 `json:"step,omitempty"`
}

func (NumberConfig) fieldType() string { return "number" }

type SliderConfig struct {
	Min  float64  `json:"min"`
	Max  float64  `json:"max"`
	Step *float64 `json:"step,omitempty"`
}

func (SliderConfig) fieldType() string { return "slider" }

type DateConfig struct {
	DateFormat string `json:"dateFormat,omitempty"`
	Date       *bool  `json:"date,omitempty"`
	Time       *bool  `json:"time,omitempty"`
}

func (DateConfig) fieldType() string { return "date" }

// Option is one choice of a select or chips field. The reference application
// accepts a bare scalar or an object with a value and a display text.
type Option struct {
	Value any    `json:"value"`
	Text  string `json:"text,omitempty"`
}

type OptionImages struct {
	Path      string `json:"option_image_path,omitempty"`
	Width     *int64 `json:"option_image_width,omitempty"`
	Extension string `json:"option_image_extension,omitempty"`
}

type SelectConfig struct {
	Options  []Option `json:"options"`
	Multiple bool     `json:"multiple,omitempty"`
	OptionImages
}

func (SelectConfig) fieldType() string { return "select" }

type SelectFromQueryConfig struct {
	QueryGlob   string `json:"query_glob"`
	QueryString string `json:"query_string"`
	Multiple    bool   `json:"multiple,omitempty"`
	OptionImages
}

func (SelectFromQueryConfig) fieldType() string { return "select-from-query" }

type ChipsConfig struct {
	Options []Option `json:"options,omitempty"`
}

func (ChipsConfig) fieldType() string { return "chips" }

type FontPickerConfig struct {
	Limit      *int64   `json:"limit,omitempty"`
	Families   []string `json:"families,omitempty"`
	Categories []string `json:"categories,omitempty"`
	Variants   []string `json:"variants,omitempty"`
}

func (FontPickerConfig) fieldType() string { return "font-picker" }

type FontIconPickerConfig struct {
	Multiple bool `json:"multiple,omitempty"`
}

func (FontIconPickerConfig) fieldType() string { return "fonticon-picker" }

type ImageSelectConfig struct {
	Path          string   `json:"path"`
	ButtonTitle   string   `json:"buttonTitle,omitempty"`
	Extensions    []string `json:"extensions,omitempty"`
	ForceFileName string   `json:"forceFileName,omitempty"`
	RealFsPath    string   `json:"real_fs_path,omitempty"`
}

func (ImageSelectConfig) fieldType() string { return "image-select" }

type EisenhouwerConfig struct {
	DataSetsPath                   string   `json:"dataSetsPath,omitempty"`
	DataSetsKeyToLabel             bool     `json:"dataSetsKeyToLabel,omitempty"`
	DataSetsLabelPath              string   `json:"dataSetsLabelPath,omitempty"`
	DataSetsDataPointsPath         string   `json:"dataSetsDataPointsPath,omitempty"`
	DataSetsDataPointsKeyToItem    bool     `json:"dataSetsDataPointsKeyToItem,omitempty"`
	DataSetsDataPointPosXPath      string   `json:"dataSetsDataPointPosXPath,omitempty"`
	DataSetsDataPointPosYPath      string   `json:"dataSetsDataPointPosYPath,omitempty"`
	DataSetsDataPointLabelTemplate string   `json:"dataSetsDataPointLabelTemplate,omitempty"`
	XScaleTitle                    string   `json:"xScaleTitle,omitempty"`
	YScaleTitle                    string   `json:"yScaleTitle,omitempty"`
	PointRadius                    *float64 `json:"pointRadius,omitempty"`
	LabelDoNow                     string   `json:"labelDoNow,omitempty"`
	LabelToPlan                    string   `json:"labelToPlan,omitempty"`
	LabelDelegate                  string   `json:"labelDelegate,omitempty"`
	LabelDelete                    string   `json:"labelDelete,omitempty"`
}

func (EisenhouwerConfig) fieldType() string { return "eisenhouwer" }

type SectionConfig struct {
	Type string `json:"-"`
}

func (c SectionConfig) fieldType() string { return c.Type }

type NestConfig struct {
	MaxDepth *int64 `json:"max_depth,omitempty"`
}

func (NestConfig) fieldType() string { return "nest" }

type AccordionConfig struct {
	DynFormSearchKey    string `json:"dynFormSearchKey,omitempty"`
	DynFormObjectFile   string `json:"dynFormObjectFile,omitempty"`
	DynFormObjectRoot   string `json:"dynFormObjectRoot,omitempty"`
	ArrayIndicesAreKeys bool   `json:"arrayIndicesAreKeys,omitempty"`
	DisableCreate       bool   `json:"disableCreate,omitempty"`
	DisableDelete       bool   `json:"disableDelete,omitempty"`
	DisableSort         bool   `json:"disableSort,omitempty"`
	Lazy                bool   `json:"lazy,omitempty"`
	LazyTemp            bool   `json:"lazyTemp,omitempty"`
}

func (AccordionConfig) fieldType() string { return "accordion" }

type PullConfig struct {
	Group string `json:"group,omitempty"`
}

func (PullConfig) fieldType() string { return "pull" }

// LeafArrayConfig is the odd one: a leaf-array carries one child field rather
// than a list, because every item in the array has the same shape.
type LeafArrayConfig struct {
	Field *Field `json:"field,omitempty"`
}

func (LeafArrayConfig) fieldType() string { return "leaf-array" }

type BundleManagerConfig struct {
	Path                 string   `json:"path"`
	Extensions           []string `json:"extensions,omitempty"`
	ForceFileName        string   `json:"forceFileName,omitempty"`
	MaxItems             *int64   `json:"maxItems,omitempty"`
	AddButtonLocationTop bool     `json:"addButtonLocationTop,omitempty"`
}

func (BundleManagerConfig) fieldType() string { return "bundle-manager" }

type BundleImageThumbnailConfig struct {
	Src string `json:"src,omitempty"`
}

func (BundleImageThumbnailConfig) fieldType() string { return "bundle-image-thumbnail" }

type EmptyLineConfig struct {
	Amount *int64 `json:"amount,omitempty"`
}

func (EmptyLineConfig) fieldType() string { return "empty-line" }

type InfoConfig struct {
	Content    string `json:"content"`
	Size       string `json:"size,omitempty"`
	LineHeight string `json:"lineHeight,omitempty"`
}

func (InfoConfig) fieldType() string { return "info" }

// UnknownConfig keeps a field whose type this version does not know, so a
// model written for a newer Quiqr still opens.
type UnknownConfig struct {
	DeclaredType string         `json:"declaredType"`
	Properties   map[string]any `json:"properties,omitempty"`
}

func (UnknownConfig) fieldType() string { return TypeUnknown }

const TypeUnknown = "unknown"

// KnownFieldTypes is every field type the Quiqr desktop app registers. See
// docs/QUIQR-MODEL.md.
var KnownFieldTypes = []string{
	"string", "markdown", "easymde", "readonly", "uniq",
	"number", "slider", "boolean", "date",
	"select", "select-from-query", "chips", "color",
	"font-picker", "fonticon-picker", "image-select", "eisenhouwer",
	"section", "nest", "accordion", "pull", "leaf-array", "bundle-manager",
	"hidden", "empty-line", "info", "bundle-image-thumbnail",
}

// ContainerFieldTypes are the types whose children live in a fields list.
var ContainerFieldTypes = []string{"section", "nest", "accordion", "pull", "bundle-manager"}

func IsKnownFieldType(name string) bool {
	for _, known := range KnownFieldTypes {
		if known == name {
			return true
		}
	}
	return false
}

func isContainer(name string) bool {
	for _, container := range ContainerFieldTypes {
		if container == name {
			return true
		}
	}
	return false
}
