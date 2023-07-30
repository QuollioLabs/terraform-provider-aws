// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// Flatten "flattens" an AWS SDK for Go v2 API data structure into
// a resource's "business logic" data structure, implemented using
// Terraform Plugin Framework data types.
// The API data structure's fields are walked and exported fields that
// have a corresponding field in the resource's data structure (and a
// suitable target data type) are copied.
func Flatten(ctx context.Context, apiObject, tfObject any) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(walkStructFields(ctx, apiObject, tfObject, flattenVisitor{})...)
	if diags.HasError() {
		diags.AddError("AutoFlEx", fmt.Sprintf("Flatten[%T, %T]", apiObject, tfObject))
		return diags
	}

	return diags
}

type flattenVisitor struct{}

func (visitor flattenVisitor) visit(ctx context.Context, fieldName string, vFrom, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	valTo, ok := vTo.Interface().(attr.Value)
	if !ok {
		diags.AddError("AutoFlEx", fmt.Sprintf("does not implement attr.Value: %s", vTo.Kind()))
		return diags
	}

	tTo := valTo.Type(ctx)
	switch vFrom.Kind() {
	case reflect.Bool:
		diags.Append(visitor.bool(ctx, vFrom, tTo, vTo)...)
		return diags

	case reflect.Float32, reflect.Float64:
		diags.Append(visitor.float(ctx, vFrom, tTo, vTo)...)
		return diags

	case reflect.Int32, reflect.Int64:
		diags.Append(visitor.int(ctx, vFrom, tTo, vTo)...)
		return diags

	case reflect.String:
		diags.Append(visitor.string(ctx, vFrom, tTo, vTo)...)
		return diags

	case reflect.Ptr:
		diags.Append(visitor.ptr(ctx, vFrom, tTo, vTo)...)
		return diags

	case reflect.Slice:
		diags.Append(visitor.slice(ctx, vFrom, tTo, vTo)...)
		return diags

	case reflect.Map:
		diags.Append(visitor.map_(ctx, vFrom, tTo, vTo)...)
		return diags
	}

	diags.Append(visitor.newIncompatibleTypesError(vFrom, tTo))
	return diags
}

// bool copies an AWS API bool value to a compatible Plugin Framework field.
func (visitor flattenVisitor) bool(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.BoolTypable:
		v, d := tTo.ValueFromBool(ctx, types.BoolValue(vFrom.Bool()))
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		//
		// bool -> types.Bool.
		//
		vTo.Set(reflect.ValueOf(v))
		return diags
	}

	diags.Append(visitor.newIncompatibleTypesError(vFrom, tTo))
	return diags
}

// float copies an AWS API float value to a compatible Plugin Framework field.
func (visitor flattenVisitor) float(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.Float64Typable:
		v, d := tTo.ValueFromFloat64(ctx, types.Float64Value(vFrom.Float()))
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		//
		// float32/float64 -> types.Float64.
		//
		vTo.Set(reflect.ValueOf(v))
		return diags
	}

	diags.Append(visitor.newIncompatibleTypesError(vFrom, tTo))
	return diags
}

// int copies an AWS API int value to a compatible Plugin Framework field.
func (visitor flattenVisitor) int(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.Int64Typable:
		v, d := tTo.ValueFromInt64(ctx, types.Int64Value(vFrom.Int()))
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		//
		// int32/int64 -> types.Int64.
		//
		vTo.Set(reflect.ValueOf(v))
		return diags
	}

	diags.Append(visitor.newIncompatibleTypesError(vFrom, tTo))
	return diags
}

// string copies an AWS API string value to a compatible Plugin Framework field.
func (visitor flattenVisitor) string(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.StringTypable:
		v, d := tTo.ValueFromString(ctx, types.StringValue(vFrom.String()))
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		//
		// string -> types.String.
		//
		vTo.Set(reflect.ValueOf(v))
		return diags
	}

	diags.Append(visitor.newIncompatibleTypesError(vFrom, tTo))
	return diags
}

