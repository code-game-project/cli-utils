package casing

import (
	"strings"

	"github.com/iancoleman/strcase"
)

func ToOneWord(s string) string {
	return strings.ReplaceAll(strcase.ToSnake(s), "_", "")
}

func ToSnake(s string) string {
	return strcase.ToSnake(s)
}

func ToScreamingSnake(s string) string {
	return strcase.ToScreamingSnake(s)
}

func ToKebab(s string) string {
	return strcase.ToKebab(s)
}

func ToScreamingKebab(s string) string {
	return strcase.ToScreamingKebab(s)
}

func ToCamel(s string) string {
	return strcase.ToLowerCamel(s)
}

func ToPascal(s string) string {
	return strcase.ToCamel(s)
}
