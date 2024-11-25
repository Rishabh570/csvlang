# csvlang

## Demo

load input.csv

read (same as "read ALL" or "read all")
read head 10
read tail 5
read row 0 col age
read row 0
read col 0

let x = read row 0 col age

<!-- optional, later -->
<!-- read row 0:3 col 0:1
read row 0:3
read col 0:1 -->

let mydata = read row 0
append mydata
update row 0 = mydata

append "john,5,delhi"
append "joe,51,mumbai"

update row 0 col 0 "rishabh"
update row 0 "rishabh,1,blr"

delete row 1


## V1 scope

1. `load i.csv`
2. Read ops: `read` / `read row <i>` / `read row <i> col <j>` / `read col <i>`
3. Write ops: `append "john,23"` / `append <newRow>`

<!-- [1/3] Read all as 2d arr -->
read all

<!-- [2/3] Read as 1d arr -->
read row <ind>
read col <ind>

<!-- [3/3] Read specific value -->
read row <i> col <j>

<!-- [1/2] Append -->
let newRow = "john,23"
append newRow

<!-- [2/2] Append -->
let data = [["john",23],["joe",21]]
<!-- Note: unused i is not flagged -->
<!-- should flag if incompatible data -->
for i, row in data {
  append row
}

<!-- Filter row and cols -->

<!-- a built-in fn that removes null/empty fields -->

<!-- export keyword to export the csv as json or csv -->

<!-- built-ins for number and []number: add(), sub(), mul(), div(), avg(), mean([]), median([]) -->

<!-- builts-in for string and []string: len(), isEmpty()

<!-- csvObj.sort(orderByCol, asc/des) -->

<!-- csvObj.removeDuplicateRows() -->

<!-- [1/1] Delete -->
delete row 0

<!-- [1/3] assigns to x as 1d arr -->
<!-- Note: right-hand-side of equal sign will be an expression; we store evaluated expr in the var -->
let x = read row 0
let y = read col 0

<!-- [2/3] assigns to x as 2d arr -->
let x = read all

<!-- [3/3] assigns to x as a specific value -->
let x = read row 0 col 0

<!-- [1/2] Comment: single line -->
# this is a comment

<!-- [2/2] Comment: multi line -->
##
this is a multi line comment
##

<!-- [1/1] Conditional -->
<!-- Need to be mindful of closures -->
if <someExpression> {
  # do something
} else if <someOtherExpression> {
  # do something else
} else {
  # do something else now
}


<!-- FOR LOOPS -->
<!-- Need to be mindful of closures -->

<!-- [1/3] loop over 1d array -->
let myRow = read row 0
for ind, colAtInd in myRow {
  <!-- It'd be nice to have data type validation during assignment -->
  colAtInd = "thisValueDoesNotPersist"
  myRow[ind] = "thisDoes"
}

let myCol = read col 0
for ind, RowAtInd in myCol {
  <!-- It'd be nice to have data type validation during assignment -->
  RowAtInd = "thisValueDoesNotPersist"
  myCol[ind] = "thisValueDoes"
}


let allRows = read all
for i, row in allRows {
  for j, col in row {
    if col == "someValueToCheckAgainst" {
      allRows[i][j] = "conditionallyUpdatedValue"
    }

    if col == 42 {
      allRows[i][j] = 2 * col
    }
  }
}


<!-- Error handling at parsing stage -->
1. should show stack traces


<!-- EXPORT -->
1. should allow export as json, csv, txt



## Notes

- overwriting value of a globally defined var won't effect its value outside the fn. but the same is possible with an if-condition
- valid read ops
  - readall
  - read row <ind>;
  - read row <rInd> col <cInd>
  - read row <rInd> col <cInd>;