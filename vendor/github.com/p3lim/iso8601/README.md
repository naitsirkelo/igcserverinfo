# iso8601

[![GoDoc](https://godoc.org/github.com/p3lim/iso8601?status.svg)](https://godoc.org/github.com/p3lim/iso8601)

iso8601 is a simple Go package for formatting `time.Duration` according to the [ISO 8601 format](https://en.wikipedia.org/wiki/ISO_8601#Durations).

```go
package main

import (
	"fmt"
	"time"

	"github.com/p3lim/iso8601"
)

func main() {
	t := time.Now()
	time.Sleep(5 * time.Second)
	fmt.Println(iso8601.Format(time.Since(t)))
}
```
