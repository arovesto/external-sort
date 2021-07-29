# external-sort

---

### Build and run

Run: `cd cmd/sorter && go build && ./sorter`

sorter require creating temporary files, by default it is left to OS to decide where, can be overwritten

you can specify how many lines should be in memory at a time

Run test generator: `cd cmd/generator && go build && ./generator`


### Algorithm

let n be how many lines in file
let k be how many lines can be stored in memory

1. Divide file in k-sized portions
2. Take k portions and merge them into bigger file (maintain only k lines in memory)
3. Do it until have a single file

* we can't just merge into result on step 2 because n / k (how many files we have) might be greater then k (max lines in memory allowed)

