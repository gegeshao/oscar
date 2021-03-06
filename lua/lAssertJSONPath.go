package lua

import (
	"fmt"
	"github.com/mono83/oscar/util/jsonpath"
	"github.com/yuin/gopher-lua"
)

func lAssertJSONPath(L *lua.LState) int {
	tc := lContext(L)
	xpath := tc.Interpolate(L.ToString(2))
	expected := tc.Interpolate(L.ToString(3))
	doc := L.OptString(4, "")

	// Reading response body
	body := tc.Get("http.response.body")

	tc.Tracef(`Reading JSON XPath "%s"`, xpath)

	// Extracting json path
	actual, err := jsonpath.Extract([]byte(body), xpath)
	if err != nil {
		L.RaiseError(err.Error())
	} else {
		tc.Tracef(`Assert "%s" (actual, left) equals "%s"`, actual, expected)
		success := actual == expected
		if !success {
			err := fmt.Errorf(
				`JSON XPath assertion failed. "%s" (actual, left) != "%s".%s`,
				actual,
				expected,
				doc,
			)
			L.RaiseError(err.Error())
		} else {
			tc.AssertFinished(nil)
			tc.Tracef(`Assertion OK. "%s" == "%s"`, xpath, expected)
		}
	}

	return 0
}
