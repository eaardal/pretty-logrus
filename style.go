package main

import (
	"github.com/fatih/color"
	"strings"
)

const DefaultStylesKey = "default"
const HighlightStylesKey = "highlight"

var DefaultFieldStyles = map[string]KeyValueStyle{
	DefaultStylesKey: {
		Key: &Style{
			FgColor: getColorCode(color.FgYellow),
		},
		Value: &Style{
			FgColor: getColorCode(color.FgGreen),
		},
	},
	HighlightStylesKey: {
		Key: &Style{
			FgColor:   getColorCode(color.FgRed),
			Bold:      boolPtr(true),
			Italic:    boolPtr(true),
			Underline: boolPtr(true),
		},
		Value: &Style{
			FgColor:   getColorCode(color.FgRed),
			Bold:      boolPtr(true),
			Italic:    boolPtr(true),
			Underline: boolPtr(true),
		},
	},
}

var DefaultLevelStyles = map[string]Style{
	DefaultStylesKey: {
		FgColor: getColorCode(color.FgHiGreen),
	},
	"trace": {
		FgColor: getColorCode(color.FgWhite),
	},
	"debug": {
		FgColor: getColorCode(color.FgCyan),
	},
	"info": {
		FgColor: getColorCode(color.FgHiGreen),
	},
	"warning": {
		FgColor: getColorCode(color.FgYellow),
	},
	"error": {
		FgColor: getColorCode(color.FgRed),
	},
	"err": {
		FgColor: getColorCode(color.FgRed),
	},
	"fatal": {
		FgColor: getColorCode(color.FgRed),
	},
	"panic": {
		FgColor: getColorCode(color.FgRed),
	},
}

var DefaultMessageStyles = map[string]Style{
	DefaultStylesKey: {
		FgColor: getColorCode(color.FgWhite),
	},
}

var DefaultTimestampStyles = map[string]Style{
	DefaultStylesKey: {
		FgColor: getColorCode(color.FgBlue),
	},
}

func getColorCode(attr color.Attribute) *string {
	for key, value := range colorCodes {
		if value == attr {
			return &key
		}
	}
	empty := ""
	return &empty
}

var colorCodes = map[string]color.Attribute{
	"bgBlack":   color.BgBlack,
	"bgHiBlack": color.BgHiBlack,
	"fgBlack":   color.FgBlack,
	"fgHiBlack": color.FgHiBlack,

	"bgRed":   color.BgRed,
	"bgHiRed": color.BgHiRed,
	"fgRed":   color.FgRed,
	"fgHiRed": color.FgHiRed,

	"bgGreen":   color.BgGreen,
	"bgHiGreen": color.BgHiGreen,
	"fgGreen":   color.FgGreen,
	"fgHiGreen": color.FgHiGreen,

	"bgYellow":   color.BgYellow,
	"bgHiYellow": color.BgHiYellow,
	"fgYellow":   color.FgYellow,
	"fgHiYellow": color.FgHiYellow,

	"bgBlue":   color.BgBlue,
	"bgHiBlue": color.BgHiBlue,
	"fgBlue":   color.FgBlue,
	"fgHiBlue": color.FgHiBlue,

	"bgMagenta":   color.BgMagenta,
	"bgHiMagenta": color.BgHiMagenta,
	"fgMagenta":   color.FgMagenta,
	"fgHiMagenta": color.FgHiMagenta,

	"bgCyan":   color.BgCyan,
	"bgHiCyan": color.BgHiCyan,
	"fgCyan":   color.FgCyan,
	"fgHiCyan": color.FgHiCyan,

	"bgWhite":   color.BgWhite,
	"bgHiWhite": color.BgHiWhite,
	"fgWhite":   color.FgWhite,
	"fgHiWhite": color.FgHiWhite,
}

func applyStyles(styles *Style) *color.Color {
	c := color.New()
	if styles.BgColor != nil {
		c.Add(colorCodes[*styles.BgColor])
	}
	if styles.FgColor != nil {
		c.Add(colorCodes[*styles.FgColor])
	}
	if styles.Bold != nil && *styles.Bold {
		c.Add(color.Bold)
	}
	if styles.Underline != nil && *styles.Underline {
		c.Add(color.Underline)
	}
	if styles.Italic != nil && *styles.Italic {
		c.Add(color.Italic)
	}
	return c
}

