# yes

`yes` - output a string repeatedly until killed.

Repeatedly output a line with all specified STRING(s), or `y`.

---

## Usage

```sh
yes [STRING]...
yes OPTION
```

---

## Options

| Flag | Long Flag | Description |
| :--- | :--- | :--- |
| `-h` | `--help` | Display this help message and exit |
| `-v` | `--version` | Output version information and exit |

---

## Performance

`coreutils-go yes` implements memory-aligned 64 KB block streaming, repeating the target pattern across a reusable memory buffer before issuing write calls. This avoids line-by-line syscall overhead and achieves multiple gigabytes per second throughput.

---

## Examples

### 1. Default output
Continuously print `y\n`:
```sh
$ yes
y
y
y
...
```

### 2. Custom string or multiple words
```sh
$ yes custom message
custom message
custom message
...
```

### 3. Piping to automated confirmations
Automate interactive command line confirmation prompts:
```sh
$ yes | rm -i *.tmp
```

### 4. Bounded output with head
Gracefully stops when consumer pipe closes:
```sh
$ yes | head -n 3
y
y
y
```

---

## Exit Status

- **`0`**: Successful termination on broken pipe (e.g. piped into `head`) or sole `--help`/`--version`.
- **`>0`**: An unrecoverable I/O error occurred.
