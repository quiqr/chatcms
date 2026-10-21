package model

import "fmt"

func decodeFields(file, where string, list []any) ([]Field, error) {
	if list == nil {
		return nil, nil
	}

	fields := make([]Field, 0, len(list))
	for index, entry := range list {
		path := fmt.Sprintf("%s[%d]", where, index)

		tree, ok := entry.(map[string]any)
		if !ok {
			return nil, at(file, path, "a mapping")
		}

		field, err := decodeField(file, path, tree)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func decodeField(file, where string, tree map[string]any) (Field, error) {
	r := newReader(file, where, tree)

	field := Field{
		Key:                  r.requiredStr("key"),
		Type:                 r.requiredStr("type"),
		Title:                r.str("title"),
		Tip:                  r.str("tip"),
		Name:                 r.str("name"),
		Group:                r.str("group"),
		Content:              r.str("content"),
		Theme:                r.str("theme"),
		Default:              r.raw("default"),
		Hidden:               r.boolean("hidden"),
		GroupData:            r.boolean("groupdata"),
		ArrayTitle:           r.boolean("arrayTitle"),
		AutoSave:             r.boolean("autoSave"),
		FieldPromptTemplates: r.strList("field_prompt_templates"),
	}
	if r.err != nil {
		return Field{}, r.err
	}

	if isContainer(field.Type) {
		children, err := decodeFields(file, join(where, "fields"), r.list("fields"))
		if err != nil {
			return Field{}, err
		}
		field.Fields = children
	}

	config, err := decodeFieldConfig(file, where, field.Type, r)
	if err != nil {
		return Field{}, err
	}
	field.Config = config

	if r.err != nil {
		return Field{}, r.err
	}

	if !IsKnownFieldType(field.Type) {
		field.Config = UnknownConfig{DeclaredType: field.Type, Properties: r.extra()}
		field.Type = TypeUnknown
		field.Extra = nil
		return field, nil
	}

	field.Extra = r.extra()
	return field, nil
}

func decodeFieldConfig(file, where, fieldType string, r *reader) (FieldConfig, error) {
	switch fieldType {
	case "string":
		return StringConfig{
			MultiLine:        r.boolean("multiLine"),
			TxtInsertButtons: r.strList("txtInsertButtons"),
		}, nil

	case "easymde", "readonly", "uniq":
		return MultiLineConfig{Type: fieldType, MultiLine: r.boolean("multiLine")}, nil

	case "number":
		return NumberConfig{Min: r.number("min"), Max: r.number("max"), Step: r.number("step")}, nil

	case "slider":
		return SliderConfig{
			Min:  r.requiredNumber("min"),
			Max:  r.requiredNumber("max"),
			Step: r.number("step"),
		}, nil

	case "date":
		return DateConfig{
			DateFormat: r.str("dateFormat"),
			Date:       r.booleanPtr("date"),
			Time:       r.booleanPtr("time"),
		}, nil

	case "select":
		options, err := decodeOptions(file, where, r, true)
		if err != nil {
			return nil, err
		}
		return SelectConfig{
			Options:      options,
			Multiple:     r.boolean("multiple"),
			OptionImages: decodeOptionImages(r),
		}, nil

	case "select-from-query":
		return SelectFromQueryConfig{
			QueryGlob:    r.requiredStr("query_glob"),
			QueryString:  r.requiredStr("query_string"),
			Multiple:     r.boolean("multiple"),
			OptionImages: decodeOptionImages(r),
		}, nil

	case "chips":
		options, err := decodeOptions(file, where, r, false)
		if err != nil {
			return nil, err
		}
		return ChipsConfig{Options: options}, nil

	case "font-picker":
		return FontPickerConfig{
			Limit:      r.integer("limit"),
			Families:   r.strList("families"),
			Categories: r.strList("categories"),
			Variants:   r.strList("variants"),
		}, nil

	case "fonticon-picker":
		return FontIconPickerConfig{Multiple: r.boolean("multiple")}, nil

	case "image-select":
		return ImageSelectConfig{
			Path:          r.requiredStr("path"),
			ButtonTitle:   r.str("buttonTitle"),
			Extensions:    r.strList("extensions"),
			ForceFileName: r.str("forceFileName"),
			RealFsPath:    r.str("real_fs_path"),
		}, nil

	case "eisenhouwer":
		return EisenhouwerConfig{
			DataSetsPath:                   r.str("dataSetsPath"),
			DataSetsKeyToLabel:             r.boolean("dataSetsKeyToLabel"),
			DataSetsLabelPath:              r.str("dataSetsLabelPath"),
			DataSetsDataPointsPath:         r.str("dataSetsDataPointsPath"),
			DataSetsDataPointsKeyToItem:    r.boolean("dataSetsDataPointsKeyToItem"),
			DataSetsDataPointPosXPath:      r.str("dataSetsDataPointPosXPath"),
			DataSetsDataPointPosYPath:      r.str("dataSetsDataPointPosYPath"),
			DataSetsDataPointLabelTemplate: r.str("dataSetsDataPointLabelTemplate"),
			XScaleTitle:                    r.str("xScaleTitle"),
			YScaleTitle:                    r.str("yScaleTitle"),
			PointRadius:                    r.number("pointRadius"),
			LabelDoNow:                     r.str("labelDoNow"),
			LabelToPlan:                    r.str("labelToPlan"),
			LabelDelegate:                  r.str("labelDelegate"),
			LabelDelete:                    r.str("labelDelete"),
		}, nil

	case "section", "pull":
		if fieldType == "pull" {
			return PullConfig{Group: r.str("group")}, nil
		}
		return SectionConfig{Type: fieldType}, nil

	case "nest":
		return NestConfig{MaxDepth: r.integer("max_depth")}, nil

	case "accordion":
		return AccordionConfig{
			DynFormSearchKey:    r.str("dynFormSearchKey"),
			DynFormObjectFile:   r.str("dynFormObjectFile"),
			DynFormObjectRoot:   r.str("dynFormObjectRoot"),
			ArrayIndicesAreKeys: r.boolean("arrayIndicesAreKeys"),
			DisableCreate:       r.boolean("disableCreate"),
			DisableDelete:       r.boolean("disableDelete"),
			DisableSort:         r.boolean("disableSort"),
			Lazy:                r.boolean("lazy"),
			LazyTemp:            r.boolean("lazyTemp"),
		}, nil

	case "leaf-array":
		child := r.mapping("field")
		if child == nil {
			if r.err == nil {
				r.fail("field", "the single field every item of this array has")
			}
			return nil, r.err
		}
		decoded, err := decodeField(file, join(where, "field"), child)
		if err != nil {
			return nil, err
		}
		return LeafArrayConfig{Field: &decoded}, nil

	case "bundle-manager":
		return BundleManagerConfig{
			Path:                 r.requiredStr("path"),
			Extensions:           r.strList("extensions"),
			ForceFileName:        r.str("forceFileName"),
			MaxItems:             r.integer("maxItems"),
			AddButtonLocationTop: r.boolean("addButtonLocationTop"),
		}, nil

	case "bundle-image-thumbnail":
		return BundleImageThumbnailConfig{Src: r.str("src")}, nil

	case "empty-line":
		return EmptyLineConfig{Amount: r.integer("amount")}, nil

	case "info":
		return InfoConfig{
			Content:    r.requiredStr("content"),
			Size:       r.str("size"),
			LineHeight: r.str("lineHeight"),
		}, nil

	case "markdown", "boolean", "color", "hidden":
		return nil, nil

	default:
		return nil, nil
	}
}

func decodeOptionImages(r *reader) OptionImages {
	return OptionImages{
		Path:      r.str("option_image_path"),
		Width:     r.integer("option_image_width"),
		Extension: r.str("option_image_extension"),
	}
}

func decodeOptions(file, where string, r *reader, required bool) ([]Option, error) {
	list := r.list("options")
	if list == nil {
		if required && r.err == nil {
			return nil, at(file, join(where, "options"), "a list of options")
		}
		return nil, nil
	}

	options := make([]Option, 0, len(list))
	for index, entry := range list {
		path := fmt.Sprintf("%s.options[%d]", where, index)

		switch typed := entry.(type) {
		case string, int64, float64, bool:
			options = append(options, Option{Value: typed})
		case map[string]any:
			optionReader := newReader(file, path, typed)
			option := Option{Value: optionReader.raw("value"), Text: optionReader.str("text")}
			if optionReader.err != nil {
				return nil, optionReader.err
			}
			if option.Value == nil {
				return nil, at(file, join(path, "value"), "the value this option writes")
			}
			options = append(options, option)
		default:
			return nil, at(file, path, "a value, or a mapping with a value and a text")
		}
	}
	return options, nil
}
