## Configuration file specification:

If the environment variable `PRETTY_LOGRUS_HOME` is set, we'll look for a `config.json` file in that directory.

The file itself is optional and all fields in it are optional.

### Note about the `Style` object

Several text styles can be overriden in the configuration file. Each override takes the same `Style` object.

| Field path  | Description      | Type                          |
|-------------|------------------|-------------------------------|
| `fgColor`   | Foreground color | One of the available `Colors` |
| `bgColor`   | Background color | One of the available `Colors` |
| `bold`      | Bold text        | `bool`                        |
| `underline` | Underlined text  | `bool`                        |
| `italic`    | Italic text      | `bool`                        |

The app uses `github.com/fatih/color` for colorizing the output, so [all their colors](https://github.com/fatih/color) are available.

![](https://user-images.githubusercontent.com/438920/96832689-03b3e000-13f4-11eb-9803-46f4c4de3406.jpg)

The available `Colors` are:

(bg = background, fg = foreground, hi = high intensity)

| Color name    | Color   |
|---------------|---------|
| `bgBlack`     | Black   |
| `bgHiBlack`   | Black   |
| `fgBlack`     | Black   |
| `fgHiBlack`   | Black   |
| `bgRed`       | Red     |
| `bgHiRed`     | Red     |
| `fgRed`       | Red     |
| `fgHiRed`     | Red     |
| `bgGreen`     | Green   |
| `bgHiGreen`   | Green   |
| `fgGreen`     | Green   |
| `fgHiGreen`   | Green   |
| `bgYellow`    | Yellow  |
| `bgHiYellow`  | Yellow  |
| `fgYellow`    | Yellow  |
| `fgHiYellow`  | Yellow  |
| `bgBlue`      | Blue    |
| `bgHiBlue`    | Blue    |
| `fgBlue`      | Blue    |
| `fgHiBlue`    | Blue    |
| `bgMagenta`   | Magenta |
| `bgHiMagenta` | Magenta |
| `fgMagenta`   | Magenta |
| `fgHiMagenta` | Magenta |
| `bgCyan`      | Cyan    |
| `bgHiCyan`    | Cyan    |
| `fgCyan`      | Cyan    |
| `fgHiCyan`    | Cyan    |
| `bgWhite`     | White   |
| `bgHiWhite`   | White   |
| `fgWhite`     | White   |
| `fgHiWhite`   | White   |

## Configuration file content

### The `keywords` object

Contains lists of keywords the app will look for to identify specific fields in the raw log message text.

If you override any of the keyword lists, the defaults will be ignored so you might need to add those in addition to the new ones.

| Field path                   | Description                                      | Default                  |
|------------------------------|--------------------------------------------------|--------------------------|
| `keywords.messageKeywords`   | List of keywords to locate the log message field | `["msg", "message"]`     |
| `keywords.levelKeywords`     | List of keywords to locate the log level field   | `["level", "log.level"]` |
| `keywords.timestampKeywords` | List of keywords to locate the timestamp field   | `["time", "@timestamp"]` |
| `keywords.errorKeywords`     | List of keywords to locate the error field       | `["error"]`              |
| `keywords.fieldKeywords`     | List of keywords to locate data fields           | `["labels"]`             |

#### Example

```json
// config.json
{
  "keywords": {
    "messageKeywords": ["msg", "message"],
    "levelKeywords": ["level", "log.level"],
    "timestampKeywords": ["time", "@timestamp"],
    "errorKeywords": ["error"],
    "fieldKeywords": ["labels"]
  }
}
```

### The `levelStyles` object

Styles for the log level field.

| Field path                     | Description                                                                                                 | Default |
|--------------------------------|-------------------------------------------------------------------------------------------------------------|---------|
| `levelStyles.default`          | `Style` object. The default styles applied to the log level field unless a more specific override is found. | `TODO`  |
| `levelStyles.<YOUR LOG LEVEL>` | `Style` object. Overrides for specific log levels.                                                          | `TODO`  |

#### Example

```json
// config.json
{
  "levelStyles": {
    "default": {
      "fgColor": "fgGreen"
    },
    "info": {
      "fgColor": "fgHiGreen",
      "bgColor": "bgBlack",
      "bold": true,
      "underline": true,
      "italic": true
    },
    "warn": {
      "fgColor": "fgHiYellow",
      "bgColor": "bgBlack"
    },
    "error": {
      "fgColor": "fgHiRed"
    }
  }
}
```

### The `timestampStyles` object

| Field path                         | Description                                                 | Default |
|------------------------------------|-------------------------------------------------------------|---------|
| `timestampStyles.default`          | `Style` object. The default styles for the timestamp field. | `TODO`  |

#### Example

```json
// config.json
{
  "timestampStyles": {
    "default": {
      "fgColor": "fgHiCyan",
      "bgColor": "bgBlack"
    }
  }
}
```

### The `messageStyles` object

| Field path              | Description                                             | Default |
|-------------------------|---------------------------------------------------------|---------|
| `messageStyles.default` | `Style` object. The default styles for the log message. | `TODO`  |

#### Example

```json
// config.json
{
  "messageStyles": {
    "default": {
      "fgColor": "fgWhite"
    }
  }
}
```


### The `fieldStyles` object

| Field path                            | Description                                                                                       | Default                        |
|---------------------------------------|---------------------------------------------------------------------------------------------------|--------------------------------|
| `fieldStyles.default.key`             | `Style` object. The default styles to be applied to the key/name of the field.                    | `TODO`                         |
| `fieldStyles.default.value`           | `Style` object. The default styles to be applied to the value of the field.                       | `TODO`                         |
| `fieldStyles.highlight.key`           | `Style` object. The styles to be applied for fields matching the `--highlight-key <FIELD>` arg.   | `TODO`                         |
| `fieldStyles.highlight.value`         | `Style` object. The styles to be applied for fields matching the `--highlight-value <VALUE>` arg. | `TODO`                         |
| `fieldStyles.<YOUR FIELD NAME>.key`   | `Style` object. The styles to be applied for key/name for fields matching the name specified.     | `TODO (husk å nevne wildcard)` |
| `fieldStyles.<YOUR FIELD NAME>.value` | `Style` object. The styles to be applied for values for fields matching the name specified.       | `TODO (husk å nevne wildcard)` |

#### Example

```json
// config.json
{
  "fieldStyles": {
    "default": {
      "key": {
        "fgColor:": "fgYellow"
      },
      "value": {
        "fgColor": "fgGreen",
        "bold": true
      }
    },
    "highlight": {
      "key": {
        "fgColor:": "fgRed",
        "bold": true
      },
      "value": {
        "fgColor": "fgRed",
        "bold": true,
        "underline": true
      }
    },
    "trace.id": {
      "key": {
        "fgColor:": "fgHiCyan"
      },
      "value": {
        "fgColor": "fgHiCyan"
      }
    },
    "labels.*": {
      "key": {
        "fgColor:": "fgHiMagenta"
      },
      "value": {
        "fgColor": "fgWhite",
        "bgColor": "bgBlack",
        "bold": true
      }
    },
    "*meta": {
      "key": {
        "fgColor:": "fgWhite"
      },
      "value": {
        "fgColor": "fgWhite"
      }
    }
  }
}
```