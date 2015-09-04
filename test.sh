#!/bin/bash
base=$(pwd); find * -type d | while read dir ; do 
	cd $dir
      	ls *.go 2> /dev/null > /dev/null && {	
		echo "DIR: $dir"
		ls .test_disabled || {
			ls | grep -v "_test.go" | grep ".go" 2> /dev/null > /dev/null && {
				go install 2> /dev/null > /dev/null || echo -e "\e[1;31m BUILD FAIL: $dir \e[0m"
			}
			go vet || echo -e "\e[1;31m VET FAIL: $dir \e[0m" 
			golint || echo -e "\e[1;31m LINT FAIL: $dir \e[0m"  
			go test -cover $1 || echo -e "\e[1;31m TEST FAIL: $dir \e[0m"
		}	
	}
	cd $base 
done 2> /dev/null
