# cat

`cat` - concatenate FILE(s) to standard output.

Concatenate FILE(s), or standard input, to standard output.

---

## Usage

```sh
cat [OPTION]... [FILE]...
```

With no `FILE`, or when `FILE` is `-`, read standard input.

---

## Options

| Flag | Long Flag | Description |
| :--- | :--- | :--- |
| `-A` | `--show-all` | Equivalent to `-vET` |
| `-b` | `--number-nonblank` | Number nonempty output lines, overrides `-n` |
| `-e` | | Equivalent to `-vE` |
| `-E` | `--show-ends` | Display `$` at end of each line |
| `-n` | `--number` | Number all output lines |
| `-s` | `--squeeze-blank` | Suppress repeated empty output lines |
| `-t` | | Equivalent to `-vT` |
| `-T` | `--show-tabs` | Display TAB characters as `^I` |
| `-u` | | (ignored, POSIX compatibility) |
| `-v` | `--show-nonprinting` | Use `^` and `M-` notation, except for LFD and TAB |
| | `--help` | Display this help and exit |
| | `--version` | Output version information and exit |

---

## Examples

### 1. Concatenate files
Print contents of one or more files in sequence:
```sh
$ cat file1.txt file2.txt
```

### 2. Read from standard input
Read standard input when no files are specified or when `-` is given:
```sh
$ cat -
$ echo "Hello" | cat
Hello
```

### 3. Number all output lines (`-n`)
Prefix each line with a 1-indexed, tab-separated line number:
```sh
$ cat -n file.txt
     1	First line
     2	Second line
     3	Third line
```

### 4. Number only non-empty lines (`-b`)
Lines containing characters are numbered; empty lines (only newline) are skipped:
```sh
$ cat -b file.txt
     1	First line

     2	Third line
```

### 5. Squeeze repeated empty lines (`-s`)
Condense multiple consecutive blank lines into a single blank line:
```sh
$ cat -s file.txt
```

### 6. Show line endings (`-E`)
Display a `$` character at the end of each line before the newline:
```sh
$ cat -E file.txt
First line$
Second line$
```

### 7. Display tab characters (`-T`)
Print horizontal tabs as visible `^I` characters:
```sh
$ cat -T file.txt
Col1^ICol2^ICol3
```

### 8. Show non-printing characters (`-v`)
Display control characters using caret notation (`^@`–`^_`, `^?`) and high bytes with `M-`:
```sh
$ cat -v binary.dat
```

### 9. Show all (`-A` / `--show-all`)
Equivalent to `-vET` (shows non-printing characters, line ends, and tabs):
```sh
$ cat -A file.txt
Col1^ICol2$
```

---

## Exit Status

- `0`: Success.
- `1`: Minor problems (e.g., file not found, permission denied, or reading directory).
