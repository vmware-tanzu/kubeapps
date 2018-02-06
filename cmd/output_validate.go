package cmd

import (
	"fmt"
)

var report_fmt = `----------------------------------------
		test pass: %d
		test fail: %d

		Overall: %s
----------------------------------------
`
var verbose bool

//TODO use logger
func KInfo(text ...interface{}) {
	if verbose {
		fmt.Print("INF:")
		fmt.Println(text...)
	}
}

func KPass(text ...interface{}) {
	fmt.Print("## PASS:")
	fmt.Println(text...)
	passcount++
}
func KFail(text ...interface{}) {
	fmt.Print("## FAIL:")
	fmt.Println(text...)
	kpass_val = false
	passfail++
}
func Report() bool {
	overall := "PASS"
	if !kpass_val {
		overall = "FAIL"
	}
	fmt.Printf(report_fmt, passcount, passfail, overall)
	return kpass_val
}
