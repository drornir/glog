package glog

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strings"
)

func SlogValues(v any) slog.Value {
	rv := reflect.ValueOf(v)
	if rv.Type() == nil {
		return slog.AnyValue(v)
	}
	switch rv.Type().Kind() {
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return slog.AnyValue(v)
	case reflect.Pointer:
		if v == nil {
			return slog.AnyValue(v)
		}
		return SlogValues(rv.Elem().Interface())
	case reflect.Map:
		attrs := make([]slog.Attr, 0, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			attrs = append(attrs, slog.Attr{
				Key:   iter.Key().String(),
				Value: SlogValues(iter.Value().Interface()),
			})
		}
		slices.SortFunc(attrs, func(a, b slog.Attr) int {
			return strings.Compare(a.Key, b.Key)
		})
		return slog.GroupValue(attrs...)
	case reflect.Struct:
		j, err := json.Marshal(v)
		if err != nil {
			return slog.AnyValue(v)
		}
		var m map[string]any
		err = json.Unmarshal(j, &m)
		if err != nil {
			return slog.AnyValue(v)
		}
		return SlogValues(m)
	case reflect.Array, reflect.Slice:
		attrs := make([]slog.Attr, 0, rv.Len())
		for i := range rv.Len() {
			attrs = append(attrs, slog.Attr{
				Key:   fmt.Sprintf("%d", i),
				Value: SlogValues(rv.Index(i).Interface()),
			})
		}
		return slog.GroupValue(attrs...)
	case reflect.Interface:
		return SlogValues(rv.Elem().Interface())
	default:
		return slog.AnyValue(v)
	}
}
