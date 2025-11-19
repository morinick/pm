package main

import "flag"

func ParseFlags() map[string]string {
	keyFlag := flag.String("key", "", "key for decrypt backup")

	flag.Parse()

	return map[string]string{
		"key": *keyFlag,
	}
}
