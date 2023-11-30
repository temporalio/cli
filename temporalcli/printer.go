package temporalcli

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

type PrintOptions struct {
	// If not set, the value type must be a struct with exported fields which will
	// be used in the default in order of appearance.
	Fields []string
	// If not set, not printed as table in text mode. This is ignored for JSON
	// printers.
	Table *PrintTableOptions
}

type PrintTableOptions struct {
	NoHeader  bool
	Separator string
}

type Printer interface {
	// See [JsonPrinter.Print] and [TextPrinter.Print] for details on how each
	// printer prints.
	Print(PrintOptions, any) error
}

type JSONPrinter struct{ enc *json.Encoder }

type JSONPrinterOptions struct {
	// Required, will panic if not present
	Output io.Writer
}

func NewJSONPrinter(options JSONPrinterOptions) *JSONPrinter {
	if options.Output == nil {
		panic("missing output")
	}
	j := &JSONPrinter{json.NewEncoder(options.Output)}
	// TODO(cretz): Customizable?
	j.enc.SetIndent("", "  ")
	return j
}

// Print simply JSON encodes the given value.
func (j *JSONPrinter) Print(opts PrintOptions, v any) error {
	return j.enc.Encode(v)
}

type TextPrinter struct{ options TextPrinterOptions }

type TextPrinterOptions struct {
	// Required, will panic if not present
	Output io.Writer
	// Defaults to RFC3339
	FormatTime func(time.Time) string
}

func NewTextPrinter(options TextPrinterOptions) *TextPrinter {
	if options.Output == nil {
		panic("missing output")
	}
	if options.FormatTime == nil {
		options.FormatTime = func(t time.Time) string { return t.Format(time.RFC3339) }
	}
	return &TextPrinter{options}
}

// Print will print the given value. If the value is a slice, it is treated as
// if each value was given separately.
//
// Every individual must should be a struct, a pointer to a struct, or a map.
func (t *TextPrinter) Print(opts PrintOptions, v any) error {
	// Collect fields and data
	fields, data, err := allData(opts, v)
	if err != nil {
		return err
	}
	// Print table or card
	if opts.Table != nil {
		return t.printTable(*opts.Table, fields, data)
	}
	return t.printCard(fields, data)
}

func (t *TextPrinter) printCard(fields []string, data []map[string]any) error {
	for _, item := range data {
		rows := make([]map[string]any, len(fields))
		for i, field := range fields {
			rows[i] = map[string]any{"Name": field, "Value": item[field]}
		}
		if err := t.printTable(PrintTableOptions{NoHeader: true}, []string{"Name", "Value"}, rows); err != nil {
			return err
		}
		// Newline between cards
		if _, err := t.options.Output.Write([]byte("\n")); err != nil {
			return err
		}
	}
	return nil
}

func (t *TextPrinter) printTable(
	opts PrintTableOptions,
	fields []string,
	data []map[string]any,
) error {
	table := tablewriter.NewWriter(t.options.Output)
	table.SetBorder(false)
	table.SetColumnSeparator(opts.Separator)

	if !opts.NoHeader {
		table.SetHeader(fields)
		table.SetAutoFormatHeaders(false)

		if !color.NoColor {
			headerColors := make([]tablewriter.Colors, len(fields))
			for i := range headerColors {
				// TODO(cretz): Configurable header color
				headerColors[i] = tablewriter.Colors{tablewriter.FgHiMagentaColor}
			}
			table.SetHeaderColor(headerColors...)
		}
		table.SetHeaderLine(false)
	}

	for _, item := range data {
		cols := make([]string, len(fields))
		for i, field := range fields {
			cols[i] = t.textVal(item[field])
		}
		table.Append(cols)
	}
	table.Render()
	table.ClearRows()
	return nil
}

func (t *TextPrinter) textVal(v any) string {
	ref := reflect.Indirect(reflect.ValueOf(v))
	if ref.IsValid() && !ref.IsZero() && ref.Type() == reflect.TypeOf(time.Time{}) {
		return t.options.FormatTime(ref.Interface().(time.Time))
	} else if ref.Kind() == reflect.Struct && ref.CanInterface() {
		b, _ := json.Marshal(v)
		return string(b)
	}
	return fmt.Sprintf("%v", v)
}

func allData(opts PrintOptions, v any) (fields []string, data []map[string]any, err error) {
	singleItemType := reflect.TypeOf(v)
	if singleItemType.Kind() == reflect.Slice {
		singleItemType = singleItemType.Elem()
	} else {
		sliceVal := reflect.MakeSlice(reflect.SliceOf(singleItemType), 1, 1)
		sliceVal.Index(0).Set(reflect.ValueOf(v))
		v = sliceVal.Interface()
	}

	// Validate and create field getter
	fields = opts.Fields
	var fieldGetter func(field string, v reflect.Value) any
	switch singleItemType.Kind() {
	case reflect.Map:
		if len(fields) == 0 {
			return nil, nil, fmt.Errorf("must have fields if using map")
		}
		fieldGetter = func(field string, v reflect.Value) any {
			return v.MapIndex(reflect.ValueOf(field)).Interface()
		}
	case reflect.Struct:
		if len(fields) == 0 {
			fields = exportedFields(singleItemType)
		}
		fieldGetter = func(field string, v reflect.Value) any {
			return v.FieldByName(field).Interface()
		}
	case reflect.Pointer:
		if singleItemType.Elem().Kind() != reflect.Struct {
			return nil, nil, fmt.Errorf("expected map, struct, or pointer to struct, got: %v", singleItemType)
		}
		if len(fields) == 0 {
			fields = exportedFields(singleItemType.Elem())
		}
		fieldGetter = func(field string, v reflect.Value) any {
			return v.Elem().FieldByName(field).Interface()
		}
	default:
		return nil, nil, fmt.Errorf("expected map, struct, or pointer to struct, got: %v", singleItemType)
	}

	// Build data
	sliceVal := reflect.ValueOf(v)
	data = make([]map[string]any, sliceVal.Len())
	for i := range data {
		itemVal := sliceVal.Index(i)
		itemData := make(map[string]any, len(fields))
		for _, f := range fields {
			itemData[f] = fieldGetter(f, itemVal)
		}
		data[i] = itemData
	}
	return
}

func exportedFields(t reflect.Type) []string {
	ret := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		if f := t.Field(i); f.IsExported() {
			ret = append(ret, f.Name)
		}
	}
	return ret
}
