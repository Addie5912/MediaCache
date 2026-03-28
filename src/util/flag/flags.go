package flagutil

import (
	"flag"
	"fmt"
	"reflect"
)

// Parse 解析命令行参数，通过反射自动注册结构体字段为命令行 flag
// obj 须为指针类型的结构体，字段使用 `flag` 和 `desc` 标签声明参数名和描述
// 返回填充了命令行参数值后的对象
func Parse(obj interface{}) interface{} {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 复制默认值
	defaultObj := reflect.New(val.Type()).Elem()
	defaultObj.Set(val)

	parseStruct(defaultObj, val, "")
	flag.Parse()
	return obj
}

// parseStruct 递归解析结构体字段，注册命令行 flag
func parseStruct(defaultObj, newObj reflect.Value, prefix string) {
	for i := 0; i < newObj.NumField(); i++ {
		field := newObj.Type().Field(i)
		fieldValue := newObj.Field(i)
		defaultFieldValue := defaultObj.Field(i)

		flagName := field.Tag.Get("flag")
		desc := field.Tag.Get("desc")
		fullFlagName := flagName

		// 组合完整 flag 名称（支持嵌套：parent-child-field）
		if flagName == "" {
			fullFlagName = prefix
		} else if prefix == "" {
			fullFlagName = flagName
		} else {
			fullFlagName = prefix + "-" + flagName
		}

		// 递归处理嵌套结构体
		if fieldValue.Kind() == reflect.Struct {
			parseStruct(defaultFieldValue, fieldValue, fullFlagName)
			continue
		}

		// 跳过无 flag tag 的非结构体字段
		if flagName == "" {
			continue
		}

		if !fieldValue.CanAddr() {
			continue
		}

		// 通过 reflect.NewAt 获取字段可寻址指针
		fieldPtr := reflect.NewAt(field.Type, fieldValue.Addr().UnsafePointer()).Interface()

		switch field.Type.Kind() {
		case reflect.String:
			flag.StringVar(fieldPtr.(*string), fullFlagName, defaultFieldValue.String(), desc)
		case reflect.Int:
			flag.IntVar(fieldPtr.(*int), fullFlagName, int(defaultFieldValue.Int()), desc)
		case reflect.Int64:
			flag.Int64Var(fieldPtr.(*int64), fullFlagName, defaultFieldValue.Int(), desc)
		case reflect.Uint:
			flag.UintVar(fieldPtr.(*uint), fullFlagName, uint(defaultFieldValue.Uint()), desc)
		case reflect.Uint64:
			flag.Uint64Var(fieldPtr.(*uint64), fullFlagName, defaultFieldValue.Uint(), desc)
		case reflect.Bool:
			flag.BoolVar(fieldPtr.(*bool), fullFlagName, defaultFieldValue.Bool(), desc)
		case reflect.Float64:
			flag.Float64Var(fieldPtr.(*float64), fullFlagName, defaultFieldValue.Float(), desc)
		default:
			panic(fmt.Sprintf("不支持的类型: %s", field.Type))
		}
	}
}
