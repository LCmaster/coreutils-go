# true

`true` - do nothing, successfully.

Exit with a status code indicating success (`0`).

---

## Usage

```sh
true [ignored command line arguments]
true OPTION
```

---

## Options

| Flag | Long Flag | Description |
| :--- | :--- | :--- |
| `-h` | `--help` | Display this help message and exit |
| `-v` | `--version` | Output version information and exit |

> **Note:** Any other arguments or flags passed to `true` are ignored, and the command will always exit with status `0`.

---

## Examples

### 1. Basic invocation
```sh
$ true
$ echo $?
0
```

### 2. Used in shell conditional loops
```sh
while true; do
  echo "running..."
  break
done
```

### 3. Ignored arguments
```sh
$ true --any-flag ignored argument
$ echo $?
0
```

---

## Exit Status

Always exits with status `0`.
