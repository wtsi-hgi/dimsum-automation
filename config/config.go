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

	"github.com/joho/godotenv"
)

const (
	EnvVarCreds  = "DIMSUM_AUTOMATION_CREDENTIALS_FILE"
	EnvVarSheet  = "DIMSUM_AUTOMATION_SPREADSHEET_ID"
	EnvVarUser   = "DIMSUM_AUTOMATION_SQL_USER"
	EnvVarPass   = "DIMSUM_AUTOMATION_SQL_PASS"
	EnvVarHost   = "DIMSUM_AUTOMATION_SQL_HOST"
	EnvVarPort   = "DIMSUM_AUTOMATION_SQL_PORT"
	EnvVarDBName = "DIMSUM_AUTOMATION_SQL_DB"

	sqlNetwork = "tcp"
)

type Error string

func (e Error) Error() string { return string(e) }

const ErrMissingEnvs = Error("missing required environment variables")

type Config struct {
	CredentialsPath string
	SheetID         string
	User            string
	Password        string
	Host            string
	Port            string
	DBName          string
}

// FromEnv returns a new Config with properies populated from environment
// variables DIMSUM_AUTOMATION_*, where * is amongst: CREDENTIALS_FILE,
// SPREADSHEET_ID, SQL_USER, SQL_PASS, SQL_HOST, SQL_PORT, and TSQL_DB.
//
// If these environment variables are defined in a file called .env (and not
// previously set in an environment variable), they will be automatically
// loaded.
//
// Optionally supply a directory to look for the .env file in.
func FromEnv(dir ...string) (*Config, error) {
	var parentDir string
	if len(dir) == 1 {
		parentDir = dir[0] + string(os.PathSeparator)
	}

	godotenv.Load(parentDir + ".env")

	cred := os.Getenv(EnvVarCreds)
	sheet := os.Getenv(EnvVarSheet)
	user := os.Getenv(EnvVarUser)
	pass := os.Getenv(EnvVarPass)
	host := os.Getenv(EnvVarHost)
	port := os.Getenv(EnvVarPort)
	dbname := os.Getenv(EnvVarDBName)

	if cred == "" || sheet == "" || user == "" || pass == "" || host == "" || port == "" || dbname == "" {
		return nil, ErrMissingEnvs
	}

	return &Config{
		CredentialsPath: cred,
		SheetID:         sheet,
		User:            user,
		Password:        pass,
		Host:            host,
		Port:            port,
		DBName:          dbname,
	}, nil
}
