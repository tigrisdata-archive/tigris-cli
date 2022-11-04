// Copyright 2022 Tigris Data, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package templates

import (
	"embed"
)

var (
	//go:embed login/success.gohtml
	LoginSuccessful string
	//go:embed login/error.gohtml
	LoginError string

	//go:embed scaffold/go/base
	//go:embed scaffold/go/base/**/*
	//go:embed scaffold/go/base/.*
	ScaffoldGoBase embed.FS

	//go:embed scaffold/go/gin
	//go:embed scaffold/go/gin/**/*
	ScaffoldGoGin embed.FS

	//go:embed scaffold/typescript/base
	//go:embed scaffold/typescript/base/src/**/*
	ScaffoldTypeScriptBase embed.FS

	//go:embed scaffold/typescript/express
	//go:embed scaffold/typescript/express/src/**/*
	//go:embed scaffold/typescript/express/.*
	ScaffoldTypeScriptExpress embed.FS

	//go:embed scaffold/java/spring
	//go:embed scaffold/java/spring/src/main/java/_java_pkg_
	//go:embed scaffold/java/spring/src/main/java/_java_pkg_/_*
	//go:embed scaffold/java/spring/src/main/java/_java_pkg_/**/*
	//go:embed scaffold/java/spring/.gitignore
	ScaffoldJavaSpring embed.FS

	//go:embed schema
	Schema embed.FS
)
