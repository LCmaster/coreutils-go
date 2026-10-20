# echo

`echo` - write arguments to standard output.

Echo the STRING(s) to standard output, separated by spaces and terminated by a newline.

---

## Usage

```sh
echo [OPTION]... [STRING]...
```

---

## Options

| Flag | Long Flag | Description |
| :--- | :--- | :--- |
| `-n` | | Do not output the trailing newline |
| `-e` | | Enable interpretation of backslash escape sequences |
| `-E` | | Disable interpretation of backslash escapes *(default)* |
| `-h` | `--help` | Display this help message and exit |
| `-v` | `--version` | Output version information and exit |

---

## Escape Sequences

When `-e` is enabled, the following backslash-escaped sequences are recognized:

| Sequence | Name | Description |
| :--- | :--- | :--- |
| `\\` | Backslash | Outputs a literal backslash character (`\`) |
| `\a` | Alert / Bell | Produces an audible alert or terminal bell (BEL, ASCII 7) |
| `\b` | Backspace | Moves cursor back one space (ASCII 8) |
| `\c` | Cancel output | Produce no further output and suppress trailing newline |
| `\e` | Escape | Escape character (ASCII 27 / 0x1B) |
| `\f` | Form feed | Advances to the next page / form feed (ASCII 12) |
| `\n` | New line | Moves to the beginning of the next line (ASCII 10) |
| `\r` | Carriage return | Moves cursor to the start of the current line (ASCII 13) |
| `\t` | Horizontal tab | Moves cursor to the next horizontal tab stop (ASCII 9) |
| `\v` | Vertical tab | Moves cursor to the next vertical tab stop (ASCII 11) |
| `\0NNN` | Octal byte | Character with octal value `NNN` (1 to 3 octal digits, 0–377) |
| `\xHH` | Hexadecimal byte | Character with hexadecimal value `HH` (1 to 2 hex digits) |

---

## Examples

### 1. Basic text printing
Print strings separated by a space and followed by a newline:
```sh
$ echo Hello World
Hello World
```

### 2. Suppress the trailing newline (`-n`)
Useful for prompts or piping without an extra trailing line break:
```sh
$ echo -n "Enter name: "
Enter name: $
```

### 3. Interpret escape sequences (`-e`)
Enable special characters such as tabs and newlines:
```sh
$ echo -e "Line 1\nLine 2\tIndented"
Line 1
Line 2	Indented
```

### 4. Produce colored output using ANSI escape codes
Using `\e` (escape) to format text in terminals:
```sh
$ echo -e "\e[32mSuccess:\e[0m Operation completed."
```

### 5. Terminate output early (`\c`)
Any text after `\c` is ignored and the trailing newline is omitted:
```sh
$ echo -e "First part\cThis will not be printed"
First part$
```

### 6. Printing bytes using octal or hex
```sh
# Octal for 'A' (101)
$ echo -e "\0101"
A

# Hexadecimal for 'A' (41)
$ echo -e "\x41"
A
```

### 7. Disabling escape sequences (`-E`)
Treat backslashes literally (default behavior, or to override an earlier `-e`):
```sh
$ echo -E "Hello\nWorld"
Hello\nWorld
```

---

## Exit Status

- **`0`**: Success.
- **`>0`**: An error occurred while writing output.

---

## Standards & Compatibility

This command is modeled after the GNU Coreutils `echo` utility and supports common POSIX extensions for shell scripting and terminal output.