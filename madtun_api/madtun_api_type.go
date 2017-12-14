package madtun_api

type Type int16

const (
	_ Type = iota

	TypeData
	TypeDataResp

	TypeNewPath
	TypeNewPathResp
)
