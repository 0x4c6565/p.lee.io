package model

const PASTE_EXPIRES_NEVER = -1
const PASTE_EXPIRES_BURN = -2

type Paste struct {
	ID              string `db:"id"`
	Timestamp       int64  `db:"timestamp"`
	Expires         int64  `db:"expires"`
	Content         string `db:"content"`
	Syntax          string `db:"syntax"`
	ExpireTimestamp int64  `db:"expire_timestamp"`
}
