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
	"encoding/json"
	"os"

	"golang.org/x/oauth2/jwt"
)

// ServiceCredentials holds the info in a service credentials JSON file from
// https://console.developers.google.com.
type ServiceCredentials struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

// ServiceCredentialsFromFile reads the given JSON file from (as retrieved from
// https://console.developers.google.com for a service account) and parses it
// in to a form usable by New().
func ServiceCredentialsFromFile(path string) (*ServiceCredentials, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	sc := &ServiceCredentials{}
	err = json.Unmarshal(data, sc)

	return sc, nil
}

func (sc *ServiceCredentials) toJWTConfig() *jwt.Config {
	return &jwt.Config{
		Email:        sc.ClientEmail,
		PrivateKey:   []byte(sc.PrivateKey),
		PrivateKeyID: sc.PrivateKeyID,
		TokenURL:     sc.TokenURI,
		Scopes: []string{
			"https://www.googleapis.com/auth/spreadsheets.readonly",
		},
	}
}
