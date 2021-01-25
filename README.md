# go-cgroups

A example for Memory/CPU limit with Linux CGgroup in Go.

### How to test

go get github.com/piaohua/go-cgroups

- build `main`

```
$ GO111MODULE=off go build main.go

``` 

- run command `./main -name=simple_app -exec=./simple_app`

```
$ make app

$ ./main -name=app -exec=./simple_app
```


### Features

- blkio limit
- devices limit
- net limit
- freezer
- process restart


### Reference

- [man7/cgroups](http://man7.org/linux/man-pages/man7/cgroups.7.html)
- [限制cgroup的内存使用](https://segmentfault.com/a/1190000008125359)
- [mydocker](https://github.com/piaohua/mydocker)
- [box](https://github.com/piaohua/box)
