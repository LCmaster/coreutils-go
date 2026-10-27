# wc

`wc` - print newline, word, and byte counts for each file.

Print newline, word, and byte counts for each `FILE`, and a total line if more than one `FILE` is specified. A word is a non-empty, maximal sequence of characters delimited by whitespace.

With no `FILE`, or when `FILE` is `-`, read standard input.

---

## Usage

```sh
wc [OPTION]... [FILE]...
wc [OPTION]... --files0-from=F
```

With no `FILE`, or when `FILE` is `-`, read standard input.

---

## Options

| Flag | Long Flag | Description |
| :--- | :--- | :--- |
| `-c` | `--bytes` | Print the byte counts |
| `-m` | `--chars` | Print the character counts |
| `-l` | `--lines` | Print the newline counts |
| | `--files0-from=F` | Read input from the files specified by NUL-terminated names in file `F`; if `F` is `-`, read from standard input |
| `-L` | `--max-line-length` | Print the maximum display width (length of the longest line) |
| `-w` | `--words` | Print the word counts |
| | `--help` | Display this help message and exit |
| | `--version` | Output version information and exit |

> **Note**:
> - If no count option (`-c`, `-m`, `-l`, `-L`, `-w`) is specified, the default is equivalent to `-l -w -c` (lines, words, and bytes).
> - When multiple count options are selected, outputs are always printed in the canonical order: **newlines**, **words**, **characters**, **bytes**, **maximum line length**.

---

## Examples

### 1. Default Count (Lines, Words, and Bytes)

By default, `wc` outputs the number of lines, words, and bytes:

```sh
$ wc file.txt
  15  120  840 file.txt
```

### 2. Read from Standard Input

Read standard input when no files are provided or when `-` is given:

```sh
$ echo "The quick brown fox jumps over the lazy dog" | wc
       1       9      44

$ cat file.txt | wc -l
15
```

### 3. Count Lines Only (`-l`)

Count newline characters (`\n`):

```sh
$ wc -l file.txt
15 file.txt
```

### 4. Count Words Only (`-w`)

Count whitespace-delimited words:

```sh
$ wc -w file.txt
120 file.txt
```

### 5. Count Bytes (`-c`) vs. Characters (`-m`)

In UTF-8 text, multi-byte characters cause byte counts and character counts to differ:

```sh
$ echo -n "Café ☕" | wc -c
10

$ echo -n "Café ☕" | wc -m
6
```

### 6. Maximum Line Length (`-L`)

Find the display width of the longest line in the input (tabs are expanded to 8 spaces):

```sh
$ wc -L file.txt
78 file.txt
```

### 7. Multiple Files and Total Line

When more than one file is specified, `wc` prints counts for each file followed by a cumulative `total` summary line:

```sh
$ wc file1.txt file2.txt
  10   50  300 file1.txt
  20  100  700 file2.txt
  30  150 1000 total
```

### 8. Read File Names from NUL-Terminated Stream (`--files0-from`)

Process files safely without whitespace or quote issues, especially when paired with `find -print0`:

```sh
$ find . -name "*.go" -print0 | wc -l --files0-from=-
  120 ./cmd/cat/cat.go
   85 ./cmd/cat/main.go
  205 total
```

---

## Exit Status

- `0`: Success.
- `1`: Minor problems (e.g., file not found, permission denied, or error reading directory).

---

## Standards & Compatibility

This utility is modeled after GNU Coreutils `wc` and conforms to POSIX.1-2017:
- A line is defined as a sequence of characters terminated by a newline (`\n`). Characters after the last newline without a terminating newline are not counted as a line.
- A word is defined as a non-empty sequence of non-whitespace characters delimited by standard whitespace (spaces, tabs, newlines, form feeds, carriage returns).
