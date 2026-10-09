package main

// Generates frontend/src/api/models.ts from the Go types the editor API sends and receives.
// This is ported from Wails v2.16.0 (internal/binding/binding.go, MIT License, Copyright (c) 2018-Present Lea Anthony),
// which generated these models before the editor moved off Wails, so the output keeps the same shape.

import (
	"bufio"
	"bytes"
	"editor/internal/typescriptify"
	"mud/world"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/leaanthony/slicer"
)

type modelGenerator struct {
	// package name -> type name -> zero value of the type
	structs map[string]map[string]interface{}
	enums map[string]map[string]interface{}
}

func generateModels(path string) error {
	generator := &modelGenerator {
		structs: make(map[string]map[string]interface{}),
		enums: make(map[string]map[string]interface{}),
	}

	generator.addStruct(EditorConstants{})
	generator.addStruct(EditorWorld{})
	generator.addStruct(world.ItemData{})
	generator.addStruct(world.NpcData{})

	generator.addEnum(ALL_DIRECTIONS)
	generator.addEnum(ALL_NPC_DISPOSITIONS)
	generator.addEnum(ALL_NPC_MOVEMENT_TYPES)
	generator.addEnum(ALL_NPC_BEHAVIOR_PARAM_TYPES)
	generator.addEnum(ALL_CHEST_TYPES)

	models, err := generator.generate()
	if err != nil {
		return err
	}

	return os.WriteFile(path, models, 0644)
}

func (generator *modelGenerator) addStruct(value interface{}) {
	structType := reflect.TypeOf(value)
	generator.addStructType(getPackageName(structType.String()), structType.Name(), value)
}

func (generator *modelGenerator) addStructType(packageName string, structName string, value interface{}) {
	if generator.structs[packageName] == nil {
		generator.structs[packageName] = make(map[string]interface{})
	}
	if generator.structs[packageName][structName] != nil {
		return
	}
	generator.structs[packageName][structName] = value

	// Add any structs that this struct's fields reference
	structType := reflect.TypeOf(value)
	for hasElements(structType) {
		structType = structType.Elem()
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.Anonymous || !field.IsExported() {
			continue
		}

		fieldType := field.Type
		for hasElements(fieldType) {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() != reflect.Struct {
			continue
		}

		fullName := fieldType.String()
		nameParts := strings.SplitN(fullName, ".", 2)
		if len(nameParts) < 2 {
			continue
		}
		if hasExportedJsonFields(fieldType) {
			generator.addStructType(getPackageName(fullName), nameParts[1], reflect.Indirect(reflect.New(fieldType)).Interface())
		}
	}
}

// Enums are slices of struct { Value T; TSName string }, see enums.go
func (generator *modelGenerator) addEnum(enum interface{}) {
	valueField, _ := reflect.TypeOf(enum).Elem().FieldByName("Value")
	packageName := getPackageName(valueField.Type.String())
	enumName := valueField.Type.Name()

	if generator.enums[packageName] == nil {
		generator.enums[packageName] = make(map[string]interface{})
	}
	generator.enums[packageName][enumName] = enum
}

func (generator *modelGenerator) generate() ([]byte, error) {
	models := map[string]string{}
	var seen slicer.StringSlicer
	var seenEnumPackages slicer.StringSlicer
	allStructNames := getAllTypeNames(generator.structs)
	allEnumNames := getAllTypeNames(generator.enums)

	for packageName, structs := range generator.structs {
		converter := typescriptify.New()
		converter.Namespace = packageName
		converter.WithBackupDir("")
		converter.KnownStructs = allStructNames
		converter.KnownEnums = allEnumNames

		for _, structName := range sortedKeys(structs) {
			if seen.Contains(packageName + "." + structName) {
				continue
			}
			converter.Add(structs[structName])
		}

		enums, hasEnums := generator.enums[packageName]
		if hasEnums {
			for _, enumName := range sortedKeys(enums) {
				if seen.Contains(packageName + "." + enumName) {
					continue
				}
				converter.AddEnum(enums[enumName])
			}
			seenEnumPackages.Add(packageName)
		}

		code, err := converter.Convert(nil)
		if err != nil {
			return nil, err
		}
		seen.AddSlice(converter.GetGeneratedStructs())
		models[packageName] = code
	}

	// Add enums from packages that have no structs
	for packageName, enums := range generator.enums {
		if seenEnumPackages.Contains(packageName) {
			continue
		}

		converter := typescriptify.New()
		converter.Namespace = packageName
		converter.WithBackupDir("")
		for _, enumName := range sortedKeys(enums) {
			converter.AddEnum(enums[enumName])
		}

		code, err := converter.Convert(nil)
		if err != nil {
			return nil, err
		}
		models[packageName] = code
	}

	var output bytes.Buffer
	for _, packageName := range sortedKeys(models) {
		code := models[packageName]
		if strings.TrimSpace(code) == "" {
			continue
		}

		output.WriteString("export namespace " + packageName + " {\n")
		scanner := bufio.NewScanner(strings.NewReader(code))
		for scanner.Scan() {
			output.WriteString("\t" + scanner.Text() + "\n")
		}
		output.WriteString("\n}\n\n")
	}

	return output.Bytes(), nil
}

func getAllTypeNames(types map[string]map[string]interface{}) *slicer.StringSlicer {
	var result slicer.StringSlicer
	for packageName, packageTypes := range types {
		for typeName := range packageTypes {
			result.Add(packageName + "." + typeName)
		}
	}
	result.Sort()
	return &result
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func getPackageName(typeName string) string {
	result := strings.Split(typeName, ".")[0]
	result = strings.ReplaceAll(result, "[]", "")
	result = strings.ReplaceAll(result, "*", "")
	return result
}

func hasElements(typeOf reflect.Type) bool {
	kind := typeOf.Kind()
	return kind == reflect.Ptr || kind == reflect.Array || kind == reflect.Slice || kind == reflect.Map
}

func hasExportedJsonFields(typeOf reflect.Type) bool {
	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)
		// function, complex, and channel types cannot be json-encoded
		switch field.Type.Kind() {
		case reflect.Chan, reflect.Func, reflect.UnsafePointer, reflect.Complex128, reflect.Complex64:
			continue
		}

		jsonTag, hasTag := field.Tag.Lookup("json")
		if !hasTag && field.IsExported() {
			return true
		}
		jsonFieldName := strings.Split(jsonTag, ",")[0]
		if jsonFieldName != "" && jsonFieldName != "-" {
			return true
		}
	}
	return false
}
