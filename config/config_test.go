/*******************************************************************************
 * Copyright (c) 2025 Genome Research Ltd.
 *
 * Authors:
 *	- Sendu Bala <sb10@sanger.ac.uk>
 *
 * Permission is hereby granted, free of charge, to any person obtaining
 * a copy of this software and associated documentation files (the
 * "Software"), to deal in the Software without restriction, including
 * without limitation the rights to use, copy, modify, merge, publish,
 * distribute, sublicense, and/or sell copies of the Software, and to
 * permit persons to whom the Software is furnished to do so, subject to
 * the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 * EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 * MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
 * CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
 * TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 ******************************************************************************/

package config

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const filePerm = 0644

func TestConfig(t *testing.T) {
	Convey("Given a full set of env vars, you can make a config", t, func() {
		testPath := "/path"
		testSheetID := "sheetid"
		testUser := "user"
		testPass := "pass"
		testHost := "host"
		testPort := "1234"
		testDBName := "db"

		os.Setenv(EnvVarCreds, testPath)
		os.Setenv(EnvVarSheet, testSheetID)
		os.Setenv(EnvVarUser, testUser)
		os.Setenv(EnvVarPass, testPass)
		os.Setenv(EnvVarHost, testHost)
		os.Setenv(EnvVarPort, testPort)
		os.Setenv(EnvVarDBName, testDBName)

		config, err := FromEnv()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)
		So(config.CredentialsPath, ShouldEqual, testPath)
		So(config.SheetID, ShouldEqual, testSheetID)
		So(config.User, ShouldEqual, testUser)
		So(config.Password, ShouldEqual, testPass)
		So(config.Host, ShouldEqual, testHost)
		So(config.Port, ShouldEqual, testPort)
		So(config.DBName, ShouldEqual, testDBName)

		Convey("Without a full set of env vars, ConfigFromEnv fails", func() {
			os.Setenv(EnvVarUser, "")
			config, err := FromEnv()
			So(err, ShouldEqual, ErrMissingEnvs)
			So(config, ShouldBeNil)

			os.Setenv(EnvVarUser, "user")
			os.Setenv(EnvVarCreds, "")
			config, err = FromEnv()
			So(err, ShouldEqual, ErrMissingEnvs)
			So(config, ShouldBeNil)
		})

		Convey("You can load values from an .env file", func() {
			os.Unsetenv(EnvVarUser)

			origDir, err := os.Getwd()
			So(err, ShouldBeNil)

			defer func() {
				os.Chdir(origDir)
			}()

			dir := t.TempDir()
			err = os.Chdir(dir)
			So(err, ShouldBeNil)

			config, err := FromEnv()
			So(err, ShouldEqual, ErrMissingEnvs)
			So(config, ShouldBeNil)

			err = os.WriteFile(".env",
				[]byte(EnvVarUser+"=fileuser\n"+EnvVarDBName+"=filedb"), filePerm)
			So(err, ShouldBeNil)

			config, err = FromEnv()
			So(err, ShouldBeNil)
			So(config.User, ShouldEqual, "fileuser")
			So(config.CredentialsPath, ShouldEqual, testPath)
			So(config.DBName, ShouldEqual, testDBName)
		})
	})
}
