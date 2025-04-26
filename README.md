![](./public/logo-transparent.png)

[![Tests](https://github.com/rishabh570/csvlang/actions/workflows/unit-tests.yml/badge.svg)](https://github.com/rishabh570/csvlang/actions/workflows/unit-tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/rishabh570/csvlang)](https://goreportcard.com/report/github.com/rishabh570/csvlang)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/rishabh570/csvlang.svg)](https://pkg.go.dev/github.com/rishabh570/csvlang)

csvlang is a dynamic language to interact with CSV files. It is a domain-specific language that allows you to read, filter, modify, and export data from CSV files easily.

Documentation about csvlang commands can be found at [godoc.org](https://pkg.go.dev/github.com/rishabh570/csvlang).

## Contents

1. [Features](#features)
2. [Usage](#usage)
3. [Getting Started](#getting-started)
4. [Documentation](#documentation)
5. [Running Tests](#running-tests)
6. [Contributing](#contributing)
7. [License](#license)

## Features

### Read all or specific rows or columns

```
load data.csv

let rows = read row *;
let firstRow = read row 0;
let firstColumn = read row * col 0;
let filteredRows = read row * col * where age > 20;
```

### Fill empty cells with a fallback value

```
load data.csv

let rows = read row * col amount where age > 20;
let polyfilledRows = fill(rows, "name", "John Doe");
```

### Remove duplicate rows

```
load data.csv

let rows = read row *;

let uniqueRows = unique(rows);
```

### Built-in statistical functions 

To calculate the sum, average, and count of values in a column.

```
load data.csv

let rows = read row * col amount where age > 20;

let totalAmount = sum(rows);
let averageAmount = avg(rows);
let countAmount = count(rows);
```


### Export to JSON or CSV file

```
load data.csv

let rows = read row * where age > 20;

save rows as output.csv;
save rows as output.json;
```

## Usage

### Option 1: Using Go (Recommended)

```bash
go install github.com/rishabh570/csvlang@latest
```

### Option 2: Binary Installation

1. Download the latest binary for your platform from the [releases page](https://github.com/rishabh570/csvlang/releases).
2. Extract the archive.
3. Add the binary to your PATH.

### Option 3: Build from Source

```bash
git clone https://github.com/rishabh570/csvlang.git
cd csvlang
go build -o csvlang
```


## Getting Started

1. Clone the project:

```bash
git clone https://github.com/rishabh570/csvlang
```

2. Change to the project directory:

```bash
cd csvlang
```

3. Install the dependencies:

```bash
go mod tidy
```

4. Create a CSV file named `data.csv` in the project directory with the following content:

```csv
name,age,amount
John Doe,25,1000
John Doe,25,1000
Jane Smith,30,2000
Bob Brown,28,2500
```

5. Create a csvlang script named `script.csvlang` in the project directory with the following content:

```csvlang
load data.csv
let rows = read row * where age > 20;
let uniqueRows = unique(rows);
save uniqueRows as output.csv;
```

6. Run the project with your csvlang script path:

```bash
go run main.go --path <path-to-csvlang-script>
```

or, if you have installed the binary and it is present in your PATH, you can run the following command:

```bash
csvlang --path <path-to-csvlang-script>
```

## Documentation

To learn more, check out the [documentation](https://pkg.go.dev/github.com/rishabh570/csvlang).

## Running Tests

To run tests, run the following command

```bash
go test ./...
```

## Contributing

Contributions are always welcome!

See `contributing.md` for ways to get started.

## License

[MIT](https://choosealicense.com/licenses/mit/)
