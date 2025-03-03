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

package sheets

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/wtsi-hgi/dimsum-automation/config"
	"golang.org/x/oauth2/jwt"
)

const (
	userPerms       = 0600
	credentialsJSON = `{
    "type": "service_account",
    "project_id": "projectID",
    "private_key_id": "keyID",
    "private_key": "keyContent\n",
    "client_email": "user@project.iam.gserviceaccount.com",
    "client_id": "12345",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/project.iam.gserviceaccount.com",
    "universe_domain": "googleapis.com"
}`
)

func TestConfig(t *testing.T) {
	Convey("Given a credentials file, you can generate a ServiceCredentials, which can convert to a jwt.Config", t, func() {
		dir := t.TempDir()
		credPath := filepath.Join(dir, "credentials.json")
		err := os.WriteFile(credPath, []byte(credentialsJSON), userPerms)
		So(err, ShouldBeNil)

		sc, err := ServiceCredentialsFromFile(credPath)
		So(err, ShouldBeNil)
		So(sc, ShouldResemble, &ServiceCredentials{
			Type:                    "service_account",
			ProjectID:               "projectID",
			PrivateKeyID:            "keyID",
			PrivateKey:              "keyContent\n",
			ClientEmail:             "user@project.iam.gserviceaccount.com",
			ClientID:                "12345",
			AuthURI:                 "https://accounts.google.com/o/oauth2/auth",
			TokenURI:                "https://oauth2.googleapis.com/token",
			AuthProviderX509CertURL: "https://www.googleapis.com/oauth2/v1/certs",
			ClientX509CertURL:       "https://www.googleapis.com/robot/v1/metadata/x509/project.iam.gserviceaccount.com",
		})

		c := sc.toJWTConfig()
		So(c, ShouldResemble, &jwt.Config{
			Email:        "user@project.iam.gserviceaccount.com",
			PrivateKey:   []byte("keyContent\n"),
			PrivateKeyID: "keyID",
			TokenURL:     "https://oauth2.googleapis.com/token",
			Scopes: []string{
				"https://www.googleapis.com/auth/spreadsheets.readonly",
			},
		})

		Convey("You can make a ServiceCredentials from a Config", func() {
			c := &config.Config{
				CredentialsPath: credPath,
				SheetID:         "sheetID",
			}

			sc2, err := ServiceCredentialsFromConfig(c)
			So(err, ShouldBeNil)
			So(sc2, ShouldResemble, sc)
		})
	})
}
