# pwd

`pwd` - print name of current/working directory.

Print the full filename of the current working directory.

---

## Usage

```sh
pwd [OPTION]...
```

---

## Options

| Flag | Long Flag | Description |
| :--- | :--- | :--- |
| `-L` | `--logical` | Use PWD from environment, even if it contains symlinks (default) |
| `-P` | `--physical` | Avoid all symlinks, resolve to canonical physical directory |
| `-h` | `--help` | Display this help and exit |
| `-v` | `--version` | Output version information and exit |

> **Note**: If both `-L` and `-P` are specified, the last one specified takes precedence.

---

## Examples

### 1. Print current working directory (default logical)
```sh
$ pwd
/home/user/projects
```

### 2. Print physical directory resolving symlinks (`-P`)
If `/var/data` is a symlink to `/mnt/storage/data`:
```sh
$ cd /var/data
$ pwd -L
/var/data

$ pwd -P
/mnt/storage/data
```

### 3. Flag precedence
When combined or repeated, the last flag takes effect:
```sh
$ pwd -LP   # Physical mode (-P is last)
/mnt/storage/data

$ pwd -PL   # Logical mode (-L is last)
/var/data
```

---

## Exit Status

- `0`: Success.
- `1`: Failure (e.g. invalid option or unable to read working directory).
