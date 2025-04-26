![](./public/logo-transparent.png)

csvlang is a dynamic language to interact with CSV files. It is a domain-specific language that allows you to read, filter, modify, and export data from CSV files easily.

Documentation about csvlang commands can be found at [godoc.org](https://pkg.go.dev/github.com/rishabh570/csvlang).

## Contents

1. [Features](#features)
2. [Installation](#installation)
3. [Getting Started](#getting-started)
4. [Examples](#examples)
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

## Installation

<!-- TODO: add install instructions after figuring out the release process and making it go gettable -->

## Getting Started

1. Clone the project:

```bash
git clone https://github.com/rishabh570/csvlang
```

2. Change to the project directory:

```bash
cd csvlang
```

3. Run the project with your csvlang script path:

```bash
go run main.go --path <path-to-csvlang-script>
```

## Examples
<!-- 
```javascript
import Component from 'my-project'

function App() {
  return <Component />
}
``` -->

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
