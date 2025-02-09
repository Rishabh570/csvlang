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


update row 0 col 0 "rishabh"
update row 0 "rishabh,1,blr"

delete row 1


## V1 scope

1. `load i.csv`
2. Read ops: `read` / `read row <i>` / `read row <i> col <j>` / `read col <i>`

<!-- [1/3] Read all as 2d arr -->
read all

<!-- [2/3] Read as 1d arr -->
read row <ind>
read col <ind>

<!-- [3/3] Read specific value -->
read row <i> col <j>


<!-- [DONE] 1d/2d arr support -->
<!-- access elements at indexes, len, concat? -->

<!-- [DONE] read rows,cols,all + filter on col values -->
read row 0;
read row 0 where age > 18;   <!-- should be invalid -->
read row 0 col age;
read row *;
read row * where age > 18;
read row * col name;
read row * col name where age > 18;

<!-- for loop -->

<!-- built-in fn: -->
<!-- 1. removes null/empty fields -->
<!-- 2. fills null/empty fields -->
<!-- 3. removes exact duplicate rows -->

<!-- [DONE] "save" keyword to save the csv as json or csv -->
<!--1. save -->
<!--2. save as filtered.(csv|json)  -->
<!-- 3. save rows -->
<!-- 4. save rows as filtered.(csv|json) -->

<!-- built-in fns for number and []number: sum([]int), avg([]int) -->

<!-- built-in fns []string: len() -->

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




## Notes

- overwriting value of a globally defined var won't effect its value outside the fn. but the same is possible with an if-condition
- valid read ops
  - let val = read row 0;
  - let val = read row 0 where age > 12;
  - let val = read row 0 col age;
  - let val = read row *;
  - let val = read row * where age > 12;
  - let val = read row * col name;
  - let val = read row * col name where age > 12;


## Todo

- [x] For loop
- [x] fix data type issues and fill empty/incorrect valconsole.log();
- [x] Unit tests for all the features
- [x] gh workflow for tests
- [x] godoc
- [x] readme docs
- [ ] address remaining TODO comments
- [ ] change package structure
- [ ] remove main.go and build as a library?
- [ ] makefile
- [ ] goreleaser
- [ ] license
- [x] contributing.md
- [ ] Set up branch protection rules for main branch?
- [ ] Configure required status checks
- [x] .gitignore
