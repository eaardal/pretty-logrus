package main

import (
	"fmt"
	"github.com/fatih/color"
	"strings"
)

type ColorStringFn func(format string, a ...interface{}) string
type ColorFn func(format string, a ...interface{})

func NoColor(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

var colorCodeStrings = map[string]ColorStringFn{
	"black":     color.BlackString,
	"red":       color.RedString,
	"green":     color.GreenString,
	"yellow":    color.YellowString,
	"blue":      color.BlueString,
	"magenta":   color.MagentaString,
	"cyan":      color.CyanString,
	"white":     color.WhiteString,
	"hiBlack":   color.HiBlackString,
	"hiRed":     color.HiRedString,
	"hiGreen":   color.HiGreenString,
	"hiYellow":  color.HiYellowString,
	"hiBlue":    color.HiBlueString,
	"hiMagenta": color.HiMagentaString,
	"hiCyan":    color.HiCyanString,
	"hiWhite":   color.HiWhiteString,
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
	defaultTimestamp := blue(timestamp)

	if styles == nil {
		logDebug("No timestamp styles defined in config\n")
		return defaultTimestamp
	}

	defaultStyle, ok := styles["default"]
	if ok {
		logDebug("Applying default styles %+v for timestamp\n", defaultStyle)
		defaultTimestamp = applyStyles(&defaultStyle).Sprint(timestamp)
	}

	style := findStyle(styles, timestamp)
	if style == nil {
		logDebug("No style defined for timestamp %s\n", timestamp)
		return defaultTimestamp
	}

	logDebug("Applying styles %+v for timestamp\n", style)
	return applyStyles(style).Sprint(timestamp)
}

func applyMessageStyle(message string, styles map[string]Style) string {
	defaultMessage := white(message)

	if styles == nil {
		logDebug("No message styles defined in config\n")
		return defaultMessage
	}

	defaultStyle, ok := styles["default"]
	if ok {
		logDebug("Applying default styles %+v for message\n", defaultStyle)
		defaultMessage = applyStyles(&defaultStyle).Sprint(message)
	}

	style := findStyle(styles, message)
	if style == nil {
		logDebug("No style defined for message %s\n", message)
		return defaultMessage
	}

	logDebug("Applying styles %+v for message\n", style)
	return applyStyles(style).Sprint(message)
}

func applyLevelStyle(level string, styles map[string]Style) string {
	defaultLevel := cyan(level)

	if level == "warning" {
		defaultLevel = yellow(level)
	}

	if level == "error" || level == "fatal" {
		defaultLevel = red(level)
	}

	if styles == nil {
		logDebug("No level styles defined in config\n")
		return defaultLevel
	}

	style, ok := styles["default"]
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

func applyFieldNameStyle(fieldName string, styles map[string]KeyValueStyle) string {
	defaultFieldName := yellow(fieldName)

	if styles == nil {
		logDebug("No field styles defined in config\n")
		return defaultFieldName
	}

	defaultStyles, ok := styles["default"]
	if ok && defaultStyles.Key != nil {
		defaultFieldName = applyStyles(defaultStyles.Key).Sprint(fieldName)
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

func applyFieldValueStyle(fieldName, fieldValue string, styles map[string]KeyValueStyle) string {
	defaultFieldValue := green(fieldValue)

	if styles == nil {
		logDebug("No field styles defined in config\n")
		return defaultFieldValue
	}

	defaultStyles, ok := styles["default"]
	if ok && defaultStyles.Key != nil {
		defaultFieldValue = applyStyles(defaultStyles.Key).Sprint(fieldName)
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

func findStyle(styles map[string]Style, fieldName string) *Style {
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
