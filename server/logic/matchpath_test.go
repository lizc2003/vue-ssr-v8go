package logic

import (
	"fmt"
	"testing"
)

func TestMatchPath(t *testing.T) {
	testCases := []struct {
		pattern string
		path    string
		expect  bool
	}{
		// basic test
		{"*.go", "main.go", true},
		{"*.go", "test.java", false},
		{"test?.go", "test1.go", true},
		{"test?.go", "test.go", false},
		{"ab?c*d", "abYcZd", true},
		{"*", "/", false},

		// path test
		{"/src/", "/src", false},
		{"/src", "/src/main.go", true},
		{"/srx", "/src/main.go", false},
		{"/src/*.go", "/src/main.go", true},
		{"/*/src", "/src", false},
		{"src/*.go", "lib/main.go", false},
		{"/*/test", "/lib/testX", true},

		// ** test
		{"src/**/*.go", "src/main.go", false},
		{"src/**/*.go", "src/a/b/c/main.go", true},
		{"**/*.go", "any/deep/path/main.go", true},

		// mixed test
		{"src/**/test_?.go", "src/a/b/test_1.go", true},
		{"**/*test*", "a/b/c/my_test.go", true},

		// boundary test
		{"", "", true},
		{"*", "", true},
		{"?", "", false},
		{"**", "", true},
	}

	bFailed := false
	fmt.Println("test:")
	for _, tc := range testCases {
		result := MatchPath(tc.path, tc.pattern)
		status := "✓"
		if result != tc.expect {
			bFailed = true
			status = "✗"
		}
		fmt.Printf("%s pattern: %-20s path: %-30s result: %v\n",
			status, tc.pattern, tc.path, result)
	}

	if bFailed {
		t.Fail()
	}
}