// ptr copies an AWS API pointer value to a compatible Plugin Framework field.
func (visitor flattenVisitor) ptr(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch vElem := vFrom.Elem(); vFrom.Type().Elem().Kind() {
	case reflect.Bool:
		if vFrom.IsNil() {
			vTo.Set(reflect.ValueOf(types.BoolNull()))
			return diags
		}

		diags.Append(visitor.bool(ctx, vElem, tTo, vTo)...)
		return diags

	case reflect.Float32, reflect.Float64:
		if vFrom.IsNil() {
			vTo.Set(reflect.ValueOf(types.Float64Null()))
			return diags
		}

		diags.Append(visitor.float(ctx, vElem, tTo, vTo)...)
		return diags

	case reflect.Int32, reflect.Int64:
		if vFrom.IsNil() {
			vTo.Set(reflect.ValueOf(types.Int64Null()))
			return diags
		}

		diags.Append(visitor.int(ctx, vElem, tTo, vTo)...)
		return diags

	case reflect.String:
		if vFrom.IsNil() {
			vTo.Set(reflect.ValueOf(types.StringNull()))
			return diags
		}

		diags.Append(visitor.string(ctx, vElem, tTo, vTo)...)
		return diags

	case reflect.Struct:
		if tTo, ok := tTo.(fwtypes.NestedObjectType); ok {
			//
			// *struct -> types.List(OfObject).
			//
			diags.Append(visitor.ptrToStructNestedObject(ctx, vFrom, tTo, vTo)...)
			return diags
		}
	}

	diags.Append(visitor.newIncompatibleTypesError(vFrom, tTo))
	return diags
}

// slice copies an AWS API slice value to a compatible Plugin Framework field.
func (visitor flattenVisitor) slice(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tSliceElem := vFrom.Type().Elem(); tSliceElem.Kind() {
	case reflect.String:
		switch tTo := tTo.(type) {
		case basetypes.ListTypable:
			//
			// []string -> types.List(OfString).
			//
			if vFrom.IsNil() {
				vTo.Set(reflect.ValueOf(types.ListNull(types.StringType)))
				return diags
			}

			v, d := tTo.ValueFromList(ctx, FlattenFrameworkStringValueList(ctx, vFrom.Interface().([]string)))
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(v))
			return diags

		case basetypes.SetTypable:
			//
			// []string -> types.Set(OfString).
			//
			if vFrom.IsNil() {
				vTo.Set(reflect.ValueOf(types.SetNull(types.StringType)))
				return diags
			}

			v, d := tTo.ValueFromSet(ctx, FlattenFrameworkStringValueSet(ctx, vFrom.Interface().([]string)))
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(v))
			return diags
		}

	case reflect.Ptr:
		switch tSliceElem.Elem().Kind() {
		case reflect.String:
			switch tTo := tTo.(type) {
			case basetypes.ListTypable:
				//
				// []*string -> types.List(OfString).
				//
				if vFrom.IsNil() {
					vTo.Set(reflect.ValueOf(types.ListNull(types.StringType)))
					return diags
				}

				v, d := tTo.ValueFromList(ctx, FlattenFrameworkStringList(ctx, vFrom.Interface().([]*string)))
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(v))
				return diags

			case basetypes.SetTypable:
				//
				// []string -> types.Set(OfString).
				//
				if vFrom.IsNil() {
					vTo.Set(reflect.ValueOf(types.SetNull(types.StringType)))
					return diags
				}

				v, d := tTo.ValueFromSet(ctx, FlattenFrameworkStringSet(ctx, vFrom.Interface().([]*string)))
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(v))
				return diags
			}

		case reflect.Struct:
			if tTo, ok := tTo.(fwtypes.NestedObjectType); ok {
				//
				// []*struct -> types.List(OfObject).
				//
				diags.Append(visitor.sliceOfStructNestedObject(ctx, vFrom, tTo, vTo)...)
				return diags
			}
		}

	case reflect.Struct:
		if tTo, ok := tTo.(fwtypes.NestedObjectType); ok {
			//
			// []struct -> types.List(OfObject).
			//
			diags.Append(visitor.sliceOfStructNestedObject(ctx, vFrom, tTo, vTo)...)
			return diags
		}
	}

	diags.Append(visitor.newIncompatibleTypesError(vFrom, tTo))
	return diags
}

