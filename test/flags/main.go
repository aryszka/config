package main

import (
	"flag"
	"fmt"
)

func main() {
	/*
		-foo
		--foo
		-foo=1
		-foo = 1
		-foo 1
		-foo 1 2 3 -bar 4
	*/

	var (
		foo bool
		bar string
	)

	flag.BoolVar(&foo, "foo", false, "just foo")
	flag.StringVar(&bar, "bar", "", "just bar")
	flag.Parse()
	fmt.Println(foo, bar, flag.Args())
}
