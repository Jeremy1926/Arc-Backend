package database

import (
	ac "github.com/Arc-Services/Arc/shared/database/classes/anticheat"
	a "github.com/Arc-Services/Arc/shared/database/classes/auth"
	m "github.com/Arc-Services/Arc/shared/database/classes/main"
	mng "github.com/Arc-Services/Arc/shared/database/classes/management"
)

type Account = m.Account
type Session = a.Session
type Detection = ac.Detection
type Hardware = ac.Hardware
type Manager = mng.Manager