func applyTimestampStyle(timestamp string, styles map[string]Style) string {
	defaultTimestamp := color.New().Sprint(timestamp)

	if styles == nil {
		logDebug("No timestamp styles defined in config, falling back on defaults\n")
		styles = DefaultTimestampStyles
	}

	defaultStyle, ok := styles[DefaultStylesKey]
	if ok {
		logDebug("Applying default styles %+v for timestamp\n", defaultStyle)
		defaultTimestamp = applyStyles(&defaultStyle).Sprint(timestamp)
	}

	style := findStyleOverride(styles, timestamp)
	if style == nil {
		logDebug("No style defined for timestamp %s\n", timestamp)
		return defaultTimestamp
	}

	logDebug("Applying styles %+v for timestamp\n", style)
	return applyStyles(style).Sprint(timestamp)
}

func applyMessageStyle(message string, styles map[string]Style) string {
	defaultMessage := color.New().Sprint(message)

	if styles == nil {
		logDebug("No message styles defined in config, falling back on defaults\n")
		styles = DefaultMessageStyles
	}

	defaultStyle, ok := styles[DefaultStylesKey]
	if ok {
		logDebug("Applying default styles %+v for message\n", defaultStyle)
		defaultMessage = applyStyles(&defaultStyle).Sprint(message)
	}

	style := findStyleOverride(styles, message)
	if style == nil {
		logDebug("No style defined for message %s\n", message)
		return defaultMessage
	}

	logDebug("Applying styles %+v for message\n", style)
	return applyStyles(style).Sprint(message)
}

func applyLevelStyle(level string, styles map[string]Style) string {
	defaultLevel := color.New().Sprint(level)

	if styles == nil {
		logDebug("No level styles defined in config, falling back on defaults\n")
		styles = DefaultLevelStyles
	}

	style, ok := styles[DefaultStylesKey]
	if ok {
		logDebug("Applying default styles %+v for level %s\n", style, level)
		defaultLevel = applyStyles(&style).Sprint(level)
	}

	style, ok = styles[level]
	if !ok {
		logDebug("No style defined for level %s\n", level)
		return defaultLevel
	}

	logDebug("Applying styles %+v for level %s\n", style, level)
	return applyStyles(&style).Sprint(level)
}

func applyFieldNameStyle(fieldName string, styles map[string]KeyValueStyle, highlightKey string) string {
	defaultFieldName := color.New().Sprint(fieldName)

	if styles == nil {
		logDebug("No field styles defined in config, falling back on defaults\n")
		styles = DefaultFieldStyles
	}

	defaultStyles, ok := styles[DefaultStylesKey]
	if ok && defaultStyles.Key != nil {
		defaultFieldName = applyStyles(defaultStyles.Key).Sprint(fieldName)
	}

	highlightedFieldName, highlighted := tryApplyHighlightStyle(fieldName, highlightKey, styles, func(s KeyValueStyle) *Style { return s.Key })
	if highlighted {
		return highlightedFieldName
	}

	style := findKeyValueStyle(styles, fieldName)
	if style == nil {
		logDebug("No style defined for field %s\n", fieldName)
		return defaultFieldName
	}

	fieldNameStyle := style.Key
	if fieldNameStyle == nil {
		logDebug("No key style defined for field %s\n", fieldName)
		return defaultFieldName
	}

	logDebug("Applying styles %+v for field %s\n", fieldNameStyle, fieldName)
	return applyStyles(fieldNameStyle).Sprint(fieldName)
}

func applyFieldValueStyle(fieldName, fieldValue string, styles map[string]KeyValueStyle, highlightValue string) string {
	defaultFieldValue := color.New().Sprint(fieldValue)

	if styles == nil {
		logDebug("No field styles defined in config, falling back on defaults\n")
		styles = DefaultFieldStyles
	}

	defaultStyles, ok := styles[DefaultStylesKey]
	if ok && defaultStyles.Value != nil {
		defaultFieldValue = applyStyles(defaultStyles.Value).Sprint(fieldValue)
	}

	highlightedFieldValue, highlighted := tryApplyHighlightStyle(fieldValue, highlightValue, styles, func(s KeyValueStyle) *Style { return s.Value })
	if highlighted {
		return highlightedFieldValue
	}

	style := findKeyValueStyle(styles, fieldName)
	if style == nil {
		logDebug("No style defined for field %s\n", fieldName)
		return defaultFieldValue
	}

	fieldValueStyle := style.Value
	if fieldValueStyle == nil {
		logDebug("No value style defined for field %s\n", fieldName)
		return defaultFieldValue
	}

	logDebug("Applying styles %+v for field %s\n", fieldValueStyle, fieldValue)
	return applyStyles(fieldValueStyle).Sprint(fieldValue)
}

