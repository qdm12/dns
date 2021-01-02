package models

import "os"

type OSOpenFileFunc func(name string, flag int, perm os.FileMode) (*os.File, error)
