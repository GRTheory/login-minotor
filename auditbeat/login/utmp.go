package login

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/elastic/beats/v7/auditbeat/datastore"
	"github.com/elastic/elastic-agent-libs/logp"
)

const (
	bucketKeyFileRecords   = "file_records"
	bucketKeyLoginSessions = "login_sessions"
)

// Inode represents a file's inode on Linux.
type Inode uint64

// UtmpType represents the type of a UTMP file and records.
// Two types are possible: wtmp (records from the "good" file, i.e. /var/log/wtmp)
// and btmp (failed logins from /var/log/btmp).
type UtmpType uint8

const (
	// Wtmp is the "normal" wtmp file that includes successful logins, logouts,
	// adn system boots.
	Wtmp UtmpType = iota
	// Btmp contains bad logins only.
	Btmp
)

// UtmpFile represents a UTMP file at a point in time.
type UtmpFile struct {
	Inode  Inode
	Path   string
	Size   int64
	Offset int64
	Type   UtmpType
}

// UtmpFileReader can read a UTMP formatted file (usually /var/log/wtmp).
type UtmpFileReader struct {
	log            *logp.Logger
	bucket         datastore.Bucket
	config         config
	savedUtmpFiles map[Inode]UtmpFile
	loginSession   map[string]LoginRecord
}

// NewUtmpFileReader creates and initializes a new UTMP file reader.
func NewUtmpFileReader(log *logp.Logger, bucket datastore.Bucket, config config) (*UtmpFileReader, error) {
	// r := &UtmpFileReader{
	// 	log:            log,
	// 	bucket:         bucket,
	// 	config:         config,
	// 	savedUtmpFiles: make(map[Inode]UtmpFile),
	// 	loginSession:   make(map[string]LoginRecord),
	// }

	// Load state (fiel records, tty mapping) from disk.
	// err := r.r

	return nil, nil
}

func (r *UtmpFileReader) restoreStateFromDisk() error {
	return nil
}

func (r *UtmpFileReader) restoreLoginSessionsFromDisk() error {
	var decoder *gob.Decoder
	err := r.bucket.Load(bucketKeyLoginSessions, func(blob []byte) error {
		if len(blob) > 0 {
			buf := bytes.NewBuffer(blob)
			decoder = gob.NewDecoder(buf)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if decoder != nil {
		for {
			loginRecord := new(LoginRecord)
			err = decoder.Decode(loginRecord)
			if err == nil {
				r.loginSession[loginRecord.TTY] = *loginRecord
			}else if err == io.EOF {
				// Read all
				break
			}else {
				return fmt.Errorf("error decodng login record: %w", err)
			}
		}
	}
	r.log.Debugf("Restored %d open login sessions from disk", len(r.loginSession))

	return nil
}
