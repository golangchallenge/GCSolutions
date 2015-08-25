# Go Challenge 5

Tool that finds unnecessarily exported objects and helps unexport them.

Options:
 -list: Shows a list of objects that can be unexported
 -uses: Shows a list of objects that are used in other packages and lists those packages
 -n: Shows filename and line number for listed objects
 -s: Use simple name instead of full object type string
 -e string: A list of names, separated by ',', that will not be unexported
 -i: Interactive mode
