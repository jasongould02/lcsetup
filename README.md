# lcsetup

Made to make the files and folders to organize my Leetcode solutions.

```
go build lcsetup.go
```

In order to use this program from any location you can add the following to your .profile 
```
export PATH=$PATH:[PATH_TO_COMPILED_CODE]
```

## Usage

```
./lcsetup [titleSlug] [extensions]
```

## Example

https://leetcode.com/problems/isomorphic-strings/
https://leetcode.com/problems/[titleSlug]/

```
./lcsetup isomomorphic-strings java
```
