package main

import (
	"fmt"
	"strings"
	"strconv"
	"time"
	"text/template"
)

var (
	TemplateFilters template.FuncMap
)

func checkTemplate() bool {
	blnFoundError := false
	for k, v := range SQLImportConf.AssetGenericFieldMapping {
		str := fmt.Sprintf("%v", v)
		t := template.New(str).Funcs(TemplateFilters)
		_, err := t.Parse(str)
		if err != nil {
			fmt.Println("[TEMPLATE] Parsing Error: " + err.Error() + " [" + k + "]")
			blnFoundError = true
		}
	}
	for k, v := range SQLImportConf.AssetTypeFieldMapping {
		str := fmt.Sprintf("%v", v)
		t := template.New(str).Funcs(TemplateFilters)
		_, err := t.Parse(str)
		if err != nil {
			fmt.Println("[TEMPLATE] Parsing Error: " + err.Error() + " [" + k + "]")
			blnFoundError = true
		}
	}
	for _, assetType := range SQLImportConf.AssetTypes {
		for k, v := range assetType.SoftwareInventory.Mapping {
			str := fmt.Sprintf("%v", v)
			t := template.New(str).Funcs(TemplateFilters)
			_, err := t.Parse(str)
			if err != nil {
				fmt.Println("[TEMPLATE] Parsing Error: " + err.Error() + " [" + assetType.AssetType + "." + k + "]")
				blnFoundError = true
			}
		}
	}
	return blnFoundError
}

func setTemplateFilters() {
	TemplateFilters = template.FuncMap{
		"Upper": func(feature string) string {
			return strings.ToUpper(feature)
		},
		"Lower": func(feature string) string {
			return strings.ToLower(feature)
		},
		"epoch": func(feature string) string {
			result := ""
			if feature == "" {
			} else if feature == "0" {
			} else {
				t, err := strconv.ParseInt(feature,10, 0)
				if err == nil {
					md := time.Unix(t, 0)
					result = md.Format("2006-01-02 15:04:05")
				}
			}
			return result
		},
		"epoch_clear": func(feature string) string {
			result := "__clear__"
			if feature == "" {
			} else if feature == "0" {
			} else {
				t, err := strconv.ParseInt(feature,10, 0)
				if err == nil {
					md := time.Unix(t, 0)
					result = md.Format("2006-01-02 15:04:05")
				}
			}
			return result
		},
	}
}
