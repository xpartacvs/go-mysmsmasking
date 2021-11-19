package sms

import (
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

func Test_Client_Instantiation_Return_Correct_Type(t *testing.T) {
	require.IsType(t, Client{}, NewClient("", ""))
}
