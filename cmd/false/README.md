# false

`false` - do nothing, unsuccessfully.

Exit with a status code indicating failure (`1`).

---

## Usage

```sh
false [ignored command line arguments]
false OPTION
```

---

## Options

| Flag | Long Flag | Description |
| :--- | :--- | :--- |
| `-h` | `--help` | Display this help message and exit |
| `-v` | `--version` | Output version information and exit |

> **Note:** Any other arguments or flags passed to `false` are ignored, and the command will always exit with status `1`.

---

## Examples

### 1. Basic invocation
```sh
$ false
$ echo $?
1
```

### 2. Used in shell conditional loops
```sh
until false; do
  echo "false exited 1"
  break
done
```

### 3. Ignored arguments
```sh
$ false --any-flag ignored argument
$ echo $?
1
```

---

## Exit Status

- **`0`**: When sole argument is `--help` or `--version`.
- **`1`**: All other cases (failure).
