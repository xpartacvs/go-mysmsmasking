package sms

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

/*
func Test_runtimeBaseURL_Return_Env_Value(t *testing.T) {
	require.Equal(t, "http://xxx.yyy", runtimeBaseURL())
}
*/

func Test_runtimeBaseURL_Return_Default_Value(t *testing.T) {
	require.Equal(t, defaultBaseUrl, runtimeBaseURL())
}

func Test_callApi_With_Wrong_Method_Return_Error(t *testing.T) {
	c := NewClient("", "")
	_, e := c.callApi("PUT", "", "", url.Values{})
	require.Error(t, e)
}

func Test_callApi_With_Wrong_Method_Return_Error_Not_Allowed_Method(t *testing.T) {
	c := NewClient("", "")
	_, e := c.callApi("PUT", "", "", url.Values{})
	require.ErrorIs(t, e, ErrMethodNotAllowed)
}

func Test_callApi_With_GET_Not_Error(t *testing.T) {
	c := NewClient("", "")
	_, e := c.callApi("GET", "", "", url.Values{})
	require.NoError(t, e)
}

func Test_callApi_With_methodGet_Not_Error(t *testing.T) {
	c := NewClient("", "")
	_, e := c.callApi(methodGet, "", "", url.Values{})
	require.NoError(t, e)
}

func Test_callApi_With_POST_Not_Error(t *testing.T) {
	c := NewClient("", "")
	_, e := c.callApi("POST", "", "", url.Values{})
	require.NoError(t, e)
}

func Test_callApi_With_methodPost_Not_Error(t *testing.T) {
	c := NewClient("", "")
	_, e := c.callApi(methodPost, "", "", url.Values{})
	require.NoError(t, e)
}

func Test_Client_Instantiation_Return_Correct_Type(t *testing.T) {
	require.IsType(t, Client{}, NewClient("", ""))
}
