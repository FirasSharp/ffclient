# ffclient - FuckingFast.co Download Client

A fast and efficient command-line client for downloading files from [FuckingFast.co](https://fuckingfast.co/) with multi-download support, written in Go.

![Go Version](https://img.shields.io/badge/go-%3E%3D1.20-blue.svg)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## Features

- Multi-file download support
- Input from command line or text file
- Simple and fast
- Cross-platform (Windows, Linux, macOS)

## Installation

### Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/FirasSharp/ffclient/releases).

### From Source

1. Ensure you have Go (≥1.20) installed
2. Run:
   ```sh
   go install github.com/FirasSharp/ffclient@latest
   ```

## Usage

```sh
ffclient [flags]
```

### Flags

| Flag        | Description                                                                 | Default Value       |
|-------------|-----------------------------------------------------------------------------|---------------------|
| `--savePath` | Destination directory for downloaded files                                  | Current directory   |
| `--inputFile` | Text file containing URLs to download (one fuckingfast.co URL per line)    | ""                  |
| `--links`    | Comma-separated fuckingfast.co URLs                                        | ""                  |

### Examples

1. Download single file:
   ```sh
   ffclient --links "https://fuckingfast.co/file1"
   ```

2. Download multiple files:
   ```sh
   ffclient --links "https://fuckingfast.co/file1,https://fuckingfast.co/file2"
   ```

3. Download files from a text file:
   ```sh
   ffclient --inputFile downloads.txt --savePath ~/Downloads/fast_files
   ```

4. Combine both methods (files from inputFile and links will be downloaded):
   ```sh
   ffclient --inputFile downloads.txt --links "https://fuckingfast.co/another_file"
   ```

## Input File Format

The input file should contain one FuckingFast.co URL per line:
```
https://fuckingfast.co/file1
https://fuckingfast.co/file2
https://fuckingfast.co/file3
```

## License

MIT - See [LICENSE](LICENSE) for more information.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.