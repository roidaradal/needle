package needle

import (
	"strings"

	"github.com/roidaradal/fn/lang"
	"github.com/roidaradal/fn/str"
)

// Classify function as (public/private) (function/method)
func classifyFunction(line string) CodeType {
	parts := str.SpaceSplit(line)
	isMethod := parts[1][0] == '(' // check if has receiver
	isPublic := str.StartsWithUpper(parts[1])
	if isMethod {
		parts = str.CleanSplit(line, ")")
		isPublic = str.StartsWithUpper(parts[1])
		return lang.Ternary(isPublic, PUB_METHOD, PRIV_METHOD)
	} else {
		return lang.Ternary(isPublic, PUB_FUNCTION, PRIV_FUNCTION)
	}
}

// Classify type as (public/private) (struct/interface/alias) or code block
func classifyType(line string) CodeType {
	parts := str.SpaceSplit(line)
	if parts[1] == "(" {
		return CODE_GROUP
	}
	isPublic := str.StartsWithUpper(parts[1])
	if strings.HasSuffix(line, " interface {") {
		return lang.Ternary(isPublic, PUB_INTERFACE, PRIV_INTERFACE)
	} else if strings.HasSuffix(line, " struct {") {
		return lang.Ternary(isPublic, PUB_STRUCT, PRIV_STRUCT)
	} else {
		return lang.Ternary(isPublic, PUB_ALIAS, PRIV_ALIAS)
	}
}

// Classify global as (public/private) (const, var) or code block
func classifyGlobal(line string) CodeType {
	parts := str.SpaceSplit(line)
	if parts[1] == "(" {
		return CODE_GROUP
	}
	isPublic := str.StartsWithUpper(parts[1])
	if parts[0] == "const" {
		return lang.Ternary(isPublic, PUB_CONST, PRIV_CONST)
	} else {
		return lang.Ternary(isPublic, PUB_VAR, PRIV_VAR)
	}
}
