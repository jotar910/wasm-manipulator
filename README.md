# WasmManipulator Tool

## Overview
WasmManipulator is a Golang-based tool designed to work with WebAssembly (WASM) modules. It facilitates the transformation and manipulation of WASM modules, allowing users to apply specific 'advices' to their WASM code. This tool is particularly useful for developers looking to optimize, modify, or analyze WASM modules in a flexible and efficient manner.

## Documentation
For more detailed information and usage instructions, see the [WasmManipulator Documentation](./DOCS).

## Features
- Modify and transform WASM modules (.wasm or .wat format)
- Apply a set of 'advices' to the transformation process
- Support for both command line arguments and environment variables
- Generation of auxiliary JavaScript files for complex types or runtime expressions
- Detailed configuration options for precise control over the transformation process

## Installation
To install WasmManipulator, clone the repository and build the tool using Go:

```bash
git clone https://github.com/your-repo/WasmManipulator.git
cd WasmManipulator
go build -o wmr
```

## Usage
To use WasmManipulator, you can either set environment variables or pass parameters directly when executing the tool. Below is an example command:

```bash
WMR_IN_MODULE="module.wasm" ./wmr --include=advice_1,advice_2
```

Alternatively, use command-line parameters:

```bash
./wmr --data_dir="$HOME/" --in_module="data/module.wasm"
```

### Configuration Options
WasmManipulator can be configured using the following options:

| Environment Variable          | Parameter         | Type       | Default    |
| ----------------------------- | ----------------- | ---------- | ---------- |
| WMR_IN_MODULE                 | in_module         | string     | input.wasm |
| WMR_IN_TRANSFORM              | in_transform      | string     | input.yml  |
| WMR_OUT_MODULE                | out_module        | string     | output.wasm|
| WMR_OUT_JS                    | out_js            | string     | output.js  |
| WMR_OUT_MODULE_ORIG           | out_module_orig   | string     | null       |
| WMR_DEPENDENCIES_DIR          | dependencies_dir  | string     | ./dependencies/ |
| WMR_DATA_DIR                  | data_dir          | string     | ./         |
| WMR_LOG_FILE                  | log_file          | string     | null       |
| WMR_INCLUDE                   | include           | string[]   | All        |
| WMR_EXCLUDE                   | exclude           | string[]   | None       |
| WMR_PRINT_JS                  | print_js          | boolean    | false      |
| WMR_ALLOW_EMPTY               | allow_empty       | boolean    | false      |
| WMR_VERBOSE                   | verbose           | boolean    | false      |
| WMR_IGNORE_ORDER              | ignore_order      | boolean    | false      |

*(Refer to the provided documentation for a complete list of configurations.)*

## WasmManipulator Language Specification
WasmManipulator utilizes a YAML-based language for WASM transformation. This language offers a variety of fields for defining the transformation process, such as `Pointcuts`, `Aspects`, `Advices`, and more.

### Syntax
The syntax includes types like `Object`, `Array`, `String`, `Identifier`, `Type`, and others. Each has specific roles in the transformation definition.

### Example Structure
```yaml
{
  Pointcuts: Map,
  Aspects: Map,
  Context: {
    Variables: Map,
    Functions: {
      Variables: Map,
      Args: Array<{
        Name: string,
        Type: string,
      }>,
      Result: string,
      Code: string,
      Imported: {
        Module: string,
        Field: string,
      },
      Exported: string,
    },
  },
  Advices: {
    Pointcut: string,
    Variables: Map,
    Advice: string,
    Order: i32,
    All: boolean,
    Smart: boolean,
  },
  Start: string,
  Templates: Map,
}
```
*(Refer to the detailed documentation for an in-depth understanding of each field and type.)*

## Author
WasmManipulator was created and is maintained by João Rodrigues. For inquiries or suggestions, you can reach him via email at [joaordev@gmail.com](mailto:joaordev@gmail.com).

## Contributing
Contributions to WasmManipulator are welcome. Please ensure that your contributions adhere to the coding standards and include appropriate tests. To contribute:

1. Fork the repository.
2. Create a new branch for each feature or improvement.
3. Submit a pull request with a clear description of the changes.

## License
WasmManipulator is licensed under the Creative Commons Attribution-ShareAlike 4.0 International License (CC BY-SA 4.0). For more details, see the [LICENSE](./LICENSE) file.

## Support
For support and further inquiries, please open an issue in the GitHub repository or contact the maintainers directly.

## Acknowledgments
Special thanks to all contributors and supporters of the WasmManipulator project.
