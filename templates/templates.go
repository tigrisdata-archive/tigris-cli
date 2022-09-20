package templates

import _ "embed"

var (
	//go:embed login/success.gohtml
	LoginSuccessful string
	//go:embed login/error.gohtml
	LoginError string
)
