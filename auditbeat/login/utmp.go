package login

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"syscall"

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
	loginSessions  map[string]LoginRecord
}

// NewUtmpFileReader creates and initializes a new UTMP file reader.
func NewUtmpFileReader(log *logp.Logger, bucket datastore.Bucket, config config) (*UtmpFileReader, error) {
	r := &UtmpFileReader{
		log:            log,
		bucket:         bucket,
		config:         config,
		savedUtmpFiles: make(map[Inode]UtmpFile),
		loginSessions:  make(map[string]LoginRecord),
	}

	// Load state (fiel records, tty mapping) from disk.
	err := r.restoreStateFromDisk()
	if err != nil {
		return nil, fmt.Errorf("failed to restore state from disk: %w", err)
	}

	return r, nil
}

// Close performs any cleanup tasks when the UTMP reader is done.
func (r *UtmpFileReader) Close() error {
	if r.bucket != nil {
		return r.bucket.Close()
	}
	return nil
}

// ReadNew returns any new UTMP entries in any files matching the configured pattern.
func (r *UtmpFileReader) ReadNew() (<-chan LoginRecord, <-chan error) {
	loginRecordC := make(chan LoginRecord)
	errorC := make(chan error)

	go func() {
		defer logp.Recover("A panic occured while collecting login information")
		defer close(loginRecordC)
		defer close(errorC)

		// wtmpFiles, err := r.find
	}()

	return loginRecordC, errorC
}

func (r *UtmpFileReader) findFiles(filePattern string, utmpType UtmpType) ([]UtmpFile, error) {
	paths, err := filepath.Glob(filePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to expand file pattern %v: %w", filePattern, err)
	}

	// Sort paths in reverse order (oldest/most-rorated file first)
	sort.Sort(sort.Reverse(sort.StringSlice(paths)))

	var utmpFiles []UtmpFile
	for _, path := range paths {
		fileInfo, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				// Skip - file might have been rotated out
				r.log.Debugf("File %v does not exist anymore.", path)
				continue
			} else {
				return nil, fmt.Errorf("unexpected error when looking up file %v: %w", path, err)
			}
		}

		utmpFiles = append(utmpFiles, UtmpFile{
			Inode: Inode(fileInfo.Sys().(*syscall.Stat_t).Ino),
			Path:  path,
			Size:  fileInfo.Size(),
			Type:  utmpType,
		})
	}

	return utmpFiles, nil
}

// deleteOldUtmpFiles cleans up old UTMP file records where the inode no longer exists.
func (r *UtmpFileReader) deleteOldUtmpFiles(existingFiles *[]UtmpFile) {
	existingInodes := make(map[Inode]struct{})
	for _, utmpFile := range *existingFiles {
		existingInodes[utmpFile.Inode] = struct{}{}
	}

	
}

func (r *UtmpFileReader) saveStateToDisk() error {
	err := r.saveFileRecordsToDisk()
	if err != nil {
		return err
	}
	err = r.saveLoginSessionsToDisk()
	if err != nil {
		return err
	}

	return nil
}

func (r *UtmpFileReader) saveFileRecordsToDisk() error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	for _, utmpFile := range r.savedUtmpFiles {
		err := encoder.Encode(utmpFile)
		if err != nil {
			return fmt.Errorf("error encoding UTMP file record: %w", err)
		}
	}
	err := r.bucket.Store(bucketKeyFileRecords, buf.Bytes())
	if err != nil {
		return fmt.Errorf("error writing UTMP fiel records to disk: %w", err)
	}

	r.log.Debugf("Wrote %d UTMP file records to disk", len(r.savedUtmpFiles))
	return nil
}

func (r *UtmpFileReader) saveLoginSessionsToDisk() error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	for _, loginRecord := range r.loginSessions {
		err := encoder.Encode(loginRecord)
		if err != nil {
			return fmt.Errorf("error encoding login record: %w", err)
		}
	}

	err := r.bucket.Store(bucketKeyLoginSessions, buf.Bytes())
	if err != nil {
		return fmt.Errorf("error writing login records to disk: %w", err)
	}

	r.log.Debugf("Wrote %d open login sessions to disk", len(r.loginSessions))
	return nil
}

func (r *UtmpFileReader) restoreStateFromDisk() error {
	err := r.restoreFileRecordsFromDisk()
	if err != nil {
		return err
	}

	err = r.restoreLoginSessionsFromDisk()
	if err != nil {
		return err
	}

	return nil
}

func (r *UtmpFileReader) restoreFileRecordsFromDisk() error {
	var decoder *gob.Decoder
	err := r.bucket.Load(bucketKeyFileRecords, func(blob []byte) error {
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
			var utmpFile UtmpFile
			err = decoder.Decode(&utmpFile)
			if err == nil {
				r.savedUtmpFiles[utmpFile.Inode] = utmpFile
			} else if err == io.EOF {
				// Read all
				break
			} else {
				return fmt.Errorf("error decoding file record: %w", err)
			}
		}
	}
	r.log.Debugf("Restored %d UTMP file records from disk", len(r.savedUtmpFiles))

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
				r.loginSessions[loginRecord.TTY] = *loginRecord
			} else if err == io.EOF {
				// Read all
				break
			} else {
				return fmt.Errorf("error decodng login record: %w", err)
			}
		}
	}
	r.log.Debugf("Restored %d open login sessions from disk", len(r.loginSessions))

	return nil
}
