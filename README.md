# Comments Checker

Comments Checker is a pre-commit hook written in Go, designed for fast parsing and computation of Python files to indicate the presence of `#TODO`, `#BUG`, or other specified tags in your python files. This tool helps maintain code quality by identifying and flagging these comments before committing changes.

## Features

- **Fast Parsing**: Efficiently scans Python files in your repository.
- **Customizable Tags**: Search for customizable tags like TODO, BUG, FIXME, etc.
- **Pre-Commit Hook**: Easily integrates with pre-commit or CI tools like GitHub Actions.
- **Fail Conditions**: Option to set a threshold for the number of tags found, failing the commit if exceeded.
- **TOML Configuration**: Configure the behavior of Comments Checker using a TOML file.

## Available Flags

The following flags can be used with the CLI to customize its behavior:

- `-dir`: Specifies the root directory to scan for Python files (default is current directory).
- `-tags`: Dash-separated list of tags to search for (default is "TODO-BUG-FIXME").
- `-mode`: Mode of operation, either "commit" or "root" (default is "commit").
- `-fail`: Fail the commit if the number of tags found exceeds this number (default is 0).
- `-config` : The path to the TOML configuration file to use. It will overide the over `-tags` and `-fail` flags if specified.

## TOML Configuration

To use a TOML configuration file with the `-config` flag, you need to provide a TOML file that includes the **tool.comche** table (a collection of key-value pairs) and the following keys:
- **tags**: an array of strings representing the list of tags to search for.
- **fail**: the maximum number of tags found before the commit is aborted.

Example:
```toml
[tool.comche]
tags = [
  "TODO",
  "FIXME",
  "BUG",
]
fail = 0
```

## Usage

To use Comments Checker as a pre-commit hook (with a TOML configuration file), add the following to your `.pre-commit-config.yaml` file:

```yaml
repos:
  - repo: https://gitlab.com/Adrien_RIAUX/comche
    rev: v0.1.2
    hooks:
      - id: comche
        args: [-config=<path_to_the_toml>]
```
Where *<path_to_the_toml>* refers to the location of your TOML file, such as *pyproject.toml*. Be sure to omit quotation marks around the path when specifying it.

Alternatively, you can run Comments Checker manually (you need to have go installed):

```bash
git clone https://gitlab.com/Adrien_RIAUX/comche
go run main.go -dir="./path/to/your/code" -tags="TODO-BUG-FIXME" -mode="commit" -fail=5
```

## Contributing

Contributions are welcome! Please fork the repository and submit pull requests. For any questions or issues, please open an issue in this repository.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
