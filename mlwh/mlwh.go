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

package mlwh

import (
	"database/sql"
	"time"

	"github.com/go-sql-driver/mysql"
)

const (
	sqlDriverName   = "mysql"
	connMaxLifetime = time.Minute * 3
	maxOpenConns    = 10
	maxIdleConns    = 10
)

// MLWH is a connection to the MLWH database.
type MLWH struct {
	pool *sql.DB
}

// New returns a new MLWH connection using mysql.Config that you can get from
// MySQLConfigFromConfig(config.FromEnv()).
func New(c *mysql.Config) (*MLWH, error) {
	pool, err := sql.Open(sqlDriverName, c.FormatDSN())
	if err != nil {
		return nil, err
	}

	pool.SetConnMaxLifetime(connMaxLifetime)
	pool.SetMaxOpenConns(maxOpenConns)
	pool.SetMaxIdleConns(maxIdleConns)

	return &MLWH{pool: pool}, pool.Ping()
}

// Sample represents a sample in the MLWH, including study and run information.
type Sample struct {
	StudyID    string
	StudyName  string
	RunID      string
	SampleID   string
	SampleName string
	ManualQC   bool
}

const getSamples = `
SELECT DISTINCT st.id_study_lims as StudyID, st.name as StudyName,
r.id_run as RunID, sa.sanger_sample_id as SangerSampleID,
sa.supplier_name as SupplierName, fc.manual_qc as ManualQC
FROM iseq_flowcell fc
JOIN study st on st.id_study_tmp = fc.id_study_tmp
JOIN iseq_run r on r.id_flowcell_lims = fc.id_flowcell_lims
JOIN sample sa on sa.id_sample_tmp = fc.id_sample_tmp
WHERE st.faculty_sponsor = ? and (fc.manual_qc = '1' or fc.manual_qc = '0')
`

// SamplesForSponsor returns all samples in the MLWH for the given sponsor.
func (m *MLWH) SamplesForSponsor(sponsor string) ([]Sample, error) {
	rows, err := m.pool.Query(getSamples, sponsor)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var samples []Sample

	for rows.Next() {
		var sample Sample

		if err := rows.Scan(
			&sample.StudyID,
			&sample.StudyName,
			&sample.RunID,
			&sample.SampleID,
			&sample.SampleName,
			&sample.ManualQC,
		); err != nil {
			return nil, err
		}

		samples = append(samples, sample)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return samples, nil
}

// Close closes the connection to the MLWH.
func (m *MLWH) Close() error {
	return m.pool.Close()
}