// map_ copies an AWS API map value to a compatible Plugin Framework field.
func (visitor flattenVisitor) map_(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tMapKey := vFrom.Type().Key(); tMapKey.Kind() {
	case reflect.String:
		switch tMapElem := vFrom.Type().Elem(); tMapElem.Kind() {
		case reflect.String:
			switch tTo := tTo.(type) {
			case basetypes.MapTypable:
				//
				// map[string]string -> types.Map(OfString).
				//
				if vFrom.IsNil() {
					vTo.Set(reflect.ValueOf(types.MapNull(types.StringType)))
					return diags
				}

				v, d := tTo.ValueFromMap(ctx, FlattenFrameworkStringValueMap(ctx, vFrom.Interface().(map[string]string)))
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(v))
				return diags
			}

		case reflect.Ptr:
			switch tMapElem.Elem().Kind() {
			case reflect.String:
				switch tTo := tTo.(type) {
				case basetypes.MapTypable:
					//
					// map[string]*string -> types.Map(OfString).
					//
					if vFrom.IsNil() {
						vTo.Set(reflect.ValueOf(types.MapNull(types.StringType)))
						return diags
					}

					v, d := tTo.ValueFromMap(ctx, FlattenFrameworkStringMap(ctx, vFrom.Interface().(map[string]*string)))
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(v))
					return diags
				}
			}
		}
	}

	diags.Append(visitor.newIncompatibleTypesError(vFrom, tTo))
	return diags
}

// ptrToStructNestedObject copies an AWS API *struct value to a compatible Plugin Framework NestedObjectValue field.
func (visitor flattenVisitor) ptrToStructNestedObject(ctx context.Context, vFrom reflect.Value, tTo fwtypes.NestedObjectType, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	if vFrom.IsNil() {
		val, d := tTo.NullValue(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(val))
		return diags
	}

	// Create a new target structure and walk its fields.
	to, d := tTo.NewObjectPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	diags.Append(walkStructFields(ctx, vFrom.Elem().Interface(), to, visitor)...)
	if diags.HasError() {
		return diags
	}

	// Set the target structure as a nested Object.
	val, d := tTo.ValueFromObjectPtr(ctx, to)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	vTo.Set(reflect.ValueOf(val))
	return diags
}

// sliceOfStructNestedObject copies an AWS API []struct value to a compatible Plugin Framework NestedObjectValue field.
func (visitor flattenVisitor) sliceOfStructNestedObject(ctx context.Context, vFrom reflect.Value, tTo fwtypes.NestedObjectType, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	if vFrom.IsNil() {
		val, d := tTo.NullValue(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(val))
		return diags
	}

	// Create a new target slice and flatten each element.
	n := vFrom.Len()
	to, d := tTo.NewObjectSlice(ctx, n, n)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	t := reflect.ValueOf(to)
	for i := 0; i < n; i++ {
		target, d := tTo.NewObjectPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		diags.Append(walkStructFields(ctx, vFrom.Index(i).Interface(), target, visitor)...)
		if diags.HasError() {
			return diags
		}

		t.Index(i).Set(reflect.ValueOf(target))
	}

	// Set the target structure as a nested Object.
	val, d := tTo.ValueFromObjectSlice(ctx, to)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	vTo.Set(reflect.ValueOf(val))
	return diags
}

func (visitor flattenVisitor) newIncompatibleTypesError(vFrom reflect.Value, tTo attr.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic("Incompatible types", fmt.Sprintf("%s cannot be flattened to %s", vFrom.Kind(), tTo))
}