func findKeyValueStyle(styles map[string]KeyValueStyle, fieldName string) *KeyValueStyle {
	style, ok := styles[fieldName]
	if ok {
		return &style
	}

	for key := range styles {
		if strings.HasSuffix(key, "*") {
			cleanKey := strings.TrimSuffix(key, "*")

			if strings.HasPrefix(fieldName, cleanKey) {
				logDebug("Field '%s' is in list because of trailing wildcard '%s'\n", fieldName, key)
				style, ok = styles[key]
				if ok {
					return &style
				}
			}
		}

		if strings.HasPrefix(key, "*") {
			cleanKey := strings.TrimPrefix(key, "*")

			if strings.HasSuffix(fieldName, cleanKey) {
				logDebug("Field '%s' is in list because of leading wildcard '%s'\n", fieldName, key)
				style, ok = styles[key]
				if ok {
					return &style
				}
			}
		}

		if strings.HasPrefix(key, "*") && strings.HasSuffix(key, "*") {
			cleanKey := strings.Trim(key, "*")

			if strings.Contains(fieldName, cleanKey) {
				logDebug("Field '%s' is in list because of leading and trailing wildcard '%s'\n", fieldName, key)
				style, ok = styles[key]
				if ok {
					return &style
				}
			}
		}
	}

	return nil
}

func findStyleOverride(styles map[string]Style, fieldName string) *Style {
	style, ok := styles[fieldName]
	if ok {
		return &style
	}

	for key := range styles {
		if strings.HasSuffix(key, "*") {
			cleanKey := strings.TrimSuffix(key, "*")

			if strings.HasPrefix(fieldName, cleanKey) {
				logDebug("Field '%s' is in list because of trailing wildcard '%s'\n", fieldName, key)
				style, ok = styles[key]
				if ok {
					return &style
				}
			}
		}

		if strings.HasPrefix(key, "*") {
			cleanKey := strings.TrimPrefix(key, "*")

			if strings.HasSuffix(fieldName, cleanKey) {
				logDebug("Field '%s' is in list because of leading wildcard '%s'\n", fieldName, key)
				style, ok = styles[key]
				if ok {
					return &style
				}
			}
		}

		if strings.HasPrefix(key, "*") && strings.HasSuffix(key, "*") {
			cleanKey := strings.Trim(key, "*")

			if strings.Contains(fieldName, cleanKey) {
				logDebug("Field '%s' is in list because of leading and trailing wildcard '%s'\n", fieldName, key)
				style, ok = styles[key]
				if ok {
					return &style
				}
			}
		}
	}

	return nil
}

func tryApplyHighlightStyle(value string, highlight string, styles map[string]KeyValueStyle, selectStyle func(s KeyValueStyle) *Style) (string, bool) {
	if highlight == "" {
		return value, false
	}

	defaultHlStyle := color.New(color.FgHiRed, color.Bold, color.Italic, color.Underline)

	applyHighlightStyle := func(value string) string {
		hlStyle, hasHlStyle := styles["highlight"]
		if hasHlStyle && selectStyle(hlStyle) != nil {
			return applyStyles(selectStyle(hlStyle)).Sprint(value)
		}
		return defaultHlStyle.Sprint(value)
	}

	if highlight == value {
		return applyHighlightStyle(value), true
	}

	if strings.HasPrefix(highlight, "*") {
		cleanKey := strings.TrimPrefix(highlight, "*")

		if strings.HasSuffix(value, cleanKey) {
			return applyHighlightStyle(value), true
		}
	}

	if strings.HasSuffix(highlight, "*") {
		cleanKey := strings.TrimSuffix(highlight, "*")

		if strings.HasPrefix(value, cleanKey) {
			return applyHighlightStyle(value), true
		}
	}

	if strings.HasPrefix(highlight, "*") && strings.HasSuffix(highlight, "*") {
		cleanKey := strings.Trim(highlight, "*")

		if strings.Contains(value, cleanKey) {
			return applyHighlightStyle(value), true
		}
	}

	return value, false
}
