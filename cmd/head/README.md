# head

`head` - output the first part of files.

Print the first 10 lines of each FILE to standard output. With more than one FILE, precede each with a header giving the file name. With no FILE, or when FILE is `-`, read standard input.

---

## Usage

```sh
head [OPTION]... [FILE]...
```

With no `FILE`, or when `FILE` is `-`, read standard input.

---

## Options

| Flag | Long Flag | Description |
| :--- | :--- | :--- |
| `-c [-]NUM` | `--bytes=[-]NUM` | Print the first NUM bytes of each file; with leading `-`, print all but the last NUM bytes |
| `-n [-]NUM` | `--lines=[-]NUM` | Print the first NUM lines instead of the first 10; with leading `-`, print all but the last NUM lines |
| `-q` | `--quiet`, `--silent` | Never print headers giving file names |
| `-v` | `--verbose` | Always print headers giving file names |
| `-z` | `--zero-terminated` | Line delimiter is NUL (0x00), not newline |
| | `--help` | Display help message and exit |
| | `--version` | Output version information and exit |

### Multiplier Suffixes for NUM

NUM may have a multiplier suffix:
- `b`: 512
- `kB`, `KB`: 1,000
- `k`, `K`, `KiB`: 1,024
- `MB`: 1,000,000
- `m`, `M`, `MiB`: 1,048,576
- `GB`: 1,000,000,000
- `g`, `G`, `GiB`: 1,073,741,824
- And so on for `T`, `P`.

Legacy syntax `-NUM` (e.g. `-5` or `-20`) is supported as shorthand for `-n NUM`.

---

## Examples

### 1. Default (First 10 Lines)
```sh
$ head file.txt
$ head -
```

### 2. Custom Line Count (`-n`)
```sh
# Print the first 5 lines
$ head -n 5 file.txt
# Legacy shortcut
$ head -5 file.txt
```

### 3. Print by Bytes (`-c`)
```sh
# Print the first 64 bytes
$ head -c 64 file.txt
# Print the first 1 KiB
$ head -c 1K file.txt
```

### 4. Negative Counts (All but the Last N)
```sh
# Print all lines except the trailing 3 lines
$ head -n -3 file.txt

# Print all bytes except the trailing 10 bytes
$ head -c -10 file.txt
```

### 5. Multiple Files and Headers
```sh
$ head file1.txt file2.txt
==> file1.txt <==
line 1
line 2

==> file2.txt <==
line A
line B
```

### 6. Suppressing or Forcing Headers (`-q`, `-v`)
```sh
# Suppress headers when displaying multiple files
$ head -q file1.txt file2.txt

# Always show header even for a single file or stdin
$ head -v file.txt
$ echo "hello" | head -v
==> standard input <==
hello
```
