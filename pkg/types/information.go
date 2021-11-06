package types

import (
	"fmt"
)

type Field struct {
	Name  string
	Value interface{}
}

func (f *Field) String() string {
	switch f.Value.(type) {
	case int, uint, int64, int32, int16, int8, uint64, uint32, uint16, uint8:
		return fmt.Sprintf("%d", f.Value)
	case bool:
		return fmt.Sprintf("%v", f.Value)
	default:
		return fmt.Sprintf("%s", f.Value)
	}
}

type Record struct {
	Fields []Field
}

func (f *Record) AddField(name string, value interface{}) {
	f.Fields = append(f.Fields, Field{
		Name:  name,
		Value: value,
	})
}

type Item struct {
	Record
	Header []Field
}

func NewItem() Item {
	return Item{}
}

func (i Item) GetString(s string) string {
	for _, f := range i.Fields {
		if f.Name == s {
			return f.Value.(string)
		}
	}
	return ""
}

func (i Item) GetUint8(s string) uint8 {
	for _, f := range i.Fields {
		if f.Name == s {
			return f.Value.(uint8)
		}
	}
	return 0
}

type Table struct {
	Type   string
	Header []Field
	Rows   []Record
}
