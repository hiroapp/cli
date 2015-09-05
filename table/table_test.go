package table

import (
	"strings"
	"testing"
)

func TestTable(t *testing.T) {
	{
		tbl := New().
			Add(String("1"), String("2"), String("3")).
			Add(String("1"), String("23"), String("456")).
			Add(String("7890"), String("1"), String("23"))
		want := strings.TrimLeft(`
1    2  3
1    23 456
7890 1  23
`, "\n")
		if got := tbl.String(); got != want {
			t.Errorf("got:\n%s\nwant:\n%s\n", got, want)
		}
	}
	{
		tbl := New().
			Add(String("1").Align(Right), String("2"), String("3")).
			Add(String("1"), String("23"), String("456")).
			Add(String("7890"), String("1").Align(Right), String("23").Align(Right))
		want := strings.TrimLeft(`
   1 2  3
1    23 456
7890  1  23
`, "\n")
		if got := tbl.String(); got != want {
			t.Errorf("got:\n%s\nwant:\n%s\n", got, want)
		}
	}
}
