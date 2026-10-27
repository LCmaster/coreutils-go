# touch

`touch` - change file timestamps.

Update the access and modification times of each `FILE` to the current time.

A `FILE` argument that does not exist is created empty, unless `-c` is supplied.

A `FILE` argument string of `-` is handled specially and causes `touch` to change the times of the file associated with standard output.

---

## Usage

```sh
touch [OPTION]... FILE...
```

---

## Options

| Flag | Long Flag | Description |
| :--- | :--- | :--- |
| `-a` | | Change only the access time |
| `-c` | `--no-create` | Do not create any files |
| `-d` | `--date=STRING` | Parse `STRING` and use it instead of current time |
| `-f` | | (Ignored; for BSD compatibility) |
| `-m` | | Change only the modification time |
| `-r` | `--reference=FILE` | Use this file's times instead of current time |
| `-t` | `STAMP` | Use `[[CC]YY]MMDDhhmm[.ss]` instead of current time |
| | `--time=WORD` | Change the specified time: `WORD` is `access`, `atime`, or `use` (equivalent to `-a`); `WORD` is `modify` or `mtime` (equivalent to `-m`) |
| | `--help` | Display this help message and exit |
| | `--version` | Output version information and exit |

> **Note**:
> - If neither `-a` nor `-m` (nor `--time`) is specified, both the access time and modification time are updated.
> - In `-t STAMP`, two-digit years `YY` in the range `69–99` refer to `1969–1999`, and `00–68` refer to `2000–2068`. If `CCYY` is omitted, the current year is used.

---

## Examples

### 1. Create a New Empty File or Update Timestamps

If `file.txt` does not exist, it is created as an empty file; otherwise, its access and modification times are updated to the current time:

```sh
$ touch file.txt
```

### 2. Touch Multiple Files at Once

Create or update several files in a single command:

```sh
$ touch a.txt b.txt c.txt
```

### 3. Update Timestamps Without Creating Missing Files (`-c`)

Prevent `touch` from creating files that do not already exist:

```sh
$ touch -c missing.txt
```

### 4. Change Only Access Time (`-a`) or Modification Time (`-m`)

Update only the access time while preserving the existing modification time (or vice versa):

```sh
$ touch -a file.txt
$ touch -m file.txt
$ touch --time=atime file.txt
$ touch --time=mtime file.txt
```

### 5. Set Explicit Timestamp Using POSIX Format (`-t`)

Specify a timestamp in `[[CC]YY]MMDDhhmm[.ss]` format (e.g., May 15, 2026 at 14:30:00):

```sh
$ touch -t 202605151430.00 file.txt
```

### 6. Set Timestamp from a Date String (`-d` / `--date`)

Parse human-readable ISO-8601, RFC3339, or epoch (`@SECONDS`) date strings:

```sh
$ touch -d "2026-05-15 14:30:00" file.txt
$ touch --date="2026-01-01" file.txt
$ touch -d "@1700000000" file.txt
```

### 7. Copy Timestamps from a Reference File (`-r` / `--reference`)

Set `target.txt`'s access and modification times to match those of `source.txt`:

```sh
$ touch -r source.txt target.txt
```

---

## Exit Status

- `0`: All files were updated or created successfully.
- `1`: An error occurred (e.g., missing file operand, invalid date/timestamp format, missing reference file, or permission denied).

---

## Standards & Compatibility

This utility is modeled after GNU Coreutils `touch` and conforms to POSIX.1-2017:
- Supports standard POSIX `-t [[CC]YY]MMDDhhmm[.ss]` timestamp parsing and two-digit year century inference.
- Supports GNU extensions including `-d`/`--date`, `--reference`, `--no-create`, and `--time=WORD`.
