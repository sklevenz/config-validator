package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("ocv", "A validation tool for operation configuration")

	showCmd        = app.Command("show", "Show schema definition")
	showSchemaFile = showCmd.Arg("schema", "Absolute file name to a schema yaml file").Required().ExistingFile()

	validateCmd           = app.Command("validate", "Validate schema definition")
	validationSchemaFile  = validateCmd.Arg("schema", "Absolute file name to a schema yaml file").Required().ExistingFile()
	validationConfigFiles = validateCmd.Arg("config", "Absolute file names to a configuration yaml files").Required().ExistingFiles()
)

type anonymousMap map[interface{}]interface{}
type anonymousStringMap map[string]interface{}

type providedType string
type statusType string
type annotationType string
type annotationsType []annotationType

// PropertyType ...
type PropertyType struct {
	Name        string `yaml:"property"`
	Annotations annotationsType
	Default     interface{}
	Description string
}

// SchemaType ...
type SchemaType struct {
	Properties []PropertyType
}

const (
	provided providedType = "provided"
	missing  providedType = "missing"
	obsolete providedType = "obsolete"

	valid statusType = "valid"
	flaw  statusType = "flaw"

	required   annotationType = "required"
	optional   annotationType = "optional"
	deprecated annotationType = "deprecated"
)

// ValidationResultType ...
type ValidationResultType struct {
	PropertyName string
	Annotations  annotationsType
	Provided     providedType
	Status       statusType
	DefaultValue interface{}
	ActualValue  interface{}
}

type showCommandStruct struct{}
type validationCommandStruct struct{}

func (n *showCommandStruct) run(c *kingpin.ParseContext) error {
	data := readFile(*showSchemaFile)
	schema := unmarshalSchema(&data)

	renderPropertiesSchema(schema.Properties, os.Stdout)
	fmt.Print("\n\n")

	return nil
}

func readFile(path string) []byte {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	return data
}

func unmarshalSchema(data *[]byte) *SchemaType {

	var tmp struct {
		TmpSchema SchemaType `yaml:"schema"`
	}
	err := yaml.Unmarshal([]byte(*data), &tmp)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	schema := tmp.TmpSchema
	return &schema
}

func renderPropertiesSchema(properties []PropertyType, writer io.Writer) {
	table := tablewriter.NewWriter(writer)
	table.SetColWidth(80)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetHeader([]string{"Property", "Default", "Annotations", "Description"})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	for _, entry := range properties {
		table.Append([]string{entry.Name, fmt.Sprint(entry.Default), fmt.Sprint(entry.Annotations), entry.Description})
	}
	table.Render()
}

func renderValidationResult(validationResults []ValidationResultType, writer io.Writer) {
	table := tablewriter.NewWriter(writer)
	table.SetColWidth(80)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetHeader([]string{"Property", "Required", "Provided", "Vaildation State", "Default Value", "Actual Value"})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	for _, result := range validationResults {
		table.Append([]string{result.PropertyName, fmt.Sprint(result.Annotations), string(result.Provided), string(result.Status), fmt.Sprint(result.DefaultValue), fmt.Sprint(result.ActualValue)})
	}
	table.Render()
}

func readConfig() []anonymousMap {
	var config []anonymousMap

	for _, file := range *validationConfigFiles {
		data := readFile(file)

		var result anonymousMap
		err := yaml.Unmarshal([]byte(data), &result)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		config = append(config, result)
	}
	return config
}

func (annotations annotationsType) hasAnnotation(annotation annotationType) bool {
	for _, a := range annotations {
		if a == annotation {
			return true
		}
	}
	return false
}

func validateRequiredProperties(schema *SchemaType, config *[]anonymousMap, result *[]ValidationResultType) {
	for _, property := range schema.Properties {
		found := false
		validationResult := ValidationResultType{}
		validationResult.PropertyName = property.Name

		validationResult.Annotations = property.Annotations
		validationResult.DefaultValue = property.Default

		for _, configMap := range *config {
			val, ok := configMap.containsKey(property.Name)

			if ok {
				found = true
				validationResult.ActualValue = val
				validationResult.Provided = provided
				validationResult.Status = valid

				*result = append(*result, validationResult)
				break
			}
		}

		if !found {
			validationResult.ActualValue = nil
			validationResult.Provided = missing

			if property.Annotations.hasAnnotation(required) {
				validationResult.Status = flaw
			} else {
				validationResult.Status = valid
			}
			*result = append(*result, validationResult)
		}
	}
}

func validateObsoleteProperties(schema *SchemaType, config *[]anonymousMap, result *[]ValidationResultType) {
	allConfiguredProperties := make(anonymousStringMap)
	for _, configMap := range *config {
		tmp := configMap.walk("")
		for k, v := range tmp {
			allConfiguredProperties[k] = v
		}
	}
	for k, v := range allConfiguredProperties {
		var found = false
		for _, p := range schema.Properties {
			if k == p.Name {
				found = true
				break
			}
		}
		if !found {
			validationResult := ValidationResultType{}
			validationResult.PropertyName = k
			validationResult.Provided = obsolete
			validationResult.DefaultValue = nil
			validationResult.ActualValue = v
			validationResult.Status = flaw
			*result = append(*result, validationResult)

		}
	}
}

func validate(schema *SchemaType, config *[]anonymousMap) []ValidationResultType {
	var result []ValidationResultType

	validateRequiredProperties(schema, config, &result)
	validateObsoleteProperties(schema, config, &result)

	return result
}

func (m anonymousMap) walk(path string) anonymousStringMap {
	result := make(anonymousStringMap)
	for k, v := range m {
		var newPath string
		if path == "" {
			newPath = k.(string)
		} else {
			newPath = path + "." + k.(string)
		}
		if v != nil && reflect.TypeOf(v).Kind() == reflect.Map {
			tmpResult := v.(anonymousMap).walk(newPath)
			for x, y := range tmpResult {
				if y == nil || reflect.TypeOf(y).Kind() != reflect.Map {
					result[x] = y
				}
			}
		}
		if v == nil || reflect.TypeOf(v).Kind() != reflect.Map {
			result[newPath] = v
		}
	}
	return result
}

func (v *validationCommandStruct) run(c *kingpin.ParseContext) error {
	data := readFile(*validationSchemaFile)

	schema := unmarshalSchema(&data)
	config := readConfig()
	validationResults := validate(schema, &config)

	renderValidationResult(validationResults, os.Stdout)

	return nil
}

func (m anonymousMap) containsKey(key string) (interface{}, bool) {
	keys := strings.SplitN(key, ".", 2)
	if len(keys) > 1 {
		if val, ok := m[keys[0]]; ok {
			return val.(anonymousMap).containsKey(keys[1])
		}
		return nil, false
	}
	if val, ok := m[keys[0]]; ok {
		return val, true
	}

	return nil, false
}

func main() {
	app.Version("0.1")

	showCommand := &showCommandStruct{}
	showCmd.Action(showCommand.run)

	validateCommand := &validationCommandStruct{}
	validateCmd.Action(validateCommand.run)

	_, err := app.Parse(os.Args[1:])

	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
