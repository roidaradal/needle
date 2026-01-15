package needle

import (
	"fmt"

	"github.com/roidaradal/fn/str"
)

const (
	tbody        = "tbody"
	tr           = "tr"
	td           = "td"
	th           = "th"
	left         = "left"
	center       = "center"
	right        = "right"
	centerLocal  = "center local"
	centerGlobal = "center global"
)

type Tag struct {
	id      string
	title   string
	class   string
	onclick string
	rowspan int
	colspan int
}

type TagOption func(*Tag)

// Wrap string by given tags (last is inner tag)
func wrapTags(text string, tags ...string) string {
	i := len(tags) - 1
	for i >= 0 {
		text = wrapTag(tags[i], text)
		i--
	}
	return text
}

// Create button tag
func button(text string, options ...TagOption) string {
	return wrapTag("button", text, options...)
}

// Wrap string by given tag
func wrapTag(tag, text string, options ...TagOption) string {
	t := &Tag{}
	for _, opt := range options {
		opt(t)
	}
	var title, rowspan, colspan string
	var class, id, onclick string
	if t.class != "" {
		class = strAttr("class", t.class)
	}
	if t.id != "" {
		id = strAttr("id", t.id)
	}
	if t.onclick != "" {
		onclick = strAttr("onclick", t.onclick)
	}
	if t.title != "" {
		class = strAttr("title", t.title)
	}
	if t.rowspan > 1 {
		rowspan = intAttr("rowspan", t.rowspan)
	}
	if t.colspan > 1 {
		colspan = intAttr("colspan", t.colspan)
	}
	attrs := str.Join(" ", id, class, title, rowspan, colspan, onclick)
	return fmt.Sprintf("<%s %s>%s</%s>", tag, attrs, text, tag)
}

// Create string attribute
func strAttr(attr string, value string) string {
	return fmt.Sprintf("%s=%q", attr, value)
}

// Create int attribute
func intAttr(attr string, value int) string {
	return fmt.Sprintf("%s='%d'", attr, value)
}

// Tag option: class
func withClass(class string) TagOption {
	return func(tag *Tag) {
		tag.class = class
	}
}

// Tag option: title
func withTitle(title string) TagOption {
	return func(tag *Tag) {
		tag.title = title
	}
}

// Tag option: id
func withID(id string) TagOption {
	return func(tag *Tag) {
		tag.id = id
	}
}

// Tag option: onclick
func onclick(onclickFn string) TagOption {
	return func(tag *Tag) {
		tag.onclick = onclickFn
	}
}

// Tag option: rowspan
func withRowspan(rowspan int) TagOption {
	return func(tag *Tag) {
		tag.rowspan = rowspan
	}
}

// Tag option: colspan
func withColspan(colspan int) TagOption {
	return func(tag *Tag) {
		tag.colspan = colspan
	}
}
